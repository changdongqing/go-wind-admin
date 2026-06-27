package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-admin/app/admin/service/internal/data"
	"go-wind-admin/app/admin/service/internal/data/seed"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"

	"go-wind-admin/pkg/middleware/auth"
)

// FeatureService 特征服务 / Feature service
//
// 设计依据 / Design ref: docs/thingmodel/sheji/12-特征后端实现设计.md §3
// 核心职责 / Core responsibilities:
//   - CRUD（透传 repo）
//   - 写入前 spec 校验（feature_validator.go）
//   - 特化列与 spec 一致性同步（约束 F4/F17）
//   - 单位引用计数维护（property 引用 unit 时调 unitRepo.IncReferenceCount）
//   - 关系完整性校验（删除前查 ReferencedByRelation，约束 F11）
type FeatureService struct {
	adminV1.FeatureServiceHTTPServer

	log      *log.Helper
	repo     *data.FeatureRepo
	unitRepo *data.UnitRepo
}

// NewFeatureService 构造特征服务
func NewFeatureService(
	ctx *bootstrap.Context,
	repo *data.FeatureRepo,
	unitRepo *data.UnitRepo,
) *FeatureService {
	return &FeatureService{
		log:      ctx.NewLoggerHelper("feature/service/admin-service"),
		repo:     repo,
		unitRepo: unitRepo,
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

// ValidateSpec 校验 spec（不落库，前端表单实时校验用）/ Validate spec without persisting
func (s *FeatureService) ValidateSpec(ctx context.Context, req *thingmodelV1.ValidateFeatureSpecRequest) (*thingmodelV1.ValidateFeatureSpecResponse, error) {
	errs := validateFeatureSpecForType(req.GetFeatureType(), req.GetSpec())
	// 关系目标存在性的 DB 校验（V15）—— 仅当 source/target 指向同表 feature 时
	errs = append(errs, s.validateRelationTargets(ctx, req.GetSpec())...)
	return &thingmodelV1.ValidateFeatureSpecResponse{
		Valid:  len(errs) == 0,
		Errors: errs,
	}, nil
}

// Create 创建特征 / Create feature
func (s *FeatureService) Create(ctx context.Context, req *thingmodelV1.CreateFeatureRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	if req.Data.FeatureType == nil {
		return nil, thingmodelV1.ErrorFeatureTypeInvalid("featureType required")
	}

	// 1. 校验 spec（按 oneof 分支分派）
	if errs := validateFeatureSpecForType(req.Data.GetFeatureType(), req.Data.GetSpec()); len(errs) > 0 {
		return nil, thingmodelV1.ErrorFeatureSpecInvalid("%v", errs)
	}
	// 1.1 关系目标存在性校验
	if rErrs := s.validateRelationTargets(ctx, req.Data.GetSpec()); len(rErrs) > 0 {
		return nil, thingmodelV1.ErrorFeatureRelationTargetNotFound("%v", rErrs)
	}
	// 2. 同步特化列（spec → 抽取列，保证一致性，约束 F4/F17）
	syncSpecializedColumns(req.Data)

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	if err = s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	// 3. 若是 property 且引用了 unit，维护 unit.reference_count +1（约束 F6/F10）
	s.adjustUnitReference(ctx, req.Data, +1)

	return &emptypb.Empty{}, nil
}

// Update 更新特征 / Update feature
func (s *FeatureService) Update(ctx context.Context, req *thingmodelV1.UpdateFeatureRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	// 仅当请求带了 spec 时才做 spec 校验（FieldMask 部分更新可不带 spec）
	if req.Data.Spec != nil {
		// 优先用 data.feature_type；若未带，则需先从 DB 取出当前 feature_type 来决定校验分支
		ft := req.Data.GetFeatureType()
		if ft == thingmodelV1.FeatureType_FEATURE_TYPE_UNSPECIFIED {
			cur, _ := s.repo.Get(ctx, &thingmodelV1.GetFeatureRequest{
				QueryBy: &thingmodelV1.GetFeatureRequest_Id{Id: req.GetId()},
			})
			if cur != nil {
				ft = cur.GetFeatureType()
			}
		}
		if errs := validateFeatureSpecForType(ft, req.Data.GetSpec()); len(errs) > 0 {
			return nil, thingmodelV1.ErrorFeatureSpecInvalid("%v", errs)
		}
		if rErrs := s.validateRelationTargets(ctx, req.Data.GetSpec()); len(rErrs) > 0 {
			return nil, thingmodelV1.ErrorFeatureRelationTargetNotFound("%v", rErrs)
		}
		// 同步特化列
		syncSpecializedColumns(req.Data)
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

	// property 的 unit 引用变更：旧 unit -1，新 unit +1
	s.handleUnitReferenceChange(ctx, req)

	if err = s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Delete 批量删除 / Batch delete
//
// 删除前置检查：
//   - 约束 F11：被 RELATION 引用的 feature 拒绝删除（提示用户先解除关系）
//   - 约束 F10：被删 property 引用的 unit，reference_count -1（兑现单位管理挂钩点）
func (s *FeatureService) Delete(ctx context.Context, req *thingmodelV1.DeleteFeatureRequest) (*emptypb.Empty, error) {
	ids := req.GetIds()
	if len(ids) == 0 {
		return nil, thingmodelV1.ErrorBadRequest("ids is required")
	}

	// 1. 关系引用守卫
	for _, id := range ids {
		referenced, err := s.repo.ReferencedByRelation(ctx, id)
		if err != nil {
			return nil, err
		}
		if referenced {
			return nil, thingmodelV1.ErrorFeatureInUseCannotDelete(
				"feature %d is referenced by relation(s), remove relations first", id)
		}
	}

	// 2. 单位引用计数维护（删除前预取被删 property 的 unit 引用）
	for _, id := range ids {
		feat, _ := s.repo.Get(ctx, &thingmodelV1.GetFeatureRequest{
			QueryBy: &thingmodelV1.GetFeatureRequest_Id{Id: id},
		})
		if feat != nil {
			s.adjustUnitReference(ctx, feat, -1)
		}
	}

	// 3. 物理删除
	if err := s.repo.BatchDelete(ctx, ids); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// ImportFeatures 批量导入特征（保底方案：种子未初始化或需补录时用 Excel 导入）。
// ImportFeatures imports features in bulk (idempotent by code).
//
// 数据来源：前端解析 Excel 后按行提交（见 ImportFeatureRow）。
// specJson 是与种子 map 同构的 JSON 字符串，这里复用 seed.BuildFeatureSpecFromMap
// 还原为 proto FeatureSpec oneof，保证导入与种子走同一套 spec 构造/校验逻辑。
//
// 行为：
//   - 每行先做 spec 校验（validateFeatureSpecForType）+ 特化列同步，再按 code 幂等 upsert；
//   - skip_invalid=true：单行失败收集后继续，末尾汇总（默认，推荐用于大批量导入）；
//   - skip_invalid=false：遇第一条错误即整批中止返回。
//   - identifier 的 (tenant_id, identifier) 唯一性由 DB 索引兜底；导入前不额外预检
//     （与种子同策略），重复会在该行报错并按 skip_invalid 处理。
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
	skipInvalid := req.GetSkipInvalid() // 默认 false；前端默认传 true

	var errs []string
	for i, row := range rows {
		code := row.GetCode()
		if code == "" {
			// 行号从 2 起（Excel 第 1 行是表头）
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
				return nil, thingmodelV1.ErrorFeatureSpecInvalid("%s: %v", code, perr)
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
	// 失败明细截断到前 20 条，避免响应体过大
	if len(errs) > 20 {
		resp.Errors = append(resp.Errors, errs[:20]...)
		resp.Errors = append(resp.Errors, fmt.Sprintf("...共 %d 条失败，仅展示前 20 条", len(errs)))
	} else {
		resp.Errors = errs
	}
	return resp, nil
}

// buildFeatureFromImportRow 把单行导入数据组装为可落库的 Feature：
// 解析 specJson → 构造 FeatureSpec oneof → spec 校验 → 同步特化列 → 设置公共字段。
// 纯函数（不访问 DB / 不依赖 receiver），便于单测覆盖导入解析与校验逻辑。
func buildFeatureFromImportRow(_ context.Context, row *thingmodelV1.ImportFeatureRow, userID uint32) (*thingmodelV1.Feature, error) {
	featureType := seedParseFeatureType(row.GetFeatureType())
	if featureType == thingmodelV1.FeatureType_FEATURE_TYPE_UNSPECIFIED {
		return nil, fmt.Errorf("featureType 非法: %q", row.GetFeatureType())
	}

	// 解析 specJson → map → proto FeatureSpec（复用种子构造逻辑）
	var specMap map[string]any
	if j := strings.TrimSpace(row.GetSpecJson()); j != "" {
		if err := json.Unmarshal([]byte(j), &specMap); err != nil {
			return nil, fmt.Errorf("specJson 解析失败: %v", err)
		}
	}
	specProto := seed.BuildFeatureSpecFromMap(row.GetFeatureType(), specMap)

	// spec 校验（与 Create 同一套规则）
	if errs := validateFeatureSpecForType(featureType, specProto); len(errs) > 0 {
		return nil, fmt.Errorf("spec 校验失败: %s", strings.Join(errs, "; "))
	}

	feature := &thingmodelV1.Feature{
		FeatureType:     &featureType,
		Code:            trans.Ptr(row.GetCode()),
		Identifier:      trans.Ptr(row.GetIdentifier()),
		Name:            trans.Ptr(row.GetName()),
		NameEn:          trans.Ptr(row.GetNameEn()),
		Description:     trans.Ptr(row.GetDescription()),
		ApplicableScope: trans.Ptr(row.GetApplicableScope()),
		SortOrder:       trans.Ptr(row.GetSortOrder()),
		IsEnabled:       trans.Ptr(true),
		Spec:            specProto,
		CreatedBy:       trans.Ptr(userID),
	}
	// 同步特化列（spec → 抽取列，与 Create 一致）
	syncSpecializedColumns(feature)
	return feature, nil
}

// seedParseFeatureType 解析 featureType 字符串为 proto 枚举（本地小工具，避免与 seed 包循环引用）。
func seedParseFeatureType(s string) thingmodelV1.FeatureType {
	if v, ok := thingmodelV1.FeatureType_value[s]; ok {
		return thingmodelV1.FeatureType(v)
	}
	return thingmodelV1.FeatureType_FEATURE_TYPE_UNSPECIFIED
}

// ===== 特化列同步 / Sync specialized columns =====

// syncSpecializedColumns 将 spec oneof 分支内的字段同步到特化列，保证一致性（约束 F4/F17）
// Sync specialized columns from spec oneof branches.
func syncSpecializedColumns(f *thingmodelV1.Feature) {
	if f == nil || f.Spec == nil {
		return
	}
	switch sp := f.Spec.Spec.(type) {
	case *thingmodelV1.FeatureSpec_Property:
		if sp.Property != nil {
			dt := sp.Property.GetDataType()
			f.DataType = &dt
			am := sp.Property.GetAccessMode()
			f.AccessMode = &am
		}
	case *thingmodelV1.FeatureSpec_Event:
		if sp.Event != nil {
			lv := sp.Event.GetLevel()
			f.EventLevel = &lv
		}
	case *thingmodelV1.FeatureSpec_Service:
		if sp.Service != nil {
			cm := sp.Service.GetCallMode()
			f.CallMode = &cm
		}
	case *thingmodelV1.FeatureSpec_Relation:
		if sp.Relation != nil {
			f.RelationType = trans.Ptr(sp.Relation.GetRelationType())
		}
	}
}

// ===== 单位引用计数维护（打通单位管理挂钩点）/ Unit reference maintenance =====

// adjustUnitReference 调整单位引用计数（delta 可正可负）。
// 仅对 property 且 spec.property.unit.unitId>0 的特征生效。
//
// 失败仅记日志，不阻断主流程（特征已落库，由对账任务兜底）。
func (s *FeatureService) adjustUnitReference(ctx context.Context, f *thingmodelV1.Feature, delta int32) {
	unitID := extractPropertyUnitID(f)
	if unitID == 0 {
		return
	}
	if err := s.unitRepo.IncReferenceCount(ctx, unitID, delta); err != nil {
		s.log.Errorf("adjust unit %d reference_count by %d failed: %v", unitID, delta, err)
	}
}

// handleUnitReferenceChange 更新 property 时，若 unit 引用变更，则旧 -1、新 +1
func (s *FeatureService) handleUnitReferenceChange(ctx context.Context, req *thingmodelV1.UpdateFeatureRequest) {
	if req == nil || req.Data == nil {
		return
	}
	// 仅 property 才涉及
	if req.Data.GetFeatureType() != thingmodelV1.FeatureType_PROPERTY &&
		req.Data.GetFeatureType() != thingmodelV1.FeatureType_FEATURE_TYPE_UNSPECIFIED {
		return
	}
	newUnitID := extractPropertyUnitID(req.Data)
	// 取旧数据
	cur, err := s.repo.Get(ctx, &thingmodelV1.GetFeatureRequest{
		QueryBy: &thingmodelV1.GetFeatureRequest_Id{Id: req.GetId()},
	})
	if err != nil || cur == nil {
		return
	}
	if cur.GetFeatureType() != thingmodelV1.FeatureType_PROPERTY {
		// 旧记录不是 property，只考虑新 +1
		if newUnitID > 0 {
			if err := s.unitRepo.IncReferenceCount(ctx, newUnitID, +1); err != nil {
				s.log.Errorf("adjust unit %d +1 failed: %v", newUnitID, err)
			}
		}
		return
	}
	oldUnitID := extractPropertyUnitID(cur)
	if oldUnitID == newUnitID {
		return
	}
	if oldUnitID > 0 {
		if err := s.unitRepo.IncReferenceCount(ctx, oldUnitID, -1); err != nil {
			s.log.Errorf("adjust unit %d -1 failed: %v", oldUnitID, err)
		}
	}
	if newUnitID > 0 {
		if err := s.unitRepo.IncReferenceCount(ctx, newUnitID, +1); err != nil {
			s.log.Errorf("adjust unit %d +1 failed: %v", newUnitID, err)
		}
	}
}

// extractPropertyUnitID 提取 property spec.unit.unitId（其它类型返回 0）
func extractPropertyUnitID(f *thingmodelV1.Feature) uint32 {
	if f == nil || f.Spec == nil {
		return 0
	}
	prop, ok := f.Spec.Spec.(*thingmodelV1.FeatureSpec_Property)
	if !ok || prop.Property == nil {
		return 0
	}
	u := prop.Property.GetUnit()
	if u == nil {
		return 0
	}
	return u.GetUnitId()
}

// ===== 关系目标存在性校验（V15）/ Relation target existence check =====

// validateRelationTargets 若 spec 是 relation 且 source/target.kind=feature，
// 则校验对应 id 在 feature 表内存在。kind=external 跳过（本期弱校验）。
// Returns a list of error strings; empty when no relation refs or all refs are valid.
func (s *FeatureService) validateRelationTargets(ctx context.Context, spec *thingmodelV1.FeatureSpec) []string {
	if spec == nil {
		return nil
	}
	rel, ok := spec.Spec.(*thingmodelV1.FeatureSpec_Relation)
	if !ok || rel.Relation == nil {
		return nil
	}
	var errs []string
	checkEnt := func(label string, ref *thingmodelV1.EntityRef) {
		if ref == nil {
			return
		}
		if ref.GetKind() != "feature" {
			return
		}
		id := ref.GetId()
		if id == 0 {
			errs = append(errs, label+": kind=feature but id=0")
			return
		}
		exist, err := s.repo.IsExist(ctx, id)
		if err != nil {
			errs = append(errs, label+": query existence failed")
			return
		}
		if !exist {
			errs = append(errs, label+": feature id "+itoa(id)+" not found")
		}
	}
	checkEnt("relation.source", rel.Relation.GetSource())
	checkEnt("relation.target", rel.Relation.GetTarget())
	return errs
}

// itoa 简洁 uint32→string（避免 strconv 引入）
func itoa(n uint32) string {
	if n == 0 {
		return "0"
	}
	var buf [10]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
