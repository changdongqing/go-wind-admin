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
	"go-wind-admin/app/admin/service/internal/data/ent/predicate"
	"go-wind-admin/app/admin/service/internal/data/ent/productfeature"
	"go-wind-admin/app/admin/service/internal/data/ent/schema"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// ProductFeatureRepo 产品下特征条目仓库 / Product feature repository
//
// 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/04-后端实现设计.md §1
//
// 关键点：
//   - feature_snapshot 为 *thingmodelV1.FeatureSpec（oneof）走 protojson；
//   - override_spec 为 *thingmodelV1.FeatureOverrideSpec（轻量白名单）走 protojson；
//   - 6 个 enum：source / feature_type / data_type / access_mode / event_level / call_mode；
//   - BatchCreate 在事务内创建 N 条（PullFromDefault 等批量场景）；
//   - reference_count 维护由 service 层负责。
type ProductFeatureRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[thingmodelV1.ProductFeature, ent.ProductFeature]

	sourceConverter      *mapper.EnumTypeConverter[thingmodelV1.ProductFeatureSource, productfeature.Source]
	featureTypeConverter *mapper.EnumTypeConverter[thingmodelV1.FeatureType, productfeature.FeatureType]
	dataTypeConverter    *mapper.EnumTypeConverter[thingmodelV1.DataType, productfeature.DataType]
	accessModeConverter  *mapper.EnumTypeConverter[thingmodelV1.AccessMode, productfeature.AccessMode]
	eventLevelConverter  *mapper.EnumTypeConverter[thingmodelV1.EventLevel, productfeature.EventLevel]
	callModeConverter    *mapper.EnumTypeConverter[thingmodelV1.CallMode, productfeature.CallMode]

	repository *entCrud.Repository[
		ent.ProductFeatureQuery, ent.ProductFeatureSelect,
		ent.ProductFeatureCreate, ent.ProductFeatureCreateBulk,
		ent.ProductFeatureUpdate, ent.ProductFeatureUpdateOne,
		ent.ProductFeatureDelete,
		predicate.ProductFeature,
		thingmodelV1.ProductFeature, ent.ProductFeature,
	]
}

// NewProductFeatureRepo 构造仓库
func NewProductFeatureRepo(
	ctx *bootstrap.Context,
	entClient *entCrud.EntClient[*ent.Client],
) *ProductFeatureRepo {
	repo := &ProductFeatureRepo{
		log:       ctx.NewLoggerHelper("product-feature/repo/admin-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[thingmodelV1.ProductFeature, ent.ProductFeature](),

		sourceConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.ProductFeatureSource, productfeature.Source,
		](thingmodelV1.ProductFeatureSource_name, thingmodelV1.ProductFeatureSource_value),

		featureTypeConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.FeatureType, productfeature.FeatureType,
		](thingmodelV1.FeatureType_name, thingmodelV1.FeatureType_value),

		dataTypeConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.DataType, productfeature.DataType,
		](thingmodelV1.DataType_name, thingmodelV1.DataType_value),

		accessModeConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.AccessMode, productfeature.AccessMode,
		](thingmodelV1.AccessMode_name, thingmodelV1.AccessMode_value),

		eventLevelConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.EventLevel, productfeature.EventLevel,
		](thingmodelV1.EventLevel_name, thingmodelV1.EventLevel_value),

		callModeConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.CallMode, productfeature.CallMode,
		](thingmodelV1.CallMode_name, thingmodelV1.CallMode_value),
	}
	repo.init()
	return repo
}

