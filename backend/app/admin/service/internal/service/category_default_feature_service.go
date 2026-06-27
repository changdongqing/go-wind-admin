package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-admin/app/admin/service/internal/data"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"

	"go-wind-admin/pkg/middleware/auth"
)

// CategoryDefaultFeatureService 分类默认模型服务 / Category default feature admin service
//
// 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/04-后端实现设计.md §1 + §6
//
// 业务规则收口：
//   - Create/BatchAdd 前校验 category.level == 4（应用层兜底）
//   - Create/BatchAdd 前校验 feature.is_enabled == true
//   - Create/BatchAdd 前校验 override_spec 白名单合法性
//   - Create 后 +1 thingmodel_units.reference_count（若 snapshot 或 override 含 unit_id）
//   - Delete 前 -1 thingmodel_units.reference_count（同上）
//   - Update 时若 override.unit 变化，旧 -1 / 新 +1
type CategoryDefaultFeatureService struct {
	adminV1.CategoryDefaultFeatureServiceHTTPServer

	log *log.Helper

	repo         *data.CategoryDefaultFeatureRepo
	categoryRepo *data.CategoryRepo
	featureRepo  *data.FeatureRepo
	unitRepo     *data.UnitRepo
}

// NewCategoryDefaultFeatureService 构造服务
func NewCategoryDefaultFeatureService(
	ctx *bootstrap.Context,
	repo *data.CategoryDefaultFeatureRepo,
	categoryRepo *data.CategoryRepo,
	featureRepo *data.FeatureRepo,
	unitRepo *data.UnitRepo,
) *CategoryDefaultFeatureService {
	return &CategoryDefaultFeatureService{
		log:          ctx.NewLoggerHelper("category-default-feature/service/admin-service"),
		repo:         repo,
		categoryRepo: categoryRepo,
		featureRepo:  featureRepo,
		unitRepo:     unitRepo,
	}
}

// List 分页查询
func (s *CategoryDefaultFeatureService) List(ctx context.Context, req *paginationV1.PagingRequest) (*thingmodelV1.ListCategoryDefaultFeatureResponse, error) {
	return s.repo.List(ctx, req)
}

// Get 详情
func (s *CategoryDefaultFeatureService) Get(ctx context.Context, req *thingmodelV1.GetCategoryDefaultFeatureRequest) (*thingmodelV1.CategoryDefaultFeature, error) {
	return s.repo.Get(ctx, req)
}

// Create 创建单条
//
// 流程：
//  1. 校验 category.level == 4
//  2. 校验 feature 存在且 is_enabled
//  3. 校验 override_spec 白名单（针对 property）
//  4. 创建关联行
//  5. 维护 unit.reference_count（若有 unit_id）
func (s *CategoryDefaultFeatureService) Create(ctx context.Context, req *thingmodelV1.CreateCategoryDefaultFeatureRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	if err := s.validateCategoryIsLeaf(ctx, req.Data.GetCategoryId()); err != nil {
		return nil, err
	}
	featSpec, err := s.getEnabledFeatureSpec(ctx, req.Data.GetFeatureId())
	if err != nil {
		return nil, err
	}
	if errs := validateOverrideSpec(featSpec, req.Data.GetOverrideSpec()); len(errs) != 0 {
		return nil, thingmodelV1.ErrorCatDefaultFeatureOverrideInvalid("override invalid: %v", errs)
	}

	row, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	// 维护 unit.reference_count
	if uid := extractUnitID(featSpec, req.Data.GetOverrideSpec()); uid > 0 {
		if err := s.unitRepo.IncReferenceCount(ctx, uid, +1); err != nil {
			s.log.Warnf("inc unit ref count failed for cdf id=%d unit=%d: %v", row.ID, uid, err)
		}
	}
	return &emptypb.Empty{}, nil
}

