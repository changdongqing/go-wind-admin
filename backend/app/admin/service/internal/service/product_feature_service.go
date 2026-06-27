package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-admin/app/admin/service/internal/data"
	"go-wind-admin/app/admin/service/internal/data/ent/product"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"

	"go-wind-admin/pkg/middleware/auth"
)

// ProductFeatureService 产品下特征条目服务 / Product feature admin service
//
// 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/04-后端实现设计.md §4 + §6
//
// 核心业务规则：
//   - Create：仅 GLOBAL/LOCAL（DEFAULT 必须通过 PullFromDefault）
//     GLOBAL 触发 thing_features 引用计数 +1（本期 Feature 无 ref_count 列，只维护 unit）
//     GLOBAL/LOCAL 若 spec.unit 含 unitId 触发 unit ref_count +1
//   - PullFromDefault：批量从分类默认模型拷贝到产品（值复制 + override 合并）
//     SKIP/REPLACE 冲突策略；source=DEFAULT 不重复 +1 ref_count
//   - Update：仅白名单字段；产品 PUBLISHED 后冻结结构
//   - Delete：source=GLOBAL 时 -1 ref_count；DEFAULT/LOCAL 不动
type ProductFeatureService struct {
	adminV1.ProductFeatureServiceHTTPServer

	log *log.Helper

	repo               *data.ProductFeatureRepo
	productRepo        *data.ProductRepo
	featureRepo        *data.FeatureRepo
	unitRepo           *data.UnitRepo
	catDefaultFeatRepo *data.CategoryDefaultFeatureRepo
}

// NewProductFeatureService 构造服务
func NewProductFeatureService(
	ctx *bootstrap.Context,
	repo *data.ProductFeatureRepo,
	productRepo *data.ProductRepo,
	featureRepo *data.FeatureRepo,
	unitRepo *data.UnitRepo,
	catDefaultFeatRepo *data.CategoryDefaultFeatureRepo,
) *ProductFeatureService {
	return &ProductFeatureService{
		log:                ctx.NewLoggerHelper("product-feature/service/admin-service"),
		repo:               repo,
		productRepo:        productRepo,
		featureRepo:        featureRepo,
		unitRepo:           unitRepo,
		catDefaultFeatRepo: catDefaultFeatRepo,
	}
}

// List 分页查询
func (s *ProductFeatureService) List(ctx context.Context, req *paginationV1.PagingRequest) (*thingmodelV1.ListProductFeatureResponse, error) {
	return s.repo.List(ctx, req)
}

// Get 详情（返回 feature_snapshot + override_spec + effective_spec）
func (s *ProductFeatureService) Get(ctx context.Context, req *thingmodelV1.GetProductFeatureRequest) (*thingmodelV1.ProductFeature, error) {
	dto, err := s.repo.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	if dto != nil {
		// 后端合并 effective_spec
		dto.EffectiveSpec = effectiveSpec(dto.GetFeatureSnapshot(), dto.GetOverrideSpec())
	}
	return dto, nil
}