func (r *ProductFeatureRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.ProductFeatureQuery, ent.ProductFeatureSelect,
		ent.ProductFeatureCreate, ent.ProductFeatureCreateBulk,
		ent.ProductFeatureUpdate, ent.ProductFeatureUpdateOne,
		ent.ProductFeatureDelete,
		predicate.ProductFeature,
		thingmodelV1.ProductFeature, ent.ProductFeature,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	// 6 个 enum（针对 *ENTITY → *DTO 类型对的 pointer 字段：data_type/access_mode/event_level/call_mode）
	r.mapper.AppendConverters(r.sourceConverter.NewConverterPair())
	r.mapper.AppendConverters(r.featureTypeConverter.NewConverterPair())
	r.mapper.AppendConverters(r.dataTypeConverter.NewConverterPair())
	r.mapper.AppendConverters(r.accessModeConverter.NewConverterPair())
	r.mapper.AppendConverters(r.eventLevelConverter.NewConverterPair())
	r.mapper.AppendConverters(r.callModeConverter.NewConverterPair())

	// ⚠️ 重要：ent struct 中 Source / FeatureType 是 *非指针*（required enum），
	// 但 EnumTypeConverter.NewConverterPair() 注册的是 *ENTITY → *DTO；
	// 不补这两组非指针转换器会导致 copier 跳过该字段、DTO.Source/FeatureType 始终为 nil，
	// 前端看到的就是 proto 零值 PRODUCT_FEATURE_SOURCE_UNSPECIFIED / FEATURE_TYPE_UNSPECIFIED。
	r.mapper.AppendConverters(nonPointerEnumConverter[
		productfeature.Source, thingmodelV1.ProductFeatureSource,
	](thingmodelV1.ProductFeatureSource_value))
	r.mapper.AppendConverters(nonPointerEnumConverter[
		productfeature.FeatureType, thingmodelV1.FeatureType,
	](thingmodelV1.FeatureType_value))

	// 2 个 JSON wrapper converter
	r.mapper.AppendConverters(productFeatureSnapshotConverterPair())
	r.mapper.AppendConverters(featureOverrideSpecConverterPair())
}

// nonPointerEnumConverter 注册 `非指针 ENTITY → *DTO` 单向 copier 转换器。
// 解决 EnumTypeConverter 只覆盖 pointer 字段对的盲区（required enum 字段不带 *）。
func nonPointerEnumConverter[
	ENTITY ~string, DTO ~int32,
](valueMap map[string]int32) []copier.TypeConverter {
	srcType := ENTITY("")
	var dstType *DTO
	return []copier.TypeConverter{
		{
			SrcType: srcType,
			DstType: dstType,
			Fn: func(src interface{}) (interface{}, error) {
				s, _ := src.(ENTITY)
				v, ok := valueMap[string(s)]
				if !ok {
					return (*DTO)(nil), nil
				}
				d := DTO(v)
				return &d, nil
			},
		},
	}
}

