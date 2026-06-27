package data

import (
	"context"
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
	"go-wind-admin/app/admin/service/internal/data/ent/feature"
	"go-wind-admin/app/admin/service/internal/data/ent/predicate"
	"go-wind-admin/app/admin/service/internal/data/ent/schema"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// FeatureRepo 特征仓库 / Feature repository
//
// 设计依据 / Design ref: docs/thingmodel/sheji/12-特征后端实现设计.md §2
// 镜像 unit_repo.go 实现风格，差异点：
//   - 5 个 enum 字段（feature_type / data_type / access_mode / event_level / call_mode）
//   - spec 字段是 proto FeatureSpec（oneof）作为 JSON 强类型目标
//   - 提供 ListByType / ReferencedByRelation 两个领域辅助方法
type FeatureRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[thingmodelV1.Feature, ent.Feature]

	// 5 个 enum 转换器 / Five enum converters
	featureTypeConverter *mapper.EnumTypeConverter[thingmodelV1.FeatureType, feature.FeatureType]
	dataTypeConverter    *mapper.EnumTypeConverter[thingmodelV1.DataType, feature.DataType]
	accessModeConverter  *mapper.EnumTypeConverter[thingmodelV1.AccessMode, feature.AccessMode]
	eventLevelConverter  *mapper.EnumTypeConverter[thingmodelV1.EventLevel, feature.EventLevel]
	callModeConverter    *mapper.EnumTypeConverter[thingmodelV1.CallMode, feature.CallMode]

	repository *entCrud.Repository[
		ent.FeatureQuery, ent.FeatureSelect,
		ent.FeatureCreate, ent.FeatureCreateBulk,
		ent.FeatureUpdate, ent.FeatureUpdateOne,
		ent.FeatureDelete,
		predicate.Feature,
		thingmodelV1.Feature, ent.Feature,
	]
}

// NewFeatureRepo 构造特征仓库 / Construct feature repository
func NewFeatureRepo(
	ctx *bootstrap.Context,
	entClient *entCrud.EntClient[*ent.Client],
) *FeatureRepo {
	repo := &FeatureRepo{
		log:       ctx.NewLoggerHelper("feature/repo/admin-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[thingmodelV1.Feature, ent.Feature](),

		featureTypeConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.FeatureType, feature.FeatureType,
		](thingmodelV1.FeatureType_name, thingmodelV1.FeatureType_value),

		dataTypeConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.DataType, feature.DataType,
		](thingmodelV1.DataType_name, thingmodelV1.DataType_value),

		accessModeConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.AccessMode, feature.AccessMode,
		](thingmodelV1.AccessMode_name, thingmodelV1.AccessMode_value),

		eventLevelConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.EventLevel, feature.EventLevel,
		](thingmodelV1.EventLevel_name, thingmodelV1.EventLevel_value),

		callModeConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.CallMode, feature.CallMode,
		](thingmodelV1.CallMode_name, thingmodelV1.CallMode_value),
	}

	repo.init()

	return repo
}

func (r *FeatureRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.FeatureQuery, ent.FeatureSelect,
		ent.FeatureCreate, ent.FeatureCreateBulk,
		ent.FeatureUpdate, ent.FeatureUpdateOne,
		ent.FeatureDelete,
		predicate.Feature,
		thingmodelV1.Feature, ent.Feature,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	// 注册 5 个枚举转换器 / Register five enum converters
	r.mapper.AppendConverters(r.featureTypeConverter.NewConverterPair())
	r.mapper.AppendConverters(r.dataTypeConverter.NewConverterPair())
	r.mapper.AppendConverters(r.accessModeConverter.NewConverterPair())
	r.mapper.AppendConverters(r.eventLevelConverter.NewConverterPair())
	r.mapper.AppendConverters(r.callModeConverter.NewConverterPair())

	// 注册 Spec 字段的双向 protojson 转换器：
	//   entity *schema.FeatureSpecField ↔ dto *thingmodelV1.FeatureSpec
	// 否则 mapper（copier）发现两端类型不同会跳过 Spec 字段，
	// 导致 List/Get 返回的 DTO.Spec 永远为 nil（即使 DB 有数据）。
	r.mapper.AppendConverters(featureSpecConverterPair())
}