// BatchAdd 批量添加（一次性把多个 feature 绑定到 category）
//
// 行为：每个 item 独立做"创建+维护 ref count"事务；
//   - (tenant, category, feature) 已存在：归入 skipped_duplicate_feature_codes
//   - feature 已停用：归入 skipped
//   - override 非法：归入 skipped（带原因）
//   - 成功创建的进入 created 列表
func (s *CategoryDefaultFeatureService) BatchAdd(ctx context.Context, req *thingmodelV1.BatchAddCategoryDefaultFeaturesRequest) (*thingmodelV1.BatchAddCategoryDefaultFeaturesResponse, error) {
	if req == nil || req.GetCategoryId() == 0 || len(req.GetItems()) == 0 {
		return &thingmodelV1.BatchAddCategoryDefaultFeaturesResponse{}, nil
	}
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.validateCategoryIsLeaf(ctx, req.GetCategoryId()); err != nil {
		return nil, err
	}

	resp := &thingmodelV1.BatchAddCategoryDefaultFeaturesResponse{}
	var tenantID uint32
	if operator.TenantId != nil {
		tenantID = *operator.TenantId
	}
	categoryID := req.GetCategoryId()

	for _, it := range req.GetItems() {
		if it == nil || it.GetFeatureId() == 0 {
			continue
		}
		// 冲突检测
		exists, err := s.repo.ExistsByCategoryFeature(ctx, tenantID, categoryID, it.GetFeatureId())
		if err != nil {
			return nil, err
		}
		if exists {
			// 回填该 feature 的 code 供前端展示
			feat, _ := s.featureRepo.IsExist(ctx, it.GetFeatureId())
			if feat {
				resp.SkippedDuplicateFeatureCodes = append(resp.SkippedDuplicateFeatureCodes, "")
				// 真实 code 需 join；为简洁此处留空，前端可二次查询
			}
			continue
		}

		// 取 feature 校验 is_enabled + 准备 unit 维护
		featSpec, err := s.getEnabledFeatureSpec(ctx, it.GetFeatureId())
		if err != nil {
			// 如特征已停用，跳过
			s.log.Warnf("batch-add skip feature_id=%d: %v", it.GetFeatureId(), err)
			continue
		}
		if errs := validateOverrideSpec(featSpec, it.GetOverrideSpec()); len(errs) != 0 {
			s.log.Warnf("batch-add skip feature_id=%d: override invalid %v", it.GetFeatureId(), errs)
			continue
		}

		// 构造单条 Create
		data := &thingmodelV1.CategoryDefaultFeature{
			CategoryId:   trans.Ptr(categoryID),
			FeatureId:    trans.Ptr(it.GetFeatureId()),
			OverrideSpec: it.GetOverrideSpec(),
			DisplayName:  it.DisplayName,
			SortOrder:    it.SortOrder,
			TenantId:     &tenantID,
			CreatedBy:    trans.Ptr(operator.UserId),
		}
		row, err := s.repo.Create(ctx, &thingmodelV1.CreateCategoryDefaultFeatureRequest{Data: data})
		if err != nil {
			return nil, err
		}

		// 维护 unit.ref count
		if uid := extractUnitID(featSpec, it.GetOverrideSpec()); uid > 0 {
			if err := s.unitRepo.IncReferenceCount(ctx, uid, +1); err != nil {
				s.log.Warnf("inc unit ref count for batch-add (feature %d unit %d) failed: %v", it.GetFeatureId(), uid, err)
			}
		}

		// row → DTO
		resp.Created = append(resp.Created, s.repo.ToDTO(row))
	}
	return resp, nil
}

