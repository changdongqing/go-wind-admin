package data

import (
	"context"
	"encoding/json"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/jinzhu/copier"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"

	"go-wind-admin/app/admin/service/internal/data/ent"
	"go-wind-admin/app/admin/service/internal/data/ent/categorydefaultfeature"
	"go-wind-admin/app/admin/service/internal/data/ent/feature"
	"go-wind-admin/app/admin/service/internal/data/ent/predicate"
	"go-wind-admin/app/admin/service/internal/data/ent/schema"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// CategoryDefaultFeatureRepo 分类默认模型条目仓库 / Category default feature repository
//
// 设计依据 / Design ref:
//   - docs/thingmodel/sheji/模型管理/04-后端实现设计.md §1
//   - docs/thingmodel/sheji/修改记录/CR-001-结构化约束下沉到模型层.md
//
// CR-001（2026-06-29）后变更：
//   - override_spec(FeatureOverrideSpec) → spec(FeatureSpec)：承载完整结构化约束；
//   - 新增 5 个冗余特化列读写（data_type/access_mode/event_level/call_mode/relation_type）；
//   - reference_count 维护仍由 service 层负责（事务内 ±1 thingmodel_units）。
type CategoryDefaultFeatureRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[thingmodelV1.CategoryDefaultFeature, ent.CategoryDefaultFeature]

	// 5 个冗余特化列的 enum 转换器
	dataTypeConverter   *mapper.EnumTypeConverter[thingmodelV1.DataType, categorydefaultfeature.DataType]
	accessModeConverter *mapper.EnumTypeConverter[thingmodelV1.AccessMode, categorydefaultfeature.AccessMode]
	eventLevelConverter *mapper.EnumTypeConverter[thingmodelV1.EventLevel, categorydefaultfeature.EventLevel]
	callModeConverter   *mapper.EnumTypeConverter[thingmodelV1.CallMode, categorydefaultfeature.CallMode]

	repository *entCrud.Repository[
		ent.CategoryDefaultFeatureQuery, ent.CategoryDefaultFeatureSelect,
		ent.CategoryDefaultFeatureCreate, ent.CategoryDefaultFeatureCreateBulk,
		ent.CategoryDefaultFeatureUpdate, ent.CategoryDefaultFeatureUpdateOne,
		ent.CategoryDefaultFeatureDelete,
		predicate.CategoryDefaultFeature,
		thingmodelV1.CategoryDefaultFeature, ent.CategoryDefaultFeature,
	]
}

// NewCategoryDefaultFeatureRepo 构造仓库
func NewCategoryDefaultFeatureRepo(
	ctx *bootstrap.Context,
	entClient *entCrud.EntClient[*ent.Client],
) *CategoryDefaultFeatureRepo {
	repo := &CategoryDefaultFeatureRepo{
		log:       ctx.NewLoggerHelper("category-default-feature/repo/admin-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[thingmodelV1.CategoryDefaultFeature, ent.CategoryDefaultFeature](),

		dataTypeConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.DataType, categorydefaultfeature.DataType,
		](thingmodelV1.DataType_name, thingmodelV1.DataType_value),
		accessModeConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.AccessMode, categorydefaultfeature.AccessMode,
		](thingmodelV1.AccessMode_name, thingmodelV1.AccessMode_value),
		eventLevelConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.EventLevel, categorydefaultfeature.EventLevel,
		](thingmodelV1.EventLevel_name, thingmodelV1.EventLevel_value),
		callModeConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.CallMode, categorydefaultfeature.CallMode,
		](thingmodelV1.CallMode_name, thingmodelV1.CallMode_value),
	}
	repo.init()
	return repo
}

func (r *CategoryDefaultFeatureRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.CategoryDefaultFeatureQuery, ent.CategoryDefaultFeatureSelect,
		ent.CategoryDefaultFeatureCreate, ent.CategoryDefaultFeatureCreateBulk,
		ent.CategoryDefaultFeatureUpdate, ent.CategoryDefaultFeatureUpdateOne,
		ent.CategoryDefaultFeatureDelete,
		predicate.CategoryDefaultFeature,
		thingmodelV1.CategoryDefaultFeature, ent.CategoryDefaultFeature,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	// CR-001：spec 字段 protojson 双向转换器
	r.mapper.AppendConverters(cdfSpecConverterPair())

	// 5 个冗余特化列的枚举转换器
	r.mapper.AppendConverters(r.dataTypeConverter.NewConverterPair())
	r.mapper.AppendConverters(r.accessModeConverter.NewConverterPair())
	r.mapper.AppendConverters(r.eventLevelConverter.NewConverterPair())
	r.mapper.AppendConverters(r.callModeConverter.NewConverterPair())
}

// cdfSpecConverterPair 返回 CDF.spec 字段的双向类型转换对：
// entity *schema.FeatureSpecField ↔ dto *thingmodelV1.FeatureSpec
func cdfSpecConverterPair() []copier.TypeConverter {
	return []copier.TypeConverter{
		// entity → dto
		{
			SrcType: (*schema.FeatureSpecField)(nil),
			DstType: (*thingmodelV1.FeatureSpec)(nil),
			Fn: func(src interface{}) (interface{}, error) {
				f, _ := src.(*schema.FeatureSpecField)
				return schema.UnwrapFeatureSpec(f), nil
			},
		},
		// dto → entity
		{
			SrcType: (*thingmodelV1.FeatureSpec)(nil),
			DstType: (*schema.FeatureSpecField)(nil),
			Fn: func(src interface{}) (interface{}, error) {
				s, _ := src.(*thingmodelV1.FeatureSpec)
				return schema.WrapFeatureSpec(s), nil
			},
		},
	}
}

// ===== CRUD =====

// Count 统计数量
func (r *CategoryDefaultFeatureRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().CategoryDefaultFeature.Query()
	if len(whereCond) != 0 {
		builder.Modify(whereCond...)
	}
	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, thingmodelV1.ErrorInternalServerError("query count failed")
	}
	return count, nil
}

// List 分页查询
//
// 过滤特殊处理：
//   - feature_type 不是本表的列（位于关联 feature 表）。前端传入时会被 go-crud 当作未知列触发
//     SQL 异常导致 500，因此这里预先剥离并翻译为 HasFeatureWith(feature.FeatureTypeEQ) 谓词。
//
// 返回值增强：
//   - 列表返回的 DTO 中 feature_code/feature_identifier/feature_name/feature_type
//     四个只读字段，由本方法通过 feature_id 批量回查 feature 表后回填。
func (r *CategoryDefaultFeatureRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*thingmodelV1.ListCategoryDefaultFeatureResponse, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	// 拆掉 query 中的 feature_type，转化为 builder 上的 edge 谓词；避免 SQL 报"unknown column"。
	extraPreds := r.translateForeignFilters(req)

	builder := r.entClient.Client().CategoryDefaultFeature.Query()
	if len(extraPreds) > 0 {
		builder = builder.Where(extraPreds...)
	}
	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &thingmodelV1.ListCategoryDefaultFeatureResponse{Total: 0, Items: nil}, nil
	}

	// 联表回填只读字段
	if err := r.enrichItemsWithFeatures(ctx, ret.Items); err != nil {
		r.log.Warnf("enrich category_default_features with feature fields failed: %v", err)
	}

	return &thingmodelV1.ListCategoryDefaultFeatureResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

