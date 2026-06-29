package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-admin/app/admin/service/internal/data"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"

	"go-wind-admin/pkg/middleware/auth"
)

// CategoryDefaultFeatureService 分类默认模型服务 / Category default feature admin service
//
// 设计依据 / Design ref:
//   - docs/thingmodel/sheji/模型管理/04-后端实现设计.md §1 + §6
//   - docs/thingmodel/sheji/修改记录/CR-001-结构化约束下沉到模型层.md
//
// CR-001 后业务规则：
//   - Create/BatchAdd 前校验 category.level == 4（应用层兜底）
//   - Create/BatchAdd 前校验 feature.is_enabled == true
//   - Create/BatchAdd/Update 前调用 spec_validator（V1–V17）校验完整 spec
//   - 写入前调用 syncSpecializedColumnsCDF 同步 5 个冗余抽取列
//   - Create 后 +1 thingmodel_units.reference_count（若 spec.property.unit.unit_id 有值）
//   - Update 时若 unit 变化，旧 -1 / 新 +1
//   - Delete 前 -1 thingmodel_units.reference_count
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

// Create 创建单条（CR-001 后承载完整 spec）
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
	featureType, err := s.getEnabledFeatureType(ctx, req.Data.GetFeatureId())
	if err != nil {
		return nil, err
	}
	// CR-001：完整 spec 校验（允许 spec 为 nil，由用户后续逐项编辑）
	if req.Data.Spec != nil {
		if mismatch := validateSpecTypeMismatch(featureType, req.Data.GetSpec()); mismatch != "" {
			return nil, thingmodelV1.ErrorFeatureSpecInvalid("%s", mismatch)
		}
		if errs := validateFeatureSpecForType(featureType, req.Data.GetSpec()); len(errs) != 0 {
			return nil, thingmodelV1.ErrorFeatureSpecInvalid("spec invalid: %v", errs)
		}
	}
	// 同步冗余抽取列
	syncSpecializedColumnsCDF(req.Data)

	if _, err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	// 维护 unit.reference_count
	if uid := extractPropertyUnitIDFromSpec(req.Data.GetSpec()); uid > 0 {
		if err := s.unitRepo.IncReferenceCount(ctx, uid, +1); err != nil {
			s.log.Warnf("inc unit ref count failed for cdf unit=%d: %v", uid, err)
		}
	}
	return &emptypb.Empty{}, nil
}

// BatchAdd 批量添加（一次性把多个 feature 绑定到 category）
//
// CR-001：item.Spec 允许为 nil（用户后续单独编辑），但若给了 spec 则做完整校验。
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
			resp.SkippedDuplicateFeatureCodes = append(resp.SkippedDuplicateFeatureCodes, "")
			continue
		}

		// 取 feature 校验 is_enabled
		featureType, err := s.getEnabledFeatureType(ctx, it.GetFeatureId())
		if err != nil {
			s.log.Warnf("batch-add skip feature_id=%d: %v", it.GetFeatureId(), err)
			continue
		}
		// 若提供 spec 则校验
		if it.Spec != nil {
			if mismatch := validateSpecTypeMismatch(featureType, it.GetSpec()); mismatch != "" {
				s.log.Warnf("batch-add skip feature_id=%d: spec type mismatch %s", it.GetFeatureId(), mismatch)
				continue
			}
			if errs := validateFeatureSpecForType(featureType, it.GetSpec()); len(errs) != 0 {
				s.log.Warnf("batch-add skip feature_id=%d: spec invalid %v", it.GetFeatureId(), errs)
				continue
			}
		}

		// 构造单条 Create
		data := &thingmodelV1.CategoryDefaultFeature{
			CategoryId:  trans.Ptr(categoryID),
			FeatureId:   trans.Ptr(it.GetFeatureId()),
			Spec:        it.GetSpec(),
			DisplayName: it.DisplayName,
			SortOrder:   it.SortOrder,
			TenantId:    &tenantID,
			CreatedBy:   trans.Ptr(operator.UserId),
		}
		syncSpecializedColumnsCDF(data)
		if _, err := s.repo.Create(ctx, &thingmodelV1.CreateCategoryDefaultFeatureRequest{Data: data}); err != nil {
			return nil, err
		}

		// 维护 unit.ref count
		if uid := extractPropertyUnitIDFromSpec(it.GetSpec()); uid > 0 {
			if err := s.unitRepo.IncReferenceCount(ctx, uid, +1); err != nil {
				s.log.Warnf("inc unit ref count for batch-add (feature %d unit %d) failed: %v", it.GetFeatureId(), uid, err)
			}
		}

		// 注：repo.Create 不返回 row（CR-001 重构后未暴露 row 给 service）。
		// 这里跳过 resp.Created 的填充——前端 BatchAdd 后通过 List 重新拉取展示。
	}
	return resp, nil
}

