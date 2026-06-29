package service

import (
	"context"
	"fmt"

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

// FeatureService 特征服务（CR-001 后：仅承载骨架 CRUD）/ Feature service (skeleton CRUD only)
//
// 设计依据 / Design ref:
//   - docs/thingmodel/sheji/12-特征后端实现设计.md §3
//   - docs/thingmodel/sheji/修改记录/CR-001-结构化约束下沉到模型层.md
//
// CR-001（2026-06-29）后变更：
//   - 删除 ValidateSpec RPC 与 spec 校验链；
//   - 删除特化列同步与单位引用计数维护（spec 已下沉到 CDF/PF，由它们的 service 负责）；
//   - 删除关系目标存在性校验（移到 CDF/PF service）；
//   - ImportFeatures 仅导入骨架字段（不再含 spec_json）。
type FeatureService struct {
	adminV1.FeatureServiceHTTPServer

	log  *log.Helper
	repo *data.FeatureRepo
}

// NewFeatureService 构造特征服务
func NewFeatureService(
	ctx *bootstrap.Context,
	repo *data.FeatureRepo,
) *FeatureService {
	return &FeatureService{
		log:  ctx.NewLoggerHelper("feature/service/admin-service"),
		repo: repo,
	}
}

// List 分页查询 / List
func (s *FeatureService) List(ctx context.Context, req *paginationV1.PagingRequest) (*thingmodelV1.ListFeatureResponse, error) {
	return s.repo.List(ctx, req)
}

// Get 查询详情 / Get
func (s *FeatureService) Get(ctx context.Context, req *thingmodelV1.GetFeatureRequest) (*thingmodelV1.Feature, error) {
	return s.repo.Get(ctx, req)
}

// ListByType 按特征类型查询 / List by type
func (s *FeatureService) ListByType(ctx context.Context, req *thingmodelV1.ListFeatureByTypeRequest) (*thingmodelV1.ListFeatureResponse, error) {
	return s.repo.ListByType(ctx, req)
}

// Create 创建特征骨架 / Create feature skeleton
func (s *FeatureService) Create(ctx context.Context, req *thingmodelV1.CreateFeatureRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	if req.Data.FeatureType == nil {
		return nil, thingmodelV1.ErrorFeatureTypeInvalid("featureType required")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	if err = s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Update 更新特征骨架 / Update feature skeleton
func (s *FeatureService) Update(ctx context.Context, req *thingmodelV1.UpdateFeatureRequest) (*emptypb.Empty, error) {
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

	if err = s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Delete 批量删除 / Batch delete
//
// CR-001：删除前校验由 DB 层 OnDelete:Restrict 兜底（feature 被 CDF.feature_id 引用时
// 会直接拒绝）；spec.relation 内部 source/target 引用的细粒度完整性由 CDF/PF service 自查。
func (s *FeatureService) Delete(ctx context.Context, req *thingmodelV1.DeleteFeatureRequest) (*emptypb.Empty, error) {
	ids := req.GetIds()
	if len(ids) == 0 {
		return nil, thingmodelV1.ErrorBadRequest("ids is required")
	}

	if err := s.repo.BatchDelete(ctx, ids); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// ImportFeatures 批量导入特征骨架（CR-001 后仅含骨架字段）。
// ImportFeatures imports feature skeletons in bulk (idempotent by code).
//
// 行为：
//   - 每行 upsert 按 (tenant_id, code) 幂等；
//   - skip_invalid=true：单行失败收集后继续，末尾汇总；
//   - skip_invalid=false：遇第一条错误即整批中止返回。
func (s *FeatureService) ImportFeatures(ctx context.Context, req *thingmodelV1.ImportFeaturesRequest) (*thingmodelV1.ImportFeaturesResponse, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	rows := req.GetRows()
	resp := &thingmodelV1.ImportFeaturesResponse{Total: uint32(len(rows))}
	if len(rows) == 0 {
		return resp, nil
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	skipInvalid := req.GetSkipInvalid()

	var errs []string
	for i, row := range rows {
		code := row.GetCode()
		if code == "" {
			msg := fmt.Sprintf("第%d行: code 为空", i+2)
			if !skipInvalid {
				return nil, thingmodelV1.ErrorBadRequest("%s", msg)
			}
			errs = append(errs, msg)
			continue
		}

		feature, perr := buildFeatureFromImportRow(ctx, row, operator.UserId)
		if perr != nil {
			if !skipInvalid {
				return nil, thingmodelV1.ErrorFeatureTypeInvalid("%s: %v", code, perr)
			}
			errs = append(errs, fmt.Sprintf("%s: %v", code, perr))
			continue
		}

		if err := s.repo.UpsertByCode(ctx, feature); err != nil {
			if !skipInvalid {
				return nil, err
			}
			errs = append(errs, fmt.Sprintf("%s: %v", code, err))
			continue
		}
		resp.Succeeded++
	}
	resp.Failed = uint32(len(errs))
	if len(errs) > 20 {
		resp.Errors = append(resp.Errors, errs[:20]...)
		resp.Errors = append(resp.Errors, fmt.Sprintf("...共 %d 条失败，仅展示前 20 条", len(errs)))
	} else {
		resp.Errors = errs
	}
	return resp, nil
}

// buildFeatureFromImportRow 把单行导入数据组装为可落库的 Feature（仅骨架）。
// 纯函数（不访问 DB / 不依赖 receiver），便于单测覆盖。
//
// CR-001：不再处理 spec_json；recommended_unit_category_id / semantic_tag 由 ImportFeatureRow 直接承载。
func buildFeatureFromImportRow(_ context.Context, row *thingmodelV1.ImportFeatureRow, userID uint32) (*thingmodelV1.Feature, error) {
	featureType := parseImportFeatureType(row.GetFeatureType())
	if featureType == thingmodelV1.FeatureType_FEATURE_TYPE_UNSPECIFIED {
		return nil, fmt.Errorf("featureType 非法: %q", row.GetFeatureType())
	}

	feature := &thingmodelV1.Feature{
		FeatureType:               &featureType,
		Code:                      trans.Ptr(row.GetCode()),
		Identifier:                trans.Ptr(row.GetIdentifier()),
		Name:                      trans.Ptr(row.GetName()),
		NameEn:                    trans.Ptr(row.GetNameEn()),
		Description:               trans.Ptr(row.GetDescription()),
		ApplicableScope:           trans.Ptr(row.GetApplicableScope()),
		SortOrder:                 trans.Ptr(row.GetSortOrder()),
		RecommendedUnitCategoryId: trans.Ptr(row.GetRecommendedUnitCategoryId()),
		SemanticTag:               trans.Ptr(row.GetSemanticTag()),
		IsEnabled:                 trans.Ptr(true),
		CreatedBy:                 trans.Ptr(userID),
	}
	return feature, nil
}

// parseImportFeatureType 解析 featureType 字符串为 proto 枚举。
func parseImportFeatureType(s string) thingmodelV1.FeatureType {
	if v, ok := thingmodelV1.FeatureType_value[s]; ok {
		return thingmodelV1.FeatureType(v)
	}
	return thingmodelV1.FeatureType_FEATURE_TYPE_UNSPECIFIED
}