// featureSpecConverterPair 返回 Spec 字段双向类型转换对，给 CopierMapper 注册用。
// Returns a pair of copier.TypeConverter that converts between proto FeatureSpec and the ent wrapper.
func featureSpecConverterPair() []copier.TypeConverter {
	return []copier.TypeConverter{
		// entity → dto: *schema.FeatureSpecField → *thingmodelV1.FeatureSpec
		{
			SrcType: (*schema.FeatureSpecField)(nil),
			DstType: (*thingmodelV1.FeatureSpec)(nil),
			Fn: func(src interface{}) (interface{}, error) {
				f, _ := src.(*schema.FeatureSpecField)
				return schema.UnwrapFeatureSpec(f), nil
			},
		},
		// dto → entity: *thingmodelV1.FeatureSpec → *schema.FeatureSpecField
		// （目前 Create/Update 都是直接 builder.SetSpec(WrapFeatureSpec(...))，
		//  这个方向未必触发；保留以备未来 mapper.FromDTO 使用，并便于幂等。）
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

// ===== proto enum → ent enum 辅助 / Proto-to-ent enum helpers =====

// protoToEntFeatureType 将 proto 特征类型转为 ent 类型；UNSPECIFIED 视为未提供
// Convert proto FeatureType to ent; UNSPECIFIED returns ok=false.
func protoToEntFeatureType(t thingmodelV1.FeatureType) (feature.FeatureType, bool) {
	s := t.String()
	if s == "FEATURE_TYPE_UNSPECIFIED" {
		return "", false
	}
	return feature.FeatureType(s), true
}

// protoToEntDataType 将 proto 数据类型转为 ent 类型；UNSPECIFIED 视为未提供
func protoToEntDataType(t thingmodelV1.DataType) (feature.DataType, bool) {
	s := t.String()
	if s == "DATA_TYPE_UNSPECIFIED" {
		return "", false
	}
	return feature.DataType(s), true
}

// protoToEntAccessMode 将 proto 访问模式转为 ent 类型；UNSPECIFIED 视为未提供
func protoToEntAccessMode(t thingmodelV1.AccessMode) (feature.AccessMode, bool) {
	s := t.String()
	if s == "ACCESS_MODE_UNSPECIFIED" {
		return "", false
	}
	return feature.AccessMode(s), true
}

// protoToEntEventLevel 将 proto 事件级别转为 ent 类型；UNSPECIFIED 视为未提供
func protoToEntEventLevel(t thingmodelV1.EventLevel) (feature.EventLevel, bool) {
	s := t.String()
	if s == "EVENT_LEVEL_UNSPECIFIED" {
		return "", false
	}
	return feature.EventLevel(s), true
}

// protoToEntCallMode 将 proto 调用模式转为 ent 类型；UNSPECIFIED 视为未提供
func protoToEntCallMode(t thingmodelV1.CallMode) (feature.CallMode, bool) {
	s := t.String()
	if s == "CALL_MODE_UNSPECIFIED" {
		return "", false
	}
	return feature.CallMode(s), true
}

// ===== CRUD =====

// Count 统计数量 / Count
func (r *FeatureRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().Feature.Query()
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

// List 分页查询 / List with paging
func (r *FeatureRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*thingmodelV1.ListFeatureResponse, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Feature.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &thingmodelV1.ListFeatureResponse{Total: 0, Items: nil}, nil
	}

	return &thingmodelV1.ListFeatureResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

// ListByType 按特征类型查询（不分页，供左侧树联动右侧列表用）
// List features by type (no paging, for left tree → right list)
func (r *FeatureRepo) ListByType(ctx context.Context, req *thingmodelV1.ListFeatureByTypeRequest) (*thingmodelV1.ListFeatureResponse, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Feature.Query()

	if ft, ok := protoToEntFeatureType(req.GetFeatureType()); ok {
		builder.Where(feature.FeatureTypeEQ(ft))
	}
	if req.GetOnlyEnabled() {
		builder.Where(feature.IsEnabledEQ(true))
	}
	if req.GetApplicableScope() != "" {
		builder.Where(feature.ApplicableScopeEQ(req.GetApplicableScope()))
	}

	builder.Order(ent.Asc(feature.FieldSortOrder), ent.Asc(feature.FieldID))

	entities, err := builder.All(ctx)
	if err != nil {
		r.log.Errorf("list features by type failed: %s", err.Error())
		return nil, thingmodelV1.ErrorInternalServerError("list features by type failed")
	}

	items := make([]*thingmodelV1.Feature, 0, len(entities))
	for _, e := range entities {
		items = append(items, r.mapper.ToDTO(e))
	}

	return &thingmodelV1.ListFeatureResponse{
		Total: uint64(len(items)),
		Items: items,
	}, nil
}

// IsExist 是否存在 / Is exist
func (r *FeatureRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().Feature.Query().
		Where(feature.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, thingmodelV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

// Get 查询详情 / Get
func (r *FeatureRepo) Get(ctx context.Context, req *thingmodelV1.GetFeatureRequest) (*thingmodelV1.Feature, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Feature.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *thingmodelV1.GetFeatureRequest_Id:
		whereCond = append(whereCond, feature.IDEQ(req.GetId()))
	case *thingmodelV1.GetFeatureRequest_Code:
		builder.Where(feature.CodeEQ(req.GetCode()))
	case *thingmodelV1.GetFeatureRequest_Identifier:
		builder.Where(feature.IdentifierEQ(req.GetIdentifier()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

// Create 创建 / Create
func (r *FeatureRepo) Create(ctx context.Context, req *thingmodelV1.CreateFeatureRequest) (err error) {
	if req == nil || req.Data == nil {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
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

	builder := tx.Feature.Create().
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCode(req.Data.Code).
		SetNillableIdentifier(req.Data.Identifier).
		SetNillableName(req.Data.Name).
		SetNillableNameEn(req.Data.NameEn).
		SetNillableDescription(req.Data.Description).
		SetNillableApplicableScope(req.Data.ApplicableScope).
		SetNillableRelationType(req.Data.RelationType).
		SetNillableSortOrder(req.Data.SortOrder).
		SetNillableIsEnabled(req.Data.IsEnabled).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	// 枚举字段：proto → ent
	if req.Data.FeatureType != nil {
		if v, ok := protoToEntFeatureType(req.Data.GetFeatureType()); ok {
			builder.SetFeatureType(v)
		}
	}
	if req.Data.DataType != nil {
		if v, ok := protoToEntDataType(req.Data.GetDataType()); ok {
			builder.SetDataType(v)
		}
	}
	if req.Data.AccessMode != nil {
		if v, ok := protoToEntAccessMode(req.Data.GetAccessMode()); ok {
			builder.SetAccessMode(v)
		}
	}
	if req.Data.EventLevel != nil {
		if v, ok := protoToEntEventLevel(req.Data.GetEventLevel()); ok {
			builder.SetEventLevel(v)
		}
	}
	if req.Data.CallMode != nil {
		if v, ok := protoToEntCallMode(req.Data.GetCallMode()); ok {
			builder.SetCallMode(v)
		}
	}

	// spec：proto FeatureSpec 作为 JSON 强类型目标（包装为 schema.FeatureSpecField 走 protojson）
	if req.Data.Spec != nil {
		builder.SetSpec(schema.WrapFeatureSpec(req.Data.Spec))
	}

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if _, err = builder.Save(ctx); err != nil {
		r.log.Errorf("insert feature failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("insert feature failed")
	}

	return nil
}

// UpsertByCode 按 (tenant_id, code) 幂等 upsert（导入专用）。
// UpsertByCode upserts a feature idempotently by (tenant_id, code).
//
// 与 Create 的差异：code 已存在时整体覆盖（spec/特化列/公共字段），保证"导入即权威"。
// tenant_id 取自 f.TenantId（导入场景通常为当前租户；种子是 0）。
// 注意：(tenant_id, identifier) 也有唯一索引，若 identifier 与其它行冲突仍会报错——
// 调用方（service.ImportFeatures）负责在入库前保证 identifier 不重复。
func (r *FeatureRepo) UpsertByCode(ctx context.Context, f *thingmodelV1.Feature) error {
	if f == nil {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	if f.GetCode() == "" {
		return thingmodelV1.ErrorBadRequest("code is required for upsert")
	}

	builder := r.entClient.Client().Feature.Create().
		SetNillableTenantID(f.TenantId).
		SetNillableCode(f.Code).
		SetNillableIdentifier(f.Identifier).
		SetNillableName(f.Name).
		SetNillableNameEn(f.NameEn).
		SetNillableDescription(f.Description).
		SetNillableApplicableScope(f.ApplicableScope).
		SetNillableRelationType(f.RelationType).
		SetNillableSortOrder(f.SortOrder).
		SetNillableIsEnabled(f.IsEnabled).
		SetNillableCreatedBy(f.CreatedBy).
		SetCreatedAt(time.Now())

	// 枚举字段：proto → ent
	if f.FeatureType != nil {
		if v, ok := protoToEntFeatureType(f.GetFeatureType()); ok {
			builder.SetFeatureType(v)
		}
	}
	if f.DataType != nil {
		if v, ok := protoToEntDataType(f.GetDataType()); ok {
			builder.SetDataType(v)
		}
	}
	if f.AccessMode != nil {
		if v, ok := protoToEntAccessMode(f.GetAccessMode()); ok {
			builder.SetAccessMode(v)
		}
	}
	if f.EventLevel != nil {
		if v, ok := protoToEntEventLevel(f.GetEventLevel()); ok {
			builder.SetEventLevel(v)
		}
	}
	if f.CallMode != nil {
		if v, ok := protoToEntCallMode(f.GetCallMode()); ok {
			builder.SetCallMode(v)
		}
	}

	// spec：proto FeatureSpec 作为 JSON 强类型目标
	if f.Spec != nil {
		builder.SetSpec(schema.WrapFeatureSpec(f.Spec))
	}

	// 幂等：按 (tenant_id, code) 冲突则整体覆盖（含 spec/特化列）。
	return builder.
		OnConflictColumns(feature.FieldTenantID, feature.FieldCode).
		Update(func(up *ent.FeatureUpsert) {
			up.UpdateIdentifier().
				UpdateName().
				UpdateNameEn().
				UpdateDescription().
				UpdateApplicableScope().
				UpdateSortOrder().
				UpdateIsEnabled().
				SetUpdatedAt(time.Now())

			// 特化列与 spec 强制覆盖（导入为权威源）
			if f.FeatureType != nil {
				if v, ok := protoToEntFeatureType(f.GetFeatureType()); ok {
					up.SetFeatureType(v)
				}
			}
			if f.DataType != nil {
				if v, ok := protoToEntDataType(f.GetDataType()); ok {
					up.SetDataType(v)
				}
			}
			if f.AccessMode != nil {
				if v, ok := protoToEntAccessMode(f.GetAccessMode()); ok {
					up.SetAccessMode(v)
				}
			}
			if f.EventLevel != nil {
				if v, ok := protoToEntEventLevel(f.GetEventLevel()); ok {
					up.SetEventLevel(v)
				}
			}
			if f.CallMode != nil {
				if v, ok := protoToEntCallMode(f.GetCallMode()); ok {
					up.SetCallMode(v)
				}
			}
			up.UpdateRelationType()
			if f.Spec != nil {
				up.SetSpec(schema.WrapFeatureSpec(f.Spec))
			}
		}).
		Exec(ctx)
}

// Update 更新 / Update
func (r *FeatureRepo) Update(ctx context.Context, req *thingmodelV1.UpdateFeatureRequest) (err error) {
	if req == nil || req.Data == nil {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return thingmodelV1.ErrorBadRequest("id is required")
	}

	// 如果不存在则创建 / Insert when missing
	if req.GetAllowMissing() {
		var exist bool
		exist, err = r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &thingmodelV1.CreateFeatureRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
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

	// code 为 Immutable，更新时不设置
	builder := tx.Feature.UpdateOneID(req.GetId())
	_, err = r.repository.UpdateOne(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *thingmodelV1.Feature) {
			b := builder.
				SetNillableIdentifier(req.Data.Identifier).
				SetNillableName(req.Data.Name).
				SetNillableNameEn(req.Data.NameEn).
				SetNillableDescription(req.Data.Description).
				SetNillableApplicableScope(req.Data.ApplicableScope).
				SetNillableRelationType(req.Data.RelationType).
				SetNillableSortOrder(req.Data.SortOrder).
				SetNillableIsEnabled(req.Data.IsEnabled).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())

			// 枚举字段
			if req.Data.FeatureType != nil {
				if v, ok := protoToEntFeatureType(req.Data.GetFeatureType()); ok {
					b.SetFeatureType(v)
				}
			}
			if req.Data.DataType != nil {
				if v, ok := protoToEntDataType(req.Data.GetDataType()); ok {
					b.SetDataType(v)
				} else {
					b.ClearDataType()
				}
			}
			if req.Data.AccessMode != nil {
				if v, ok := protoToEntAccessMode(req.Data.GetAccessMode()); ok {
					b.SetAccessMode(v)
				} else {
					b.ClearAccessMode()
				}
			}
			if req.Data.EventLevel != nil {
				if v, ok := protoToEntEventLevel(req.Data.GetEventLevel()); ok {
					b.SetEventLevel(v)
				} else {
					b.ClearEventLevel()
				}
			}
			if req.Data.CallMode != nil {
				if v, ok := protoToEntCallMode(req.Data.GetCallMode()); ok {
					b.SetCallMode(v)
				} else {
					b.ClearCallMode()
				}
			}

			// spec 更新
			if req.Data.Spec != nil {
				b.SetSpec(schema.WrapFeatureSpec(req.Data.Spec))
			}
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(feature.FieldID, req.GetId()))
		},
	)
	if err != nil {
		r.log.Errorf("update feature failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("update feature failed")
	}

	return err
}

// Delete 删除单个 / Delete one
func (r *FeatureRepo) Delete(ctx context.Context, id uint32) error {
	if id == 0 {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	if err := r.entClient.Client().Feature.DeleteOneID(id).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return thingmodelV1.ErrorFeatureNotFound("feature not found")
		}

		r.log.Errorf("delete one data failed: %s", err.Error())

		return thingmodelV1.ErrorInternalServerError("delete failed")
	}

	return nil
}

// BatchDelete 批量删除 / Batch delete
func (r *FeatureRepo) BatchDelete(ctx context.Context, ids []uint32) error {
	if len(ids) == 0 {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	if _, err := r.entClient.Client().Feature.Delete().
		Where(feature.IDIn(ids...)).
		Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return thingmodelV1.ErrorFeatureNotFound("feature not found")
		}

		r.log.Errorf("delete data failed: %s", err.Error())

		return thingmodelV1.ErrorInternalServerError("delete failed")
	}

	return nil
}

// ReferencedByRelation 判断指定 feature 是否被 RELATION 引用（source/target.id）
// Returns true if any RELATION feature's spec.relation.source.id or .target.id equals the given id.
//
// 用于删除前完整性校验（约束 F11）/ Used for pre-delete integrity check (constraint F11).
// 使用 PostgreSQL JSON 查询；如换 MySQL 需改 JSON_EXTRACT 语法。
func (r *FeatureRepo) ReferencedByRelation(ctx context.Context, id uint32) (bool, error) {
	if id == 0 {
		return false, nil
	}
	count, err := r.entClient.Client().Feature.Query().
		Where(feature.FeatureTypeEQ(feature.FeatureTypeRelation)).
		Modify(func(s *sql.Selector) {
			// spec 列 JSON 结构：{ "Spec": { "Relation": { "source": {"id":..}, "target":{"id":..} } } }
			// 由 proto JSON 序列化习惯，proto oneof spec.relation 序列化为 spec 字段下的 Relation 分支。
			// 这里用 PostgreSQL jsonb 路径访问 spec→Relation→source/target→id（按 proto-marshaled 形态）。
			s.Where(sql.ExprP(
				"((spec->'Relation'->'source'->>'id')::int = ? OR (spec->'Relation'->'target'->>'id')::int = ?)",
				id, id,
			))
		}).
		Count(ctx)
	if err != nil {
		r.log.Errorf("query referenced by relation failed: %s", err.Error())
		return false, thingmodelV1.ErrorInternalServerError("query referenced by relation failed")
	}
	return count > 0, nil
}