// translateForeignFilters 把 query JSON 中"不属于本表列"的字段剥离出来，转换为 edge 谓词。
// 目前仅处理 feature_type；未来如需新增（如 feature_code/feature_identifier 模糊搜索），按相同范式追加。
func (r *CategoryDefaultFeatureRepo) translateForeignFilters(req *paginationV1.PagingRequest) []predicate.CategoryDefaultFeature {
	q := req.GetQuery()
	if q == "" {
		return nil
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(q), &raw); err != nil {
		return nil
	}
	var preds []predicate.CategoryDefaultFeature
	if v, ok := raw["feature_type"]; ok {
		delete(raw, "feature_type")
		if s, ok := v.(string); ok && s != "" {
			preds = append(preds, categorydefaultfeature.HasFeatureWith(
				feature.FeatureTypeEQ(feature.FeatureType(s)),
			))
		}
	}
	// 回写剥离后的 query；如果完全空了，把 oneof 重置为 nil 以避免后端解析空对象出错
	if len(raw) == 0 {
		req.FilteringType = nil
	} else {
		newQ, _ := json.Marshal(raw)
		req.FilteringType = &paginationV1.PagingRequest_Query{Query: string(newQ)}
	}
	return preds
}

// enrichItemsWithFeatures 批量回查 feature 表，回填 feature_code/feature_identifier/feature_name/feature_type 4 个只读字段。
func (r *CategoryDefaultFeatureRepo) enrichItemsWithFeatures(ctx context.Context, items []*thingmodelV1.CategoryDefaultFeature) error {
	if len(items) == 0 {
		return nil
	}
	idSet := make(map[uint32]struct{}, len(items))
	for _, it := range items {
		if it.GetFeatureId() != 0 {
			idSet[it.GetFeatureId()] = struct{}{}
		}
	}
	if len(idSet) == 0 {
		return nil
	}
	ids := make([]uint32, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	rows, err := r.entClient.Client().Feature.Query().
		Where(feature.IDIn(ids...)).
		All(ctx)
	if err != nil {
		return err
	}
	byID := make(map[uint32]*ent.Feature, len(rows))
	for _, f := range rows {
		byID[f.ID] = f
	}
	for _, it := range items {
		f, ok := byID[it.GetFeatureId()]
		if !ok {
			continue
		}
		if f.Code != nil {
			c := *f.Code
			it.FeatureCode = &c
		}
		if f.Identifier != nil {
			id := *f.Identifier
			it.FeatureIdentifier = &id
		}
		if f.Name != nil {
			n := *f.Name
			it.FeatureName = &n
		}
		if f.FeatureType != nil {
			if v, ok := thingmodelV1.FeatureType_value[string(*f.FeatureType)]; ok {
				ft := thingmodelV1.FeatureType(v)
				it.FeatureType = &ft
			}
		}
	}
	return nil
}

// ListByCategory 按 category_id 查询某个分类的全部默认条目（不分页）。
// service 层 PullFromDefault 使用此方法获取候选条目。
func (r *CategoryDefaultFeatureRepo) ListByCategory(
	ctx context.Context,
	categoryID uint32,
	defaultFeatureIDs []uint32,
	tenantID uint32,
) ([]*ent.CategoryDefaultFeature, error) {
	if categoryID == 0 {
		return nil, thingmodelV1.ErrorBadRequest("category_id is required")
	}
	builder := r.entClient.Client().CategoryDefaultFeature.Query().
		Where(
			categorydefaultfeature.CategoryIDEQ(categoryID),
			categorydefaultfeature.TenantIDEQ(tenantID),
		)
	if len(defaultFeatureIDs) > 0 {
		builder.Where(categorydefaultfeature.IDIn(defaultFeatureIDs...))
	}
	builder.Order(ent.Asc(categorydefaultfeature.FieldSortOrder), ent.Asc(categorydefaultfeature.FieldID))
	rows, err := builder.All(ctx)
	if err != nil {
		r.log.Errorf("list by category failed: %s", err.Error())
		return nil, thingmodelV1.ErrorInternalServerError("list by category failed")
	}
	return rows, nil
}

// IsExist 是否存在
func (r *CategoryDefaultFeatureRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().CategoryDefaultFeature.Query().
		Where(categorydefaultfeature.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, thingmodelV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

// ExistsByCategoryFeature 检查 (tenant, category_id, feature_id) 是否已存在；用于 service 冲突检测。
func (r *CategoryDefaultFeatureRepo) ExistsByCategoryFeature(
	ctx context.Context, tenantID, categoryID, featureID uint32,
) (bool, error) {
	return r.entClient.Client().CategoryDefaultFeature.Query().
		Where(
			categorydefaultfeature.TenantIDEQ(tenantID),
			categorydefaultfeature.CategoryIDEQ(categoryID),
			categorydefaultfeature.FeatureIDEQ(featureID),
		).
		Exist(ctx)
}

// Get 查询详情
func (r *CategoryDefaultFeatureRepo) Get(ctx context.Context, req *thingmodelV1.GetCategoryDefaultFeatureRequest) (*thingmodelV1.CategoryDefaultFeature, error) {
	if req == nil || req.GetId() == 0 {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	builder := r.entClient.Client().CategoryDefaultFeature.Query()
	whereCond := []func(s *sql.Selector){
		func(s *sql.Selector) { s.Where(sql.EQ(categorydefaultfeature.FieldID, req.GetId())) },
	}
	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}
	return dto, nil
}

// Create 创建单条
func (r *CategoryDefaultFeatureRepo) Create(ctx context.Context, req *thingmodelV1.CreateCategoryDefaultFeatureRequest) (created *ent.CategoryDefaultFeature, err error) {
	if req == nil || req.Data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	return r.createTx(ctx, nil, req.Data, time.Now())
}

// createTx 真正的创建逻辑（可在外部事务中复用）。tx 为 nil 时使用 r.entClient.Client()。
func (r *CategoryDefaultFeatureRepo) createTx(ctx context.Context, tx *ent.Tx, data *thingmodelV1.CategoryDefaultFeature, now time.Time) (*ent.CategoryDefaultFeature, error) {
	var builder *ent.CategoryDefaultFeatureCreate
	if tx != nil {
		builder = tx.CategoryDefaultFeature.Create()
	} else {
		builder = r.entClient.Client().CategoryDefaultFeature.Create()
	}

	builder.
		SetNillableTenantID(data.TenantId).
		SetCategoryID(data.GetCategoryId()).
		SetFeatureID(data.GetFeatureId()).
		SetNillableDisplayName(data.DisplayName).
		SetNillableSortOrder(data.SortOrder).
		SetNillableIsEnabled(data.IsEnabled).
		SetNillableRelationType(data.RelationType).
		SetNillableCreatedBy(data.CreatedBy).
		SetCreatedAt(now)

	// CR-001：完整 FeatureSpec
	if data.Spec != nil {
		builder.SetSpec(schema.WrapFeatureSpec(data.Spec))
	}
	// 冗余特化列（service 在写入前应已同步好，repo 不强制再次派生）
	if data.DataType != nil {
		if v, ok := protoToEntCDFDataType(data.GetDataType()); ok {
			builder.SetDataType(v)
		}
	}
	if data.AccessMode != nil {
		if v, ok := protoToEntCDFAccessMode(data.GetAccessMode()); ok {
			builder.SetAccessMode(v)
		}
	}
	if data.EventLevel != nil {
		if v, ok := protoToEntCDFEventLevel(data.GetEventLevel()); ok {
			builder.SetEventLevel(v)
		}
	}
	if data.CallMode != nil {
		if v, ok := protoToEntCDFCallMode(data.GetCallMode()); ok {
			builder.SetCallMode(v)
		}
	}

	row, err := builder.Save(ctx)
	if err != nil {
		r.log.Errorf("insert category_default_feature failed: %s", err.Error())
		if ent.IsConstraintError(err) {
			return nil, thingmodelV1.ErrorCatDefaultFeatureDuplicate("category default feature already exists")
		}
		return nil, thingmodelV1.ErrorInternalServerError("insert category_default_feature failed")
	}
	return row, nil
}

// CreateInTx 提供给 service 层在事务中调用（同事务维护 reference_count）。
func (r *CategoryDefaultFeatureRepo) CreateInTx(ctx context.Context, tx *ent.Tx, data *thingmodelV1.CategoryDefaultFeature) (*ent.CategoryDefaultFeature, error) {
	if tx == nil || data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	return r.createTx(ctx, tx, data, time.Now())
}

// Update 更新（FieldMask 控制覆盖字段：override_spec/display_name/is_enabled/sort_order）
func (r *CategoryDefaultFeatureRepo) Update(ctx context.Context, req *thingmodelV1.UpdateCategoryDefaultFeatureRequest) (err error) {
	if req == nil || req.Data == nil {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return thingmodelV1.ErrorBadRequest("id is required")
	}

	if req.GetAllowMissing() {
		var exist bool
		exist, err = r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &thingmodelV1.CreateCategoryDefaultFeatureRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			_, err = r.Create(ctx, createReq)
			return err
		}
	}

	var tx *ent.Tx
	tx, err = r.entClient.Client().Tx(ctx)
	if err != nil {
		r.log.Errorf("start transaction failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("start transaction failed")
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.log.Errorf("transaction rollback failed: %s", rollbackErr.Error())
			}
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			r.log.Errorf("transaction commit failed: %s", commitErr.Error())
			err = thingmodelV1.ErrorInternalServerError("transaction commit failed")
		}
	}()

	builder := tx.CategoryDefaultFeature.UpdateOneID(req.GetId())

	// CR-001 跟进：绕过 r.repository.UpdateOne 的 FilterByFieldMask 路径——
	// 该工具会按 update_mask 清空 dto 上非 mask 字段，但对含 oneof 的 message 字段
	// （FeatureSpec.spec oneof）处理存在反射边缘问题，可能把 dto.Spec 内部的 oneof
	// 分支清零，导致写入失败。这里直接按 update_mask 显式 set 字段，更可控。
	mask := req.GetUpdateMask().GetPaths()
	maskSet := make(map[string]struct{}, len(mask))
	for _, p := range mask {
		// 接受 snake_case 与 lowerCamel 两种形态，规范化为 snake_case
		maskSet[normalizeMaskPath(p)] = struct{}{}
	}
	in := func(name string) bool { _, ok := maskSet[name]; return ok }

	if in("display_name") {
		if req.Data.DisplayName != nil {
			builder.SetDisplayName(req.Data.GetDisplayName())
		} else {
			builder.ClearDisplayName()
		}
	}
	if in("sort_order") && req.Data.SortOrder != nil {
		builder.SetSortOrder(req.Data.GetSortOrder())
	}
	if in("is_enabled") && req.Data.IsEnabled != nil {
		builder.SetIsEnabled(req.Data.GetIsEnabled())
	}
	if in("relation_type") {
		if req.Data.RelationType != nil {
			builder.SetRelationType(req.Data.GetRelationType())
		} else {
			builder.ClearRelationType()
		}
	}
	if in("updated_by") && req.Data.UpdatedBy != nil {
		builder.SetUpdatedBy(req.Data.GetUpdatedBy())
	}
	if in("spec") {
		if req.Data.Spec != nil {
			r.log.Infof("[CDF.Update] id=%d SetSpec property=%v event=%v service=%v relation=%v",
				req.GetId(),
				req.Data.Spec.GetProperty() != nil,
				req.Data.Spec.GetEvent() != nil,
				req.Data.Spec.GetService() != nil,
				req.Data.Spec.GetRelation() != nil)
			builder.SetSpec(schema.WrapFeatureSpec(req.Data.Spec))
		} else {
			r.log.Warnf("[CDF.Update] id=%d mask contains spec but data.spec is nil → ClearSpec", req.GetId())
			builder.ClearSpec()
		}
	} else {
		r.log.Infof("[CDF.Update] id=%d mask does NOT contain 'spec' (mask=%v)", req.GetId(), mask)
	}
	// 冗余特化列（spec 改时一般 service 已同步 syncSpecializedColumnsCDF 写到 req.Data）
	if in("data_type") {
		if req.Data.DataType != nil {
			if v, ok := protoToEntCDFDataType(req.Data.GetDataType()); ok {
				builder.SetDataType(v)
			} else {
				builder.ClearDataType()
			}
		} else {
			builder.ClearDataType()
		}
	}
	if in("access_mode") {
		if req.Data.AccessMode != nil {
			if v, ok := protoToEntCDFAccessMode(req.Data.GetAccessMode()); ok {
				builder.SetAccessMode(v)
			} else {
				builder.ClearAccessMode()
			}
		} else {
			builder.ClearAccessMode()
		}
	}
	if in("event_level") {
		if req.Data.EventLevel != nil {
			if v, ok := protoToEntCDFEventLevel(req.Data.GetEventLevel()); ok {
				builder.SetEventLevel(v)
			} else {
				builder.ClearEventLevel()
			}
		} else {
			builder.ClearEventLevel()
		}
	}
	if in("call_mode") {
		if req.Data.CallMode != nil {
			if v, ok := protoToEntCDFCallMode(req.Data.GetCallMode()); ok {
				builder.SetCallMode(v)
			} else {
				builder.ClearCallMode()
			}
		} else {
			builder.ClearCallMode()
		}
	}
	// 当 spec 改了，强制同步派生的冗余列（service.Update 已在 req.Data 上同步过冗余列）
	if in("spec") {
		if req.Data.DataType != nil {
			if v, ok := protoToEntCDFDataType(req.Data.GetDataType()); ok {
				builder.SetDataType(v)
			}
		}
		if req.Data.AccessMode != nil {
			if v, ok := protoToEntCDFAccessMode(req.Data.GetAccessMode()); ok {
				builder.SetAccessMode(v)
			}
		}
		if req.Data.EventLevel != nil {
			if v, ok := protoToEntCDFEventLevel(req.Data.GetEventLevel()); ok {
				builder.SetEventLevel(v)
			}
		}
		if req.Data.CallMode != nil {
			if v, ok := protoToEntCDFCallMode(req.Data.GetCallMode()); ok {
				builder.SetCallMode(v)
			}
		}
		if req.Data.RelationType != nil {
			builder.SetRelationType(req.Data.GetRelationType())
		}
	}

	builder.SetUpdatedAt(time.Now())

	if _, err = builder.Save(ctx); err != nil {
		r.log.Errorf("update category_default_feature failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("update category_default_feature failed")
	}
	return nil
}

// Delete 删除单条（不维护 reference_count；由 service 在事务中维护）
func (r *CategoryDefaultFeatureRepo) Delete(ctx context.Context, id uint32) error {
	if id == 0 {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	if err := r.entClient.Client().CategoryDefaultFeature.DeleteOneID(id).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return thingmodelV1.ErrorCatDefaultFeatureNotFound("category default feature not found")
		}
		r.log.Errorf("delete category_default_feature failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

// DeleteInTx 在外部事务中删除（service 层用，便于同事务维护 reference_count）。
func (r *CategoryDefaultFeatureRepo) DeleteInTx(ctx context.Context, tx *ent.Tx, id uint32) error {
	if tx == nil || id == 0 {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	if err := tx.CategoryDefaultFeature.DeleteOneID(id).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return thingmodelV1.ErrorCatDefaultFeatureNotFound("category default feature not found")
		}
		r.log.Errorf("delete in tx failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

// DeleteBatch 批量删除（service 应先逐条 take entity 以维护 reference_count，
// 再调用本方法做实际删除——不在本方法做事务，调用方决定语义）
func (r *CategoryDefaultFeatureRepo) DeleteBatch(ctx context.Context, ids []uint32) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	n, err := r.entClient.Client().CategoryDefaultFeature.Delete().
		Where(categorydefaultfeature.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		r.log.Errorf("delete batch failed: %s", err.Error())
		return 0, thingmodelV1.ErrorInternalServerError("delete batch failed")
	}
	return n, nil
}

// Reorder 在事务内批量更新 sort_order。
func (r *CategoryDefaultFeatureRepo) Reorder(ctx context.Context, items []*thingmodelV1.ReorderCategoryDefaultFeaturesRequest_Item) error {
	if len(items) == 0 {
		return nil
	}
	tx, err := r.entClient.Client().Tx(ctx)
	if err != nil {
		r.log.Errorf("start transaction failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("start transaction failed")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			r.log.Errorf("transaction commit failed: %s", commitErr.Error())
			err = thingmodelV1.ErrorInternalServerError("transaction commit failed")
		}
	}()

	for _, it := range items {
		if it == nil || it.GetId() == 0 {
			continue
		}
		_, err = tx.CategoryDefaultFeature.UpdateOneID(it.GetId()).
			SetSortOrder(it.GetSortOrder()).
			SetUpdatedAt(time.Now()).
			Save(ctx)
		if err != nil {
			r.log.Errorf("reorder item id=%d failed: %s", it.GetId(), err.Error())
			return thingmodelV1.ErrorInternalServerError("reorder failed")
		}
	}
	return nil
}

// GetByCategoryFeature 取单条 (tenant, category_id, feature_id)；service 层用于 BatchAdd 冲突时复用既有行。
func (r *CategoryDefaultFeatureRepo) GetByCategoryFeature(
	ctx context.Context, tenantID, categoryID, featureID uint32,
) (*ent.CategoryDefaultFeature, error) {
	row, err := r.entClient.Client().CategoryDefaultFeature.Query().
		Where(
			categorydefaultfeature.TenantIDEQ(tenantID),
			categorydefaultfeature.CategoryIDEQ(categoryID),
			categorydefaultfeature.FeatureIDEQ(featureID),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		r.log.Errorf("get by (cat,feat) failed: %s", err.Error())
		return nil, thingmodelV1.ErrorInternalServerError("query failed")
	}
	return row, nil
}

// ToDTO 暴露 mapper.ToDTO 供 service 层使用（避免在 service 直接持有 mapper 的耦合）。
func (r *CategoryDefaultFeatureRepo) ToDTO(e *ent.CategoryDefaultFeature) *thingmodelV1.CategoryDefaultFeature {
	if e == nil {
		return nil
	}
	return r.mapper.ToDTO(e)
}

// ===== proto enum → ent enum 辅助（CR-001 后冗余列）/ Proto-to-ent enum helpers =====

func protoToEntCDFDataType(t thingmodelV1.DataType) (categorydefaultfeature.DataType, bool) {
	s := t.String()
	if s == "DATA_TYPE_UNSPECIFIED" {
		return "", false
	}
	return categorydefaultfeature.DataType(s), true
}

func protoToEntCDFAccessMode(t thingmodelV1.AccessMode) (categorydefaultfeature.AccessMode, bool) {
	s := t.String()
	if s == "ACCESS_MODE_UNSPECIFIED" {
		return "", false
	}
	return categorydefaultfeature.AccessMode(s), true
}

func protoToEntCDFEventLevel(t thingmodelV1.EventLevel) (categorydefaultfeature.EventLevel, bool) {
	s := t.String()
	if s == "EVENT_LEVEL_UNSPECIFIED" {
		return "", false
	}
	return categorydefaultfeature.EventLevel(s), true
}

func protoToEntCDFCallMode(t thingmodelV1.CallMode) (categorydefaultfeature.CallMode, bool) {
	s := t.String()
	if s == "CALL_MODE_UNSPECIFIED" {
		return "", false
	}
	return categorydefaultfeature.CallMode(s), true
}

// normalizeMaskPath 把 FieldMask path 规范化为 snake_case（接受 lowerCamel 与 snake_case 两种形态）。
func normalizeMaskPath(p string) string {
	// 快速判断：含大写字母 → 走转换；否则视为 snake_case 直接返回。
	hasUpper := false
	for i := 0; i < len(p); i++ {
		if p[i] >= 'A' && p[i] <= 'Z' {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return p
	}
	out := make([]byte, 0, len(p)+4)
	for i := 0; i < len(p); i++ {
		c := p[i]
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				out = append(out, '_')
			}
			out = append(out, c-'A'+'a')
		} else {
			out = append(out, c)
		}
	}
	return string(out)
}
