package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-admin/app/admin/service/internal/data"
	"go-wind-admin/app/admin/service/internal/data/ent/product"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"

	"go-wind-admin/pkg/middleware/auth"
)

// ProductFeatureService 产品下特征条目服务 / Product feature admin service
//
// 设计依据 / Design ref:
//   - docs/thingmodel/sheji/模型管理/04-后端实现设计.md §4 + §6
//   - docs/thingmodel/sheji/修改记录/CR-001-结构化约束下沉到模型层.md
//
// CR-001 后核心业务规则：
//   - Create：仅 GLOBAL/LOCAL（DEFAULT 必须通过 PullFromDefault）；spec 必须由前端提供（或 GLOBAL 时由 CDF 拷贝）
//     GLOBAL/LOCAL 若 spec.unit 含 unitId 触发 unit ref_count +1
//   - PullFromDefault：直接 pf.spec = cdf.spec 深拷贝（不再做 snapshot/override 合并）
//     SKIP/REPLACE 冲突策略；source=DEFAULT 不重复 +1 ref_count
//   - Update：DRAFT 全字段可改；PUBLISHED 仅 name/description/sort_order/is_enabled 可改
//   - Delete：source=GLOBAL/LOCAL 时 -1 ref_count；DEFAULT 不动
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

// Get 详情（CR-001 后：单一 spec，无 effective_spec 合并）
func (s *ProductFeatureService) Get(ctx context.Context, req *thingmodelV1.GetProductFeatureRequest) (*thingmodelV1.ProductFeature, error) {
	return s.repo.Get(ctx, req)
}

// Create 创建（仅 GLOBAL/LOCAL；DEFAULT 必须走 PullFromDefault）
//
// CR-001 后流程：
//  1. 校验 product 存在且 status=DRAFT
//  2. 校验 source/ref_feature_id 配对
//  3. GLOBAL：从全局 feature 取骨架（code/identifier/feature_type）；LOCAL：使用前端传入骨架
//  4. spec 必填（GLOBAL/LOCAL 都由前端构造完整 spec）；做 V1–V17 校验
//  5. 同步冗余特化列
//  6. 创建行
//  7. 维护 unit.reference_count
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
	if source == thingmodelV1.ProductFeatureSource_DEFAULT {
		return nil, thingmodelV1.ErrorPfDefaultViaPullOnly("source=DEFAULT must use PullFromDefault RPC")
	}

	// 2. source/ref_feature_id 配对
	if msg := validateSourceRefMismatch(source, req.Data.GetRefFeatureId()); msg != "" {
		return nil, thingmodelV1.ErrorPfSourceRefMismatch("%s", msg)
	}

	// 3. GLOBAL：从 thing_features 读骨架（feature 不再含 spec / 特化列）
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
		// 用全局骨架覆盖前端可能传的骨架（保证 GLOBAL 路径以全局为权威）
		req.Data.FeatureType = feat.FeatureType
		req.Data.Code = feat.Code
		req.Data.Identifier = feat.Identifier
		if req.Data.Name == nil || req.Data.GetName() == "" {
			req.Data.Name = feat.Name
		}
		if req.Data.NameEn == nil || req.Data.GetNameEn() == "" {
			req.Data.NameEn = feat.NameEn
		}
		if req.Data.Description == nil || req.Data.GetDescription() == "" {
			req.Data.Description = feat.Description
		}
	}

	// 4. spec/type 一致性 + V1–V17 校验（spec 必填）
	if req.Data.Spec == nil {
		return nil, thingmodelV1.ErrorFeatureSpecInvalid("spec is required for product feature")
	}
	if msg := validateSpecTypeMismatch(req.Data.GetFeatureType(), req.Data.GetSpec()); msg != "" {
		return nil, thingmodelV1.ErrorPfSpecTypeMismatch("%s", msg)
	}
	if errs := validateFeatureSpecForType(req.Data.GetFeatureType(), req.Data.GetSpec()); len(errs) != 0 {
		return nil, thingmodelV1.ErrorFeatureSpecInvalid("spec invalid: %v", errs)
	}

	// 5. 同步冗余特化列
	syncSpecializedColumnsPF(req.Data)

	// 6. 创建
	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	// 7. 维护 unit.ref_count（GLOBAL/LOCAL 都要）
	if uid := extractPropertyUnitIDFromSpec(req.Data.GetSpec()); uid > 0 {
		_ = s.unitRepo.IncReferenceCount(ctx, uid, +1)
	}
	return &emptypb.Empty{}, nil
}

