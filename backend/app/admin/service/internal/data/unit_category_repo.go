package data

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"

	"go-wind-admin/app/admin/service/internal/data/ent"
	"go-wind-admin/app/admin/service/internal/data/ent/predicate"
	"go-wind-admin/app/admin/service/internal/data/ent/unitcategory"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// UnitCategoryRepo 物理量分类仓库 / Unit category repository
type UnitCategoryRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[thingmodelV1.UnitCategory, ent.UnitCategory]

	repository *entCrud.Repository[
		ent.UnitCategoryQuery, ent.UnitCategorySelect,
		ent.UnitCategoryCreate, ent.UnitCategoryCreateBulk,
		ent.UnitCategoryUpdate, ent.UnitCategoryUpdateOne,
		ent.UnitCategoryDelete,
		predicate.UnitCategory,
		thingmodelV1.UnitCategory, ent.UnitCategory,
	]
}

// NewUnitCategoryRepo 构造物理量分类仓库 / Construct unit category repository
func NewUnitCategoryRepo(
	ctx *bootstrap.Context,
	entClient *entCrud.EntClient[*ent.Client],
) *UnitCategoryRepo {
	repo := &UnitCategoryRepo{
		log:       ctx.NewLoggerHelper("unit-category/repo/admin-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[thingmodelV1.UnitCategory, ent.UnitCategory](),
	}

	repo.init()

	return repo
}

func (r *UnitCategoryRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.UnitCategoryQuery, ent.UnitCategorySelect,
		ent.UnitCategoryCreate, ent.UnitCategoryCreateBulk,
		ent.UnitCategoryUpdate, ent.UnitCategoryUpdateOne,
		ent.UnitCategoryDelete,
		predicate.UnitCategory,
		thingmodelV1.UnitCategory, ent.UnitCategory,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

// Count 统计数量 / Count
func (r *UnitCategoryRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().UnitCategory.Query()
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
func (r *UnitCategoryRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*thingmodelV1.ListUnitCategoryResponse, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().UnitCategory.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &thingmodelV1.ListUnitCategoryResponse{Total: 0, Items: nil}, nil
	}

	return &thingmodelV1.ListUnitCategoryResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

// IsExist 是否存在 / Is exist
func (r *UnitCategoryRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().UnitCategory.Query().
		Where(unitcategory.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, thingmodelV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

// Get 查询详情 / Get
func (r *UnitCategoryRepo) Get(ctx context.Context, req *thingmodelV1.GetUnitCategoryRequest) (*thingmodelV1.UnitCategory, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().UnitCategory.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *thingmodelV1.GetUnitCategoryRequest_Id:
		whereCond = append(whereCond, unitcategory.IDEQ(req.GetId()))
	case *thingmodelV1.GetUnitCategoryRequest_Code:
		builder.Where(unitcategory.CodeEQ(req.GetCode()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

// Create 创建 / Create
func (r *UnitCategoryRepo) Create(ctx context.Context, req *thingmodelV1.CreateUnitCategoryRequest) (err error) {
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

	builder := tx.UnitCategory.Create().
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCode(req.Data.Code).
		SetNillableName(req.Data.Name).
		SetNillableNameEn(req.Data.NameEn).
		SetNillableQuantity(req.Data.Quantity).
		SetNillableBaseUnitSymbol(req.Data.BaseUnitSymbol).
		SetNillableIcon(req.Data.Icon).
		SetNillableDescription(req.Data.Description).
		SetNillableSortOrder(req.Data.SortOrder).
		SetNillableIsEnabled(req.Data.IsEnabled).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if _, err = builder.Save(ctx); err != nil {
		r.log.Errorf("insert unit category failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("insert unit category failed")
	}

	return nil
}

// Update 更新 / Update
func (r *UnitCategoryRepo) Update(ctx context.Context, req *thingmodelV1.UpdateUnitCategoryRequest) (err error) {
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
			createReq := &thingmodelV1.CreateUnitCategoryRequest{Data: req.Data}
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
	builder := tx.UnitCategory.UpdateOneID(req.GetId())
	_, err = r.repository.UpdateOne(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *thingmodelV1.UnitCategory) {
			builder.
				SetNillableName(req.Data.Name).
				SetNillableNameEn(req.Data.NameEn).
				SetNillableQuantity(req.Data.Quantity).
				SetNillableBaseUnitSymbol(req.Data.BaseUnitSymbol).
				SetNillableIcon(req.Data.Icon).
				SetNillableDescription(req.Data.Description).
				SetNillableSortOrder(req.Data.SortOrder).
				SetNillableIsEnabled(req.Data.IsEnabled).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(unitcategory.FieldID, req.GetId()))
		},
	)
	if err != nil {
		r.log.Errorf("update unit category failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("update unit category failed")
	}

	return err
}

// Delete 删除单个 / Delete one
func (r *UnitCategoryRepo) Delete(ctx context.Context, id uint32) error {
	if id == 0 {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	if err := r.entClient.Client().UnitCategory.DeleteOneID(id).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return thingmodelV1.ErrorNotFound("unit category not found")
		}

		r.log.Errorf("delete one data failed: %s", err.Error())

		return thingmodelV1.ErrorInternalServerError("delete failed")
	}

	return nil
}

// BatchDelete 批量删除 / Batch delete
func (r *UnitCategoryRepo) BatchDelete(ctx context.Context, ids []uint32) error {
	if len(ids) == 0 {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	if _, err := r.entClient.Client().UnitCategory.Delete().
		Where(unitcategory.IDIn(ids...)).
		Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return thingmodelV1.ErrorNotFound("unit category not found")
		}

		r.log.Errorf("delete data failed: %s", err.Error())

		return thingmodelV1.ErrorInternalServerError("delete failed")
	}

	return nil
}
