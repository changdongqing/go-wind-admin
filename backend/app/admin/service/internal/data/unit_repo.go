package data

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"

	"go-wind-admin/app/admin/service/internal/data/ent"
	"go-wind-admin/app/admin/service/internal/data/ent/predicate"
	"go-wind-admin/app/admin/service/internal/data/ent/unit"
	"go-wind-admin/app/admin/service/internal/data/ent/unitcategory"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// UnitRepo 单位仓库 / Unit repository
type UnitRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper    *mapper.CopierMapper[thingmodelV1.Unit, ent.Unit]
	convTypeConverter *mapper.EnumTypeConverter[thingmodelV1.ConversionType, unit.ConversionType]

	repository *entCrud.Repository[
		ent.UnitQuery, ent.UnitSelect,
		ent.UnitCreate, ent.UnitCreateBulk,
		ent.UnitUpdate, ent.UnitUpdateOne,
		ent.UnitDelete,
		predicate.Unit,
		thingmodelV1.Unit, ent.Unit,
	]
}

// NewUnitRepo 构造单位仓库 / Construct unit repository
func NewUnitRepo(
	ctx *bootstrap.Context,
	entClient *entCrud.EntClient[*ent.Client],
) *UnitRepo {
	repo := &UnitRepo{
		log:       ctx.NewLoggerHelper("unit/repo/admin-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[thingmodelV1.Unit, ent.Unit](),
		convTypeConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.ConversionType, unit.ConversionType,
		](thingmodelV1.ConversionType_name, thingmodelV1.ConversionType_value),
	}

	repo.init()

	return repo
}

func (r *UnitRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.UnitQuery, ent.UnitSelect,
		ent.UnitCreate, ent.UnitCreateBulk,
		ent.UnitUpdate, ent.UnitUpdateOne,
		ent.UnitDelete,
		predicate.Unit,
		thingmodelV1.Unit, ent.Unit,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	// 注册换算类型枚举转换器 / Register conversion_type enum converter
	r.mapper.AppendConverters(r.convTypeConverter.NewConverterPair())
}

// protoToEntConversionType 将 proto 换算类型转为 ent 换算类型
// Convert proto ConversionType to ent ConversionType; returns ok=false for UNSPECIFIED
func protoToEntConversionType(t thingmodelV1.ConversionType) (unit.ConversionType, bool) {
	s := t.String()
	if s == "CONVERSION_TYPE_UNSPECIFIED" {
		return "", false
	}
	return unit.ConversionType(s), true
}