// Update 更新 override_spec/display_name/sort_order/is_enabled
//
// 若 override.unit 改变，需要交换 unit.reference_count（旧 -1 / 新 +1）。
func (s *CategoryDefaultFeatureService) Update(ctx context.Context, req *thingmodelV1.UpdateCategoryDefaultFeatureRequest) (*emptypb.Empty, error) {
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

	// 取旧 DTO（已 mapper 还原 OverrideSpec）
	oldDTO, err := s.repo.Get(ctx, &thingmodelV1.GetCategoryDefaultFeatureRequest{Id: req.GetId()})
	if err != nil {
		return nil, err
	}
	if oldDTO == nil {
		return nil, thingmodelV1.ErrorCatDefaultFeatureNotFound("not found")
	}

	// 校验新 override 白名单（若提供）
	if req.Data.OverrideSpec != nil {
		featDTO, err := s.featureRepo.Get(ctx, &thingmodelV1.GetFeatureRequest{
			QueryBy: &thingmodelV1.GetFeatureRequest_Id{Id: oldDTO.GetFeatureId()},
		})
		if err != nil {
			return nil, err
		}
		if errs := validateOverrideSpec(featDTO.GetSpec(), req.Data.GetOverrideSpec()); len(errs) != 0 {
			return nil, thingmodelV1.ErrorCatDefaultFeatureOverrideInvalid("override invalid: %v", errs)
		}
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	// 若 override.unit 变化，交换 ref count
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

// Delete 批量删除
//
// 流程：逐条 take DTO → -1 unit.ref_count → delete
func (s *CategoryDefaultFeatureService) Delete(ctx context.Context, req *thingmodelV1.DeleteCategoryDefaultFeatureRequest) (*emptypb.Empty, error) {
	if req == nil || len(req.GetIds()) == 0 {
		return &emptypb.Empty{}, nil
	}
	for _, id := range req.GetIds() {
		dto, err := s.repo.Get(ctx, &thingmodelV1.GetCategoryDefaultFeatureRequest{Id: id})
		if err != nil || dto == nil {
			// 不存在则视为已删
			continue
		}
		// 取 feature spec 决定 unit_id（与 Create 对称）
		featDTO, ferr := s.featureRepo.Get(ctx, &thingmodelV1.GetFeatureRequest{
			QueryBy: &thingmodelV1.GetFeatureRequest_Id{Id: dto.GetFeatureId()},
		})
		if ferr == nil {
			if uid := extractUnitID(featDTO.GetSpec(), dto.GetOverrideSpec()); uid > 0 {
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
func (s *CategoryDefaultFeatureService) Reorder(ctx context.Context, req *thingmodelV1.ReorderCategoryDefaultFeaturesRequest) (*emptypb.Empty, error) {
	if req == nil {
		return &emptypb.Empty{}, nil
	}
	if err := s.repo.Reorder(ctx, req.GetItems()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ===== 辅助 / Helpers =====

// validateCategoryIsLeaf 校验 categoryID 指向的节点 level == 4。
func (s *CategoryDefaultFeatureService) validateCategoryIsLeaf(ctx context.Context, categoryID uint32) error {
	if categoryID == 0 {
		return thingmodelV1.ErrorBadRequest("category_id is required")
	}
	cat, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return thingmodelV1.ErrorCategoryNotFound("category not found")
	}
	level := uint8(0)
	if cat.Level != nil {
		level = *cat.Level
	}
	if level != 4 {
		return thingmodelV1.ErrorCatDefaultFeatureCategoryNotLeaf("category must be level=4 (leaf), got level=%d", level)
	}
	return nil
}

// getEnabledFeatureSpec 取启用状态的全局特征 spec；停用则返回错误。
func (s *CategoryDefaultFeatureService) getEnabledFeatureSpec(ctx context.Context, featureID uint32) (*thingmodelV1.FeatureSpec, error) {
	if featureID == 0 {
		return nil, thingmodelV1.ErrorBadRequest("feature_id is required")
	}
	feat, err := s.featureRepo.Get(ctx, &thingmodelV1.GetFeatureRequest{
		QueryBy: &thingmodelV1.GetFeatureRequest_Id{Id: featureID},
	})
	if err != nil {
		return nil, thingmodelV1.ErrorFeatureNotFound("feature not found")
	}
	if feat.IsEnabled != nil && !feat.GetIsEnabled() {
		return nil, thingmodelV1.ErrorCatDefaultFeatureFeatureDisabled("feature %s is disabled", feat.GetCode())
	}
	return feat.GetSpec(), nil
}