// Create 创建（仅 GLOBAL/LOCAL；DEFAULT 必须走 PullFromDefault）
//
// 流程：
//  1. 校验 product 存在且 status=DRAFT
//  2. 校验 source/ref_feature_id 配对
//  3. 校验 spec/feature_type 一致性
//  4. GLOBAL：取全局特征 spec 做 snapshot；LOCAL：使用传入 snapshot
//  5. 校验 override 白名单
//  6. 创建行
//  7. GLOBAL/LOCAL：维护 unit.reference_count
func (s *ProductFeatureService) Create(ctx context.Context, req *thingmodelV1.CreateProductFeatureRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	// 1. 校验 product DRAFT
	if err := s.requireDraftProduct(ctx, req.Data.GetProductId()); err != nil {
		return nil, err
	}

	source := req.Data.GetSource()
	// 禁止 source=DEFAULT 走 Create（必须走 PullFromDefault）
	if source == thingmodelV1.ProductFeatureSource_DEFAULT {
		return nil, thingmodelV1.ErrorPfDefaultViaPullOnly("source=DEFAULT must use PullFromDefault RPC")
	}

	// 2. source/ref_feature_id 配对
	if msg := validateSourceRefMismatch(source, req.Data.GetRefFeatureId()); msg != "" {
		return nil, thingmodelV1.ErrorPfSourceRefMismatch(msg)
	}

	// 3. GLOBAL 时从 thing_features 读 snapshot；LOCAL 时使用传入
	if source == thingmodelV1.ProductFeatureSource_GLOBAL {
		feat, err := s.featureRepo.Get(ctx, &thingmodelV1.GetFeatureRequest{
			QueryBy: &thingmodelV1.GetFeatureRequest_Id{Id: req.Data.GetRefFeatureId()},
		})
		if err != nil {
			return nil, thingmodelV1.ErrorFeatureNotFound("ref feature not found")
		}
		if feat.IsEnabled != nil && !feat.GetIsEnabled() {
			return nil, thingmodelV1.ErrorPfFeatureDisabled("referenced feature is disabled")
		}
		// 用全局 spec 覆盖前端可能传的 snapshot（保证 GLOBAL 路径以全局为权威）
		req.Data.FeatureSnapshot = feat.GetSpec()
		req.Data.FeatureType = feat.FeatureType
		req.Data.Code = feat.Code
		req.Data.Identifier = feat.Identifier
		if req.Data.Name == nil || req.Data.GetName() == "" {
			req.Data.Name = feat.Name
		}
		// 冗余特化列同步
		req.Data.DataType = feat.DataType
		req.Data.AccessMode = feat.AccessMode
		req.Data.EventLevel = feat.EventLevel
		req.Data.CallMode = feat.CallMode
		req.Data.RelationType = feat.RelationType
	}

	// 4. spec/type 一致性
	if msg := validateSpecTypeMismatch(req.Data.GetFeatureType(), req.Data.GetFeatureSnapshot()); msg != "" {
		return nil, thingmodelV1.ErrorPfSpecTypeMismatch(msg)
	}

	// 5. override 白名单
	if errs := validateOverrideSpec(req.Data.GetFeatureSnapshot(), req.Data.GetOverrideSpec()); len(errs) != 0 {
		return nil, thingmodelV1.ErrorPfOverrideInvalid("override invalid: %v", errs)
	}

	// 6. 创建
	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	// 7. 维护 unit.ref_count（GLOBAL/LOCAL 都要）
	if uid := extractUnitID(req.Data.GetFeatureSnapshot(), req.Data.GetOverrideSpec()); uid > 0 {
		_ = s.unitRepo.IncReferenceCount(ctx, uid, +1)
	}
	return &emptypb.Empty{}, nil
}