// productFeatureSnapshotConverterPair 返回 feature_snapshot 字段的双向类型转换对
func productFeatureSnapshotConverterPair() []copier.TypeConverter {
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

// ===== proto enum → ent enum 辅助 =====

func protoToEntPFSource(t thingmodelV1.ProductFeatureSource) (productfeature.Source, bool) {
	s := t.String()
	if s == "PRODUCT_FEATURE_SOURCE_UNSPECIFIED" {
		return "", false
	}
	return productfeature.Source(s), true
}

func protoToEntPFFeatureType(t thingmodelV1.FeatureType) (productfeature.FeatureType, bool) {
	s := t.String()
	if s == "FEATURE_TYPE_UNSPECIFIED" {
		return "", false
	}
	return productfeature.FeatureType(s), true
}

func protoToEntPFDataType(t thingmodelV1.DataType) (productfeature.DataType, bool) {
	s := t.String()
	if s == "DATA_TYPE_UNSPECIFIED" {
		return "", false
	}
	return productfeature.DataType(s), true
}

func protoToEntPFAccessMode(t thingmodelV1.AccessMode) (productfeature.AccessMode, bool) {
	s := t.String()
	if s == "ACCESS_MODE_UNSPECIFIED" {
		return "", false
	}
	return productfeature.AccessMode(s), true
}

func protoToEntPFEventLevel(t thingmodelV1.EventLevel) (productfeature.EventLevel, bool) {
	s := t.String()
	if s == "EVENT_LEVEL_UNSPECIFIED" {
		return "", false
	}
	return productfeature.EventLevel(s), true
}

func protoToEntPFCallMode(t thingmodelV1.CallMode) (productfeature.CallMode, bool) {
	s := t.String()
	if s == "CALL_MODE_UNSPECIFIED" {
		return "", false
	}
	return productfeature.CallMode(s), true
}

// ===== CRUD =====

func (r *ProductFeatureRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().ProductFeature.Query()
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

func (r *ProductFeatureRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*thingmodelV1.ListProductFeatureResponse, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	builder := r.entClient.Client().ProductFeature.Query()
	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &thingmodelV1.ListProductFeatureResponse{Total: 0, Items: nil}, nil
	}
	return &thingmodelV1.ListProductFeatureResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

// ListByProduct 按产品 ID 取全部条目（service 层用于 Get/Pull 后查询）
func (r *ProductFeatureRepo) ListByProduct(ctx context.Context, productID uint32) ([]*ent.ProductFeature, error) {
	if productID == 0 {
		return nil, thingmodelV1.ErrorBadRequest("product_id is required")
	}
	rows, err := r.entClient.Client().ProductFeature.Query().
		Where(productfeature.ProductIDEQ(productID)).
		Order(ent.Asc(productfeature.FieldSortOrder), ent.Asc(productfeature.FieldID)).
		All(ctx)
	if err != nil {
		r.log.Errorf("list by product failed: %s", err.Error())
		return nil, thingmodelV1.ErrorInternalServerError("list by product failed")
	}
	return rows, nil
}

// ListByProductFeatureIds 取产品下某些 (product_id, ref_feature_id) 已存在的条目（PullFromDefault 冲突检测用）。
func (r *ProductFeatureRepo) ListByProductFeatureIds(ctx context.Context, productID uint32, featureIDs []uint32) ([]*ent.ProductFeature, error) {
	if productID == 0 || len(featureIDs) == 0 {
		return nil, nil
	}
	rows, err := r.entClient.Client().ProductFeature.Query().
		Where(
			productfeature.ProductIDEQ(productID),
			productfeature.RefFeatureIDIn(featureIDs...),
		).
		All(ctx)
	if err != nil {
		r.log.Errorf("list existing product features failed: %s", err.Error())
		return nil, thingmodelV1.ErrorInternalServerError("list existing failed")
	}
	return rows, nil
}

func (r *ProductFeatureRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().ProductFeature.Query().
		Where(productfeature.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, thingmodelV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *ProductFeatureRepo) Get(ctx context.Context, req *thingmodelV1.GetProductFeatureRequest) (*thingmodelV1.ProductFeature, error) {
	if req == nil || req.GetId() == 0 {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	builder := r.entClient.Client().ProductFeature.Query()
	whereCond := []func(s *sql.Selector){
		func(s *sql.Selector) { s.Where(sql.EQ(productfeature.FieldID, req.GetId())) },
	}
	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}
	return dto, nil
}

// GetEntityByID 取 ent 实体（service 层用，便于决策 source/feature_snapshot 等业务逻辑）
func (r *ProductFeatureRepo) GetEntityByID(ctx context.Context, id uint32) (*ent.ProductFeature, error) {
	row, err := r.entClient.Client().ProductFeature.Query().
		Where(productfeature.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, thingmodelV1.ErrorProductFeatureNotFound("product feature not found")
		}
		r.log.Errorf("get product feature entity failed: %s", err.Error())
		return nil, thingmodelV1.ErrorInternalServerError("get failed")
	}
	return row, nil
}

// Create 创建一条（service 层走 GLOBAL/LOCAL 路径会调用；DEFAULT 走 BatchCreateInTx）
func (r *ProductFeatureRepo) Create(ctx context.Context, req *thingmodelV1.CreateProductFeatureRequest) (err error) {
	if req == nil || req.Data == nil {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	_, err = r.createTx(ctx, nil, req.Data, time.Now())
	return err
}

// createTx 真正的创建逻辑（可在外部事务中复用）
func (r *ProductFeatureRepo) createTx(ctx context.Context, tx *ent.Tx, data *thingmodelV1.ProductFeature, now time.Time) (*ent.ProductFeature, error) {
	var builder *ent.ProductFeatureCreate
	if tx != nil {
		builder = tx.ProductFeature.Create()
	} else {
		builder = r.entClient.Client().ProductFeature.Create()
	}

	builder.
		SetNillableTenantID(data.TenantId).
		SetProductID(data.GetProductId()).
		SetNillableRefFeatureID(data.RefFeatureId).
		SetNillableCode(data.Code).
		SetNillableIdentifier(data.Identifier).
		SetNillableName(data.Name).
		SetNillableNameEn(data.NameEn).
		SetNillableDescription(data.Description).
		SetNillableSortOrder(data.SortOrder).
		SetNillableIsEnabled(data.IsEnabled).
		SetNillableCreatedBy(data.CreatedBy).
		SetCreatedAt(now)

	if data.Source != nil {
		if v, ok := protoToEntPFSource(data.GetSource()); ok {
			builder.SetSource(v)
		}
	}
	if data.FeatureType != nil {
		if v, ok := protoToEntPFFeatureType(data.GetFeatureType()); ok {
			builder.SetFeatureType(v)
		}
	}
	if data.DataType != nil {
		if v, ok := protoToEntPFDataType(data.GetDataType()); ok {
			builder.SetDataType(v)
		}
	}
	if data.AccessMode != nil {
		if v, ok := protoToEntPFAccessMode(data.GetAccessMode()); ok {
			builder.SetAccessMode(v)
		}
	}
	if data.EventLevel != nil {
		if v, ok := protoToEntPFEventLevel(data.GetEventLevel()); ok {
			builder.SetEventLevel(v)
		}
	}
	if data.CallMode != nil {
		if v, ok := protoToEntPFCallMode(data.GetCallMode()); ok {
			builder.SetCallMode(v)
		}
	}
	if data.RelationType != nil {
		builder.SetRelationType(data.GetRelationType())
	}

	// JSON 字段（feature_snapshot 必填；override_spec 可选）
	if data.FeatureSnapshot != nil {
		builder.SetFeatureSnapshot(schema.WrapFeatureSpec(data.FeatureSnapshot))
	}
	if data.OverrideSpec != nil {
		builder.SetOverrideSpec(schema.WrapFeatureOverrideSpec(data.OverrideSpec))
	}

	row, err := builder.Save(ctx)
	if err != nil {
		r.log.Errorf("insert product_feature failed: %s", err.Error())
		if ent.IsConstraintError(err) {
			return nil, thingmodelV1.ErrorPfDuplicateCode("product feature code or identifier already exists within product")
		}
		return nil, thingmodelV1.ErrorInternalServerError("insert product_feature failed")
	}
	return row, nil
}

// CreateInTx 暴露给 service 层使用（同事务）
func (r *ProductFeatureRepo) CreateInTx(ctx context.Context, tx *ent.Tx, data *thingmodelV1.ProductFeature) (*ent.ProductFeature, error) {
	if tx == nil || data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	return r.createTx(ctx, tx, data, time.Now())
}

// BatchCreateInTx 在同一事务内批量创建 N 条（PullFromDefault 使用）。
// 任一条失败整批回滚由 service 层的事务包裹保证。
func (r *ProductFeatureRepo) BatchCreateInTx(ctx context.Context, tx *ent.Tx, items []*thingmodelV1.ProductFeature) ([]*ent.ProductFeature, error) {
	if tx == nil || len(items) == 0 {
		return nil, nil
	}
	now := time.Now()
	out := make([]*ent.ProductFeature, 0, len(items))
	for _, it := range items {
		row, err := r.createTx(ctx, tx, it, now)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, nil
}

// Update（FieldMask；PUBLISHED 冻结规则由 service 层校验器拦截）
func (r *ProductFeatureRepo) Update(ctx context.Context, req *thingmodelV1.UpdateProductFeatureRequest) (err error) {
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
			createReq := &thingmodelV1.CreateProductFeatureRequest{Data: req.Data}
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

	builder := tx.ProductFeature.UpdateOneID(req.GetId())
	_, err = r.repository.UpdateOne(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *thingmodelV1.ProductFeature) {
			b := builder.
				SetNillableName(req.Data.Name).
				SetNillableNameEn(req.Data.NameEn).
				SetNillableDescription(req.Data.Description).
				SetNillableSortOrder(req.Data.SortOrder).
				SetNillableIsEnabled(req.Data.IsEnabled).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())

			// 仅在 LOCAL 来源时允许改 code/identifier；service 层负责前置校验
			if req.Data.Code != nil {
				b.SetCode(req.Data.GetCode())
			}
			if req.Data.Identifier != nil {
				b.SetIdentifier(req.Data.GetIdentifier())
			}

			// override_spec：传入即覆盖（白名单收口）
			if req.Data.OverrideSpec != nil {
				b.SetOverrideSpec(schema.WrapFeatureOverrideSpec(req.Data.OverrideSpec))
			}

			// LOCAL 来源允许改 feature_snapshot（service 校验把控）；DEFAULT/GLOBAL 禁改
			if req.Data.FeatureSnapshot != nil {
				b.SetFeatureSnapshot(schema.WrapFeatureSpec(req.Data.FeatureSnapshot))
			}
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(productfeature.FieldID, req.GetId()))
		},
	)
	if err != nil {
		r.log.Errorf("update product_feature failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("update product_feature failed")
	}
	return err
}

func (r *ProductFeatureRepo) Delete(ctx context.Context, id uint32) error {
	if id == 0 {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	if err := r.entClient.Client().ProductFeature.DeleteOneID(id).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return thingmodelV1.ErrorProductFeatureNotFound("product feature not found")
		}
		r.log.Errorf("delete product_feature failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

// DeleteInTx 提供给 service 层在事务中删除（同事务维护 reference_count）。
func (r *ProductFeatureRepo) DeleteInTx(ctx context.Context, tx *ent.Tx, id uint32) error {
	if tx == nil || id == 0 {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	if err := tx.ProductFeature.DeleteOneID(id).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return thingmodelV1.ErrorProductFeatureNotFound("product feature not found")
		}
		r.log.Errorf("delete in tx failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

func (r *ProductFeatureRepo) DeleteBatch(ctx context.Context, ids []uint32) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	n, err := r.entClient.Client().ProductFeature.Delete().
		Where(productfeature.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		r.log.Errorf("delete batch failed: %s", err.Error())
		return 0, thingmodelV1.ErrorInternalServerError("delete batch failed")
	}
	return n, nil
}

// Reorder 批量更新 sort_order
func (r *ProductFeatureRepo) Reorder(ctx context.Context, items []*thingmodelV1.ReorderProductFeaturesRequest_Item) error {
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
		_, err = tx.ProductFeature.UpdateOneID(it.GetId()).
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

// ToDTO mapper 暴露
func (r *ProductFeatureRepo) ToDTO(e *ent.ProductFeature) *thingmodelV1.ProductFeature {
	if e == nil {
		return nil
	}
	return r.mapper.ToDTO(e)
}

// CountByProduct 取某产品下条目总数（用于产品列表的 feature_count 统计字段）
func (r *ProductFeatureRepo) CountByProduct(ctx context.Context, productID uint32) (int, error) {
	if productID == 0 {
		return 0, nil
	}
	return r.entClient.Client().ProductFeature.Query().
		Where(productfeature.ProductIDEQ(productID)).
		Count(ctx)
}