// Count 统计数量 / Count
func (r *UnitRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().Unit.Query()
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
func (r *UnitRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*thingmodelV1.ListUnitResponse, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Unit.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &thingmodelV1.ListUnitResponse{Total: 0, Items: nil}, nil
	}

	return &thingmodelV1.ListUnitResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

// ListByCategory 按物理量分类查询单位（不分页，供属性选单位下拉框用）
// List units by category (no paging, for property unit selector)
func (r *UnitRepo) ListByCategory(ctx context.Context, req *thingmodelV1.ListUnitByCategoryRequest) (*thingmodelV1.ListUnitResponse, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Unit.Query()

	switch req.CategoryBy.(type) {
	case *thingmodelV1.ListUnitByCategoryRequest_CategoryId:
		builder.Where(unit.CategoryIDEQ(req.GetCategoryId()))
	case *thingmodelV1.ListUnitByCategoryRequest_CategoryCode:
		builder.Where(unit.HasCategoryWith(unitcategory.CodeEQ(req.GetCategoryCode())))
	default:
		return nil, thingmodelV1.ErrorBadRequest("category_id or category_code required")
	}

	if req.GetOnlyEnabled() {
		builder.Where(unit.IsEnabledEQ(true))
	}

	// 按排序号、ID 升序
	builder.Order(ent.Asc(unit.FieldSortOrder), ent.Asc(unit.FieldID))

	entities, err := builder.All(ctx)
	if err != nil {
		r.log.Errorf("list units by category failed: %s", err.Error())
		return nil, thingmodelV1.ErrorInternalServerError("list units by category failed")
	}

	items := make([]*thingmodelV1.Unit, 0, len(entities))
	for _, e := range entities {
		items = append(items, r.mapper.ToDTO(e))
	}

	return &thingmodelV1.ListUnitResponse{
		Total: uint64(len(items)),
		Items: items,
	}, nil
}

// IsExist 是否存在 / Is exist
func (r *UnitRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().Unit.Query().
		Where(unit.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, thingmodelV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

// Get 查询详情 / Get
func (r *UnitRepo) Get(ctx context.Context, req *thingmodelV1.GetUnitRequest) (*thingmodelV1.Unit, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Unit.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *thingmodelV1.GetUnitRequest_Id:
		whereCond = append(whereCond, unit.IDEQ(req.GetId()))
	case *thingmodelV1.GetUnitRequest_Code:
		builder.Where(unit.CodeEQ(req.GetCode()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

// Create 创建 / Create
func (r *UnitRepo) Create(ctx context.Context, req *thingmodelV1.CreateUnitRequest) (err error) {
	if req == nil || req.Data == nil {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	if req.Data.GetCategoryId() == 0 {
		return thingmodelV1.ErrorUnitCategoryNotFound("category_id required")
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

	builder := tx.Unit.Create().
		SetCategoryID(req.Data.GetCategoryId()).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCode(req.Data.Code).
		SetNillableSymbol(req.Data.Symbol).
		SetNillableName(req.Data.Name).
		SetNillableNameEn(req.Data.NameEn).
		SetNillableIsBase(req.Data.IsBase).
		SetNillableFactor(req.Data.Factor).
		SetNillableOffset(req.Data.Offset).
		SetNillableFormulaExpr(req.Data.FormulaExpr).
		SetNillablePrecision(req.Data.Precision).
		SetNillableIsSiUnit(req.Data.IsSiUnit).
		SetNillableIsLegalUnit(req.Data.IsLegalUnit).
		SetNillableReferenceCount(req.Data.ReferenceCount).
		SetNillableSortOrder(req.Data.SortOrder).
		SetNillableIsEnabled(req.Data.IsEnabled).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	// 换算类型：proto enum → ent enum（排除 UNSPECIFIED）
	if req.Data.ConversionType != nil {
		if ct, ok := protoToEntConversionType(req.Data.GetConversionType()); ok {
			builder.SetConversionType(ct)
		}
	}

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if _, err = builder.Save(ctx); err != nil {
		r.log.Errorf("insert unit failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("insert unit failed")
	}

	return nil
}

// Update 更新 / Update
func (r *UnitRepo) Update(ctx context.Context, req *thingmodelV1.UpdateUnitRequest) (err error) {
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
			createReq := &thingmodelV1.CreateUnitRequest{Data: req.Data}
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

	// 基准单位切换：当本次 Update 把 is_base 改为 true 时，先把同 (tenant_id, category_id) 下
	// 其它 is_base=true 的单位置为 false，避免与 partial unique index 冲突。
	// Base unit switch: when this update flips is_base to true, demote any other base unit
	// in the same (tenant_id, category_id) within the same tx to satisfy the partial unique index.
	if maskContains(req.GetUpdateMask(), "is_base") && req.Data.GetIsBase() {
		var (
			categoryID uint32
			tenantID   uint32
		)
		// 优先用请求中的 category_id / tenant_id；否则回查 DB 取当前行的值
		if req.Data.CategoryId != nil {
			categoryID = req.Data.GetCategoryId()
		}
		if req.Data.TenantId != nil {
			tenantID = req.Data.GetTenantId()
		}
		if categoryID == 0 || req.Data.TenantId == nil {
			cur, queryErr := tx.Unit.Query().
				Where(unit.IDEQ(req.GetId())).
				Select(unit.FieldCategoryID, unit.FieldTenantID).
				First(ctx)
			if queryErr != nil {
				r.log.Errorf("query current unit failed: %s", queryErr.Error())
				err = thingmodelV1.ErrorInternalServerError("query current unit failed")
				return err
			}
			if categoryID == 0 {
				if cur.CategoryID != nil {
					categoryID = *cur.CategoryID
				}
			}
			if req.Data.TenantId == nil {
				if cur.TenantID != nil {
					tenantID = *cur.TenantID
				}
			}
		}
		if _, demoteErr := tx.Unit.Update().
			Where(
				unit.IDNEQ(req.GetId()),
				unit.CategoryIDEQ(categoryID),
				unit.TenantIDEQ(tenantID),
				unit.IsBaseEQ(true),
			).
			SetIsBase(false).
			Save(ctx); demoteErr != nil {
			r.log.Errorf("demote previous base unit failed: %s", demoteErr.Error())
			err = thingmodelV1.ErrorInternalServerError("demote previous base unit failed")
			return err
		}
	}

	// code 为 Immutable，更新时不设置
	builder := tx.Unit.UpdateOneID(req.GetId())
	_, err = r.repository.UpdateOne(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *thingmodelV1.Unit) {
			b := builder.
				SetNillableSymbol(req.Data.Symbol).
				SetNillableName(req.Data.Name).
				SetNillableNameEn(req.Data.NameEn).
				SetNillableIsBase(req.Data.IsBase).
				SetNillableFactor(req.Data.Factor).
				SetNillableOffset(req.Data.Offset).
				SetNillableFormulaExpr(req.Data.FormulaExpr).
				SetNillablePrecision(req.Data.Precision).
				SetNillableIsSiUnit(req.Data.IsSiUnit).
				SetNillableIsLegalUnit(req.Data.IsLegalUnit).
				SetNillableSortOrder(req.Data.SortOrder).
				SetNillableIsEnabled(req.Data.IsEnabled).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())

			// category_id（Required 字段，仅当显式提供时更新）
			if req.Data.CategoryId != nil {
				b.SetCategoryID(req.Data.GetCategoryId())
			}
			// 换算类型：proto enum → ent enum
			if req.Data.ConversionType != nil {
				if ct, ok := protoToEntConversionType(req.Data.GetConversionType()); ok {
					b.SetConversionType(ct)
				}
			}
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(unit.FieldID, req.GetId()))
		},
	)
	if err != nil {
		r.log.Errorf("update unit failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("update unit failed")
	}

	return err
}

// Delete 删除单个 / Delete one
func (r *UnitRepo) Delete(ctx context.Context, id uint32) error {
	if id == 0 {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	if err := r.entClient.Client().Unit.DeleteOneID(id).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return thingmodelV1.ErrorNotFound("unit not found")
		}

		r.log.Errorf("delete one data failed: %s", err.Error())

		return thingmodelV1.ErrorInternalServerError("delete failed")
	}

	return nil
}

// BatchDelete 批量删除 / Batch delete
func (r *UnitRepo) BatchDelete(ctx context.Context, ids []uint32) error {
	if len(ids) == 0 {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	if _, err := r.entClient.Client().Unit.Delete().
		Where(unit.IDIn(ids...)).
		Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return thingmodelV1.ErrorNotFound("unit not found")
		}

		r.log.Errorf("delete data failed: %s", err.Error())

		return thingmodelV1.ErrorInternalServerError("delete failed")
	}

	return nil
}

// ReferencedIDs 返回给定 ids 中 reference_count > 0 的单位 ID 列表，用于删除前校验。
// ReferencedIDs returns unit IDs whose reference_count > 0 among the given ids.
// 本期 reference_count 字段恒为 0（待 thing_property 模块落地后维护），方法已就绪。
// reference_count is reserved (zero) in this phase; this method is ready for the property module.
func (r *UnitRepo) ReferencedIDs(ctx context.Context, ids []uint32) ([]uint32, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	entities, err := r.entClient.Client().Unit.Query().
		Where(unit.IDIn(ids...), unit.ReferenceCountGT(0)).
		Select(unit.FieldID).
		All(ctx)
	if err != nil {
		r.log.Errorf("query referenced units failed: %s", err.Error())
		return nil, thingmodelV1.ErrorInternalServerError("query referenced units failed")
	}
	out := make([]uint32, 0, len(entities))
	for _, e := range entities {
		out = append(out, e.ID)
	}
	return out, nil
}

// maskContains 判断 FieldMask 路径列表是否包含指定字段
// maskContains reports whether the FieldMask paths include the given field.
func maskContains(fm *fieldmaskpb.FieldMask, field string) bool {
	if fm == nil {
		return false
	}
	for _, p := range fm.GetPaths() {
		if p == field {
			return true
		}
	}
	return false
}