// Update 更新 spec/display_name/sort_order/is_enabled
//
// 若 spec.unit 改变，需要交换 unit.reference_count（旧 -1 / 新 +1）。
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

	// 诊断日志：CR-001 跟进 bug（保存后 spec 看不到）/ Diagnostic for "spec lost on save".
	specStr := "<nil>"
	if req.Data.Spec != nil {
		if b, err := protojson.Marshal(req.Data.Spec); err == nil {
			specStr = string(b)
		} else {
			specStr = "<marshal-err: " + err.Error() + ">"
		}
	}
	s.log.Infof("[CDF.Update] id=%d mask=%v spec=%s",
		req.GetId(), req.GetUpdateMask().GetPaths(), specStr)

	// 取旧 DTO（含 Spec）
	oldDTO, err := s.repo.Get(ctx, &thingmodelV1.GetCategoryDefaultFeatureRequest{Id: req.GetId()})
	if err != nil {
		return nil, err
	}
	if oldDTO == nil {
		return nil, thingmodelV1.ErrorCatDefaultFeatureNotFound("not found")
	}

	// 校验新 spec（若提供）
	if req.Data.Spec != nil {
		featDTO, err := s.featureRepo.Get(ctx, &thingmodelV1.GetFeatureRequest{
			QueryBy: &thingmodelV1.GetFeatureRequest_Id{Id: oldDTO.GetFeatureId()},
		})
		if err != nil {
			return nil, err
		}
		ft := featDTO.GetFeatureType()
		if mismatch := validateSpecTypeMismatch(ft, req.Data.GetSpec()); mismatch != "" {
			return nil, thingmodelV1.ErrorFeatureSpecInvalid("%s", mismatch)
		}
		if errs := validateFeatureSpecForType(ft, req.Data.GetSpec()); len(errs) != 0 {
			return nil, thingmodelV1.ErrorFeatureSpecInvalid("spec invalid: %v", errs)
		}
		syncSpecializedColumnsCDF(req.Data)
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	// 若 spec.unit 变化，交换 ref count
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
			continue
		}
		// CR-001：unit_id 直接从 CDF.spec 取（不再依赖 feature.spec）
		if uid := extractPropertyUnitIDFromSpec(dto.GetSpec()); uid > 0 {
			_ = s.unitRepo.IncReferenceCount(ctx, uid, -1)
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

// getEnabledFeatureType 取启用状态的全局特征类型；停用则返回错误。
// CR-001：feature 不再有 spec，只需返回 feature_type 给 CDF 用作校验分派。
func (s *CategoryDefaultFeatureService) getEnabledFeatureType(ctx context.Context, featureID uint32) (thingmodelV1.FeatureType, error) {
	if featureID == 0 {
		return thingmodelV1.FeatureType_FEATURE_TYPE_UNSPECIFIED, thingmodelV1.ErrorBadRequest("feature_id is required")
	}
	feat, err := s.featureRepo.Get(ctx, &thingmodelV1.GetFeatureRequest{
		QueryBy: &thingmodelV1.GetFeatureRequest_Id{Id: featureID},
	})
	if err != nil {
		return thingmodelV1.FeatureType_FEATURE_TYPE_UNSPECIFIED, thingmodelV1.ErrorFeatureNotFound("feature not found")
	}
	if feat.IsEnabled != nil && !feat.GetIsEnabled() {
		return thingmodelV1.FeatureType_FEATURE_TYPE_UNSPECIFIED, thingmodelV1.ErrorCatDefaultFeatureFeatureDisabled("feature %s is disabled", feat.GetCode())
	}
	return feat.GetFeatureType(), nil
}