// PullFromDefault 批量从分类默认模型拉取
//
// 流程：
//  1. 取产品（必须 DRAFT）
//  2. 取分类下的默认条目列表（可被 default_feature_ids 过滤；空=全部）
//  3. 查现有产品下与拉取目标重叠的 (product, ref_feature) 集合
//  4. 逐条处理：
//     - duplicate + SKIP → skipped
//     - duplicate + REPLACE → 删旧 + 重建
//     - 否则 → 取全局 feature spec → merge default 层 override 得到 snapshot → 创建
//  5. source=DEFAULT 不动 thing_features/unit.reference_count（计数由 default_features 那一侧维护）
func (s *ProductFeatureService) PullFromDefault(ctx context.Context, req *thingmodelV1.PullFromDefaultRequest) (*thingmodelV1.PullFromDefaultResponse, error) {
	if req == nil || req.GetProductId() == 0 {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 1. 取产品（必须 DRAFT）
	productRow, err := s.productRepo.GetEntityByID(ctx, req.GetProductId())
	if err != nil {
		return nil, err
	}
	if productRow.Status != product.StatusDraft {
		return nil, thingmodelV1.ErrorPfProductPublished("product is not DRAFT; cannot pull")
	}

	// 2. 取该分类下的默认条目
	tenantID := uint32(0)
	if operator.TenantId != nil {
		tenantID = *operator.TenantId
	}
	cdfs, err := s.catDefaultFeatRepo.ListByCategory(ctx, productRow.CategoryID, req.GetDefaultFeatureIds(), tenantID)
	if err != nil {
		return nil, err
	}

	// 3. 查现有 product_features 中已存在的 (product, ref_feature)
	featureIDs := make([]uint32, 0, len(cdfs))
	cdfDTOByFeatID := make(map[uint32]*thingmodelV1.CategoryDefaultFeature, len(cdfs))
	cdfIDByFeatID := make(map[uint32]uint32, len(cdfs))
	for _, c := range cdfs {
		featureIDs = append(featureIDs, c.FeatureID)
		cdfDTOByFeatID[c.FeatureID] = s.catDefaultFeatRepo.ToDTO(c)
		cdfIDByFeatID[c.FeatureID] = c.ID
	}
	existing, err := s.repo.ListByProductFeatureIds(ctx, productRow.ID, featureIDs)
	if err != nil {
		return nil, err
	}
	existingByFeatID := make(map[uint32]uint32, len(existing)) // ref_feature_id → product_feature.id
	for _, pf := range existing {
		if pf.RefFeatureID != nil {
			existingByFeatID[*pf.RefFeatureID] = pf.ID
		}
	}

	// 4. 逐条处理
	resp := &thingmodelV1.PullFromDefaultResponse{}
	onConflict := req.GetOnConflict()
	if onConflict == thingmodelV1.ConflictPolicy_CONFLICT_POLICY_UNSPECIFIED {
		onConflict = thingmodelV1.ConflictPolicy_SKIP
	}

	for _, c := range cdfs {
		// 冲突
		if existingPFID, dup := existingByFeatID[c.FeatureID]; dup {
			switch onConflict {
			case thingmodelV1.ConflictPolicy_SKIP:
				resp.Skipped = append(resp.Skipped, &thingmodelV1.PullFromDefaultResponse_PullSkipped{
					DefaultFeatureId: c.ID,
					Reason:           "duplicate",
				})
				continue
			case thingmodelV1.ConflictPolicy_REPLACE:
				if err := s.repo.Delete(ctx, existingPFID); err != nil {
					return nil, err
				}
			}
		}

		// 取全局 feature
		feat, err := s.featureRepo.Get(ctx, &thingmodelV1.GetFeatureRequest{
			QueryBy: &thingmodelV1.GetFeatureRequest_Id{Id: c.FeatureID},
		})
		if err != nil {
			resp.Skipped = append(resp.Skipped, &thingmodelV1.PullFromDefaultResponse_PullSkipped{
				DefaultFeatureId: c.ID,
				Reason:           "feature_not_found",
			})
			continue
		}
		if feat.IsEnabled != nil && !feat.GetIsEnabled() {
			resp.Skipped = append(resp.Skipped, &thingmodelV1.PullFromDefaultResponse_PullSkipped{
				DefaultFeatureId: c.ID,
				Reason:           "feature_disabled",
			})
			continue
		}

		// 合并 default 层 override 到 snapshot
		cdfDTO := cdfDTOByFeatID[c.FeatureID]
		snapshot := effectiveSpec(feat.GetSpec(), cdfDTO.GetOverrideSpec())

		// 构造产品特征 DTO
		sourceDefault := thingmodelV1.ProductFeatureSource_DEFAULT
		featID := c.FeatureID
		pfData := &thingmodelV1.ProductFeature{
			ProductId:       trans.Ptr(productRow.ID),
			Source:          &sourceDefault,
			RefFeatureId:    &featID,
			FeatureType:     feat.FeatureType,
			Code:            feat.Code,
			Identifier:      feat.Identifier,
			Name:            feat.Name,
			NameEn:          feat.NameEn,
			Description:     feat.Description,
			FeatureSnapshot: snapshot,
			OverrideSpec:    nil, // 产品层暂未覆写
			DataType:        feat.DataType,
			AccessMode:      feat.AccessMode,
			EventLevel:      feat.EventLevel,
			CallMode:        feat.CallMode,
			RelationType:    feat.RelationType,
			SortOrder:       cdfDTO.SortOrder,
			IsEnabled:       trans.Ptr(true),
			TenantId:        &tenantID,
			CreatedBy:       trans.Ptr(operator.UserId),
		}
		createReq := &thingmodelV1.CreateProductFeatureRequest{Data: pfData}
		if err := s.repo.Create(ctx, createReq); err != nil {
			return nil, err
		}
		// source=DEFAULT 不动 thing_features/unit ref_count
		// 返回时由 list 重新读取——为简洁此处省略，直接把 pfData 当响应（不带 id）
		resp.Created = append(resp.Created, pfData)
	}
	return resp, nil
}

// CloneFromProduct 从另一产品克隆全部特征（含 LOCAL）— 本期最小实现
func (s *ProductFeatureService) CloneFromProduct(ctx context.Context, req *thingmodelV1.CloneFromProductRequest) (*thingmodelV1.CloneFromProductResponse, error) {
	if req == nil || req.GetProductId() == 0 || req.GetSourceProductId() == 0 {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetProductId() == req.GetSourceProductId() {
		return nil, thingmodelV1.ErrorBadRequest("source and target product must differ")
	}
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 目标产品必须 DRAFT
	if err := s.requireDraftProduct(ctx, req.GetProductId()); err != nil {
		return nil, err
	}

	// 取源产品特征
	srcRows, err := s.repo.ListByProduct(ctx, req.GetSourceProductId())
	if err != nil {
		return nil, err
	}

	resp := &thingmodelV1.CloneFromProductResponse{}
	onConflict := req.GetOnConflict()
	if onConflict == thingmodelV1.ConflictPolicy_CONFLICT_POLICY_UNSPECIFIED {
		onConflict = thingmodelV1.ConflictPolicy_SKIP
	}

	// 目标已存在的 code 集合（用于冲突判断）
	dstRows, err := s.repo.ListByProduct(ctx, req.GetProductId())
	if err != nil {
		return nil, err
	}
	dstCodes := make(map[string]uint32, len(dstRows))
	for _, d := range dstRows {
		if d.Code != nil {
			dstCodes[*d.Code] = d.ID
		}
	}

	tenantID := uint32(0)
	if operator.TenantId != nil {
		tenantID = *operator.TenantId
	}

	for _, src := range srcRows {
		code := ""
		if src.Code != nil {
			code = *src.Code
		}
		if dstID, dup := dstCodes[code]; dup {
			if onConflict == thingmodelV1.ConflictPolicy_REPLACE {
				if err := s.repo.Delete(ctx, dstID); err != nil {
					return nil, err
				}
			} else {
				resp.SkippedCount++
				continue
			}
		}
		// 转 ent → DTO 再回写（snapshot/override 走 mapper 已正确处理）
		srcDTO := s.repo.ToDTO(src)
		srcDTO.Id = nil
		srcDTO.ProductId = trans.Ptr(req.GetProductId())
		srcDTO.TenantId = &tenantID
		srcDTO.CreatedBy = trans.Ptr(operator.UserId)
		srcDTO.UpdatedBy = nil

		createReq := &thingmodelV1.CreateProductFeatureRequest{Data: srcDTO}
		if err := s.repo.Create(ctx, createReq); err != nil {
			return nil, err
		}
		// GLOBAL 来源要维护 unit ref_count；DEFAULT/LOCAL 不动
		if srcDTO.GetSource() == thingmodelV1.ProductFeatureSource_GLOBAL {
			if uid := extractUnitID(srcDTO.GetFeatureSnapshot(), srcDTO.GetOverrideSpec()); uid > 0 {
				_ = s.unitRepo.IncReferenceCount(ctx, uid, +1)
			}
		}
		resp.Created = append(resp.Created, srcDTO)
	}
	return resp, nil
}

// Update 更新（FieldMask）；PUBLISHED 时禁止改结构（仅 override/sort/enabled/description 等）
func (s *ProductFeatureService) Update(ctx context.Context, req *thingmodelV1.UpdateProductFeatureRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Data.Id = trans.Ptr(req.GetId())
	req.Data.UpdatedBy = trans.Ptr(operator.UserId)
	if req.UpdateMask != nil {
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "updated_by")
	}

	// 取旧 DTO（用于 PUBLISHED 冻结判断 + override.unit 变化）
	oldDTO, err := s.repo.Get(ctx, &thingmodelV1.GetProductFeatureRequest{Id: req.GetId()})
	if err != nil {
		return nil, err
	}
	if oldDTO == nil {
		return nil, thingmodelV1.ErrorProductFeatureNotFound("not found")
	}

	// 取产品状态
	prodRow, err := s.productRepo.GetEntityByID(ctx, oldDTO.GetProductId())
	if err != nil {
		return nil, err
	}
	if prodRow.Status == product.StatusPublished {
		// 仅允许 override_spec / description / display_name / sort_order / is_enabled 等
		for _, p := range req.GetUpdateMask().GetPaths() {
			switch p {
			case "override_spec", "overrideSpec",
				"description", "name", "name_en", "nameEn",
				"sort_order", "sortOrder", "is_enabled", "isEnabled",
				"updated_by", "updatedBy":
				// allowed
			default:
				return nil, thingmodelV1.ErrorPfProductPublished("published product: field %s is frozen", p)
			}
		}
	}

	// 校验新 override 白名单
	if req.Data.OverrideSpec != nil {
		if errs := validateOverrideSpec(oldDTO.GetFeatureSnapshot(), req.Data.GetOverrideSpec()); len(errs) != 0 {
			return nil, thingmodelV1.ErrorPfOverrideInvalid("override invalid: %v", errs)
		}
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	// 维护 unit.ref_count（override.unit 变化时）
	if req.Data.OverrideSpec != nil {
		oldUID := uint32(0)
		if oldDTO.GetOverrideSpec() != nil && oldDTO.GetOverrideSpec().GetUnit() != nil {
			oldUID = oldDTO.GetOverrideSpec().GetUnit().GetUnitId()
		}
		newUID := uint32(0)
		if req.Data.GetOverrideSpec().GetUnit() != nil {
			newUID = req.Data.GetOverrideSpec().GetUnit().GetUnitId()
		}
		if oldUID != newUID {
			if oldUID > 0 {
				_ = s.unitRepo.IncReferenceCount(ctx, oldUID, -1)
			}
			if newUID > 0 {
				_ = s.unitRepo.IncReferenceCount(ctx, newUID, +1)
			}
		}
	}
	return &emptypb.Empty{}, nil
}

// Delete 批量删除（source=GLOBAL 时 -1 unit.ref_count；DEFAULT/LOCAL 不动）
func (s *ProductFeatureService) Delete(ctx context.Context, req *thingmodelV1.DeleteProductFeatureRequest) (*emptypb.Empty, error) {
	if req == nil || len(req.GetIds()) == 0 {
		return &emptypb.Empty{}, nil
	}
	for _, id := range req.GetIds() {
		dto, err := s.repo.Get(ctx, &thingmodelV1.GetProductFeatureRequest{Id: id})
		if err != nil || dto == nil {
			continue
		}

		// PUBLISHED 产品禁止删特征
		prodRow, err := s.productRepo.GetEntityByID(ctx, dto.GetProductId())
		if err != nil {
			return nil, err
		}
		if prodRow.Status == product.StatusPublished {
			return nil, thingmodelV1.ErrorPfProductPublished("published product: cannot delete features")
		}

		// 维护 unit.ref_count（仅 GLOBAL 或 LOCAL 含 unit_id）
		if dto.GetSource() == thingmodelV1.ProductFeatureSource_GLOBAL ||
			dto.GetSource() == thingmodelV1.ProductFeatureSource_LOCAL {
			if uid := extractUnitID(dto.GetFeatureSnapshot(), dto.GetOverrideSpec()); uid > 0 {
				_ = s.unitRepo.IncReferenceCount(ctx, uid, -1)
			}
		}

		if err := s.repo.Delete(ctx, id); err != nil {
			return nil, err
		}
	}
	return &emptypb.Empty{}, nil
}

// Reorder 拖拽排序
func (s *ProductFeatureService) Reorder(ctx context.Context, req *thingmodelV1.ReorderProductFeaturesRequest) (*emptypb.Empty, error) {
	if req == nil {
		return &emptypb.Empty{}, nil
	}
	if err := s.repo.Reorder(ctx, req.GetItems()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ===== 辅助 / Helpers =====

// requireDraftProduct 校验产品存在且 status=DRAFT
func (s *ProductFeatureService) requireDraftProduct(ctx context.Context, productID uint32) error {
	if productID == 0 {
		return thingmodelV1.ErrorBadRequest("product_id is required")
	}
	row, err := s.productRepo.GetEntityByID(ctx, productID)
	if err != nil {
		return err
	}
	if row.Status != product.StatusDraft {
		return thingmodelV1.ErrorPfProductPublished("product is not DRAFT")
	}
	return nil
}