// PullFromDefault 批量从分类默认模型拉取
//
// CR-001 后流程：
//  1. 取产品（必须 DRAFT）
//  2. 取分类下的默认条目列表（可被 default_feature_ids 过滤；空=全部）
//  3. 查现有产品下与拉取目标重叠的 (product, ref_feature) 集合
//  4. 逐条处理：
//     - duplicate + SKIP → skipped
//     - duplicate + REPLACE → 删旧 + 重建
//     - 否则 → 直接 pf.spec = cdf.spec 深拷贝 → 创建
//  5. source=DEFAULT 不动 unit.reference_count（计数由 CDF 那一侧维护）
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
	for _, c := range cdfs {
		featureIDs = append(featureIDs, c.FeatureID)
		cdfDTOByFeatID[c.FeatureID] = s.catDefaultFeatRepo.ToDTO(c)
	}
	existing, err := s.repo.ListByProductFeatureIds(ctx, productRow.ID, featureIDs)
	if err != nil {
		return nil, err
	}
	existingByFeatID := make(map[uint32]uint32, len(existing))
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

		// 取全局 feature 骨架（用于 code/identifier/feature_type）
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

		// CR-001：直接深拷贝 cdf.spec
		cdfDTO := cdfDTOByFeatID[c.FeatureID]
		var pfSpec *thingmodelV1.FeatureSpec
		if cdfDTO.GetSpec() != nil {
			pfSpec = proto.Clone(cdfDTO.GetSpec()).(*thingmodelV1.FeatureSpec)
		}

		// 构造产品特征 DTO
		sourceDefault := thingmodelV1.ProductFeatureSource_DEFAULT
		featID := c.FeatureID
		pfData := &thingmodelV1.ProductFeature{
			ProductId:    trans.Ptr(productRow.ID),
			Source:       &sourceDefault,
			RefFeatureId: &featID,
			FeatureType:  feat.FeatureType,
			Code:         feat.Code,
			Identifier:   feat.Identifier,
			Name:         feat.Name,
			NameEn:       feat.NameEn,
			Description:  feat.Description,
			Spec:         pfSpec,
			// 冗余特化列从 cdf 拷贝（已经派生过）
			DataType:     cdfDTO.DataType,
			AccessMode:   cdfDTO.AccessMode,
			EventLevel:   cdfDTO.EventLevel,
			CallMode:     cdfDTO.CallMode,
			RelationType: cdfDTO.RelationType,
			SortOrder:    cdfDTO.SortOrder,
			IsEnabled:    trans.Ptr(true),
			TenantId:     &tenantID,
			CreatedBy:    trans.Ptr(operator.UserId),
		}
		createReq := &thingmodelV1.CreateProductFeatureRequest{Data: pfData}
		if err := s.repo.Create(ctx, createReq); err != nil {
			return nil, err
		}
		// source=DEFAULT 不动 unit.ref_count
		resp.Created = append(resp.Created, pfData)
	}
	return resp, nil
}

// CloneFromProduct 从另一产品克隆全部特征（含 LOCAL）— CR-001 后 spec 直接拷贝
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

	// 目标已存在的 code 集合
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
		// 转 ent → DTO 再回写（spec 走 mapper 已正确处理）
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
		// GLOBAL/LOCAL 来源要维护 unit ref_count；DEFAULT 不动
		if srcDTO.GetSource() == thingmodelV1.ProductFeatureSource_GLOBAL ||
			srcDTO.GetSource() == thingmodelV1.ProductFeatureSource_LOCAL {
			if uid := extractPropertyUnitIDFromSpec(srcDTO.GetSpec()); uid > 0 {
				_ = s.unitRepo.IncReferenceCount(ctx, uid, +1)
			}
		}
		resp.Created = append(resp.Created, srcDTO)
	}
	return resp, nil
}

// Update 更新（FieldMask）
//
// CR-001 后规则：
//   - DRAFT 全字段可改（含 spec）；若 spec 改动会重新校验 V1–V17 并同步冗余列
//   - PUBLISHED 仅 name/description/sort_order/is_enabled 可改（spec 完全冻结）
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

	// 取旧 DTO（用于 PUBLISHED 冻结判断 + spec.unit 变化）
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
		// CR-001：PUBLISHED 白名单收紧为 name/description/sort_order/is_enabled
		for _, p := range req.GetUpdateMask().GetPaths() {
			switch p {
			case "description",
				"name", "name_en", "nameEn",
				"sort_order", "sortOrder", "is_enabled", "isEnabled",
				"updated_by", "updatedBy":
				// allowed
			default:
				return nil, thingmodelV1.ErrorPfProductPublished("published product: field %s is frozen", p)
			}
		}
	}

	// 校验新 spec（若提供）
	if req.Data.Spec != nil {
		ft := req.Data.GetFeatureType()
		if ft == thingmodelV1.FeatureType_FEATURE_TYPE_UNSPECIFIED {
			ft = oldDTO.GetFeatureType()
		}
		if msg := validateSpecTypeMismatch(ft, req.Data.GetSpec()); msg != "" {
			return nil, thingmodelV1.ErrorPfSpecTypeMismatch("%s", msg)
		}
		if errs := validateFeatureSpecForType(ft, req.Data.GetSpec()); len(errs) != 0 {
			return nil, thingmodelV1.ErrorFeatureSpecInvalid("spec invalid: %v", errs)
		}
		syncSpecializedColumnsPF(req.Data)
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	// 维护 unit.ref_count（spec.unit 变化时）
	if req.Data.Spec != nil {
		oldUID := extractPropertyUnitIDFromSpec(oldDTO.GetSpec())
		newUID := extractPropertyUnitIDFromSpec(req.Data.GetSpec())
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

// Delete 批量删除（source=GLOBAL/LOCAL 时 -1 unit.ref_count；DEFAULT 不动）
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

		// 维护 unit.ref_count（仅 GLOBAL/LOCAL；DEFAULT 不动）
		if dto.GetSource() == thingmodelV1.ProductFeatureSource_GLOBAL ||
			dto.GetSource() == thingmodelV1.ProductFeatureSource_LOCAL {
			if uid := extractPropertyUnitIDFromSpec(dto.GetSpec()); uid > 0 {
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
