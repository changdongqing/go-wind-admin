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
	"go-wind-admin/app/admin/service/internal/data/ent/product"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// ProductRepo 产品仓库 / Product repository
//
// 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/04-后端实现设计.md §1
//
// 关键约束（由 service 层兜底校验）/ Invariants:
//   - category_id 必须 level=4（service 校验，DB 不约束）
//   - code 与 category_id 均不可变（schema Immutable + DB 唯一索引）
//   - reference_count > 0 时禁止物理删除（service 拦截）
//   - status=PUBLISHED 后特征结构冻结（由 ProductFeatureService 校验器）
type ProductRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[thingmodelV1.Product, ent.Product]

	statusConverter *mapper.EnumTypeConverter[thingmodelV1.ProductStatus, product.Status]

	repository *entCrud.Repository[
		ent.ProductQuery, ent.ProductSelect,
		ent.ProductCreate, ent.ProductCreateBulk,
		ent.ProductUpdate, ent.ProductUpdateOne,
		ent.ProductDelete,
		predicate.Product,
		thingmodelV1.Product, ent.Product,
	]
}

// NewProductRepo 构造仓库
func NewProductRepo(
	ctx *bootstrap.Context,
	entClient *entCrud.EntClient[*ent.Client],
) *ProductRepo {
	repo := &ProductRepo{
		log:       ctx.NewLoggerHelper("product/repo/admin-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[thingmodelV1.Product, ent.Product](),

		statusConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.ProductStatus, product.Status,
		](thingmodelV1.ProductStatus_name, thingmodelV1.ProductStatus_value),
	}
	repo.init()
	return repo
}

func (r *ProductRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.ProductQuery, ent.ProductSelect,
		ent.ProductCreate, ent.ProductCreateBulk,
		ent.ProductUpdate, ent.ProductUpdateOne,
		ent.ProductDelete,
		predicate.Product,
		thingmodelV1.Product, ent.Product,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
}

// protoToEntProductStatus 将 proto 状态转为 ent 类型；UNSPECIFIED 视为未提供
func protoToEntProductStatus(s thingmodelV1.ProductStatus) (product.Status, bool) {
	v := s.String()
	if v == "PRODUCT_STATUS_UNSPECIFIED" {
		return "", false
	}
	return product.Status(v), true
}

// ===== CRUD =====

// Count 统计数量
func (r *ProductRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().Product.Query()
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
func (r *ProductRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*thingmodelV1.ListProductResponse, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	builder := r.entClient.Client().Product.Query()
	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &thingmodelV1.ListProductResponse{Total: 0, Items: nil}, nil
	}
	return &thingmodelV1.ListProductResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

// IsExist 是否存在
func (r *ProductRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().Product.Query().
		Where(product.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, thingmodelV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

// Get 查询详情（支持 by id 或 by code）
func (r *ProductRepo) Get(ctx context.Context, req *thingmodelV1.GetProductRequest) (*thingmodelV1.Product, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	builder := r.entClient.Client().Product.Query()
	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *thingmodelV1.GetProductRequest_Id:
		whereCond = append(whereCond, func(s *sql.Selector) {
			s.Where(sql.EQ(product.FieldID, req.GetId()))
		})
	case *thingmodelV1.GetProductRequest_Code:
		builder.Where(product.CodeEQ(req.GetCode()))
	}
	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}
	return dto, nil
}

// GetEntityByID 取 ent 实体（service 层用，便于 reference_count 等业务校验）
func (r *ProductRepo) GetEntityByID(ctx context.Context, id uint32) (*ent.Product, error) {
	row, err := r.entClient.Client().Product.Query().
		Where(product.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, thingmodelV1.ErrorProductNotFound("product not found")
		}
		r.log.Errorf("get product entity failed: %s", err.Error())
		return nil, thingmodelV1.ErrorInternalServerError("get failed")
	}
	return row, nil
}

// Create 创建
func (r *ProductRepo) Create(ctx context.Context, req *thingmodelV1.CreateProductRequest) (err error) {
	if req == nil || req.Data == nil {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	if req.Data.GetCode() == "" {
		return thingmodelV1.ErrorBadRequest("code is required")
	}
	if req.Data.GetCategoryId() == 0 {
		return thingmodelV1.ErrorBadRequest("category_id is required")
	}

	builder := r.entClient.Client().Product.Create().
		SetNillableTenantID(req.Data.TenantId).
		SetCode(req.Data.GetCode()).
		SetNillableName(req.Data.Name).
		SetNillableNameEn(req.Data.NameEn).
		SetCategoryID(req.Data.GetCategoryId()).
		SetNillableManufacturer(req.Data.Manufacturer).
		SetNillableModelNo(req.Data.ModelNo).
		SetNillableIcon(req.Data.Icon).
		SetNillableDescription(req.Data.Description).
		SetNillableSortOrder(req.Data.SortOrder).
		SetNillableIsEnabled(req.Data.IsEnabled).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	// status 默认 DRAFT；若调用方传非 UNSPECIFIED 则按提供值
	if req.Data.Status != nil {
		if v, ok := protoToEntProductStatus(req.Data.GetStatus()); ok {
			builder.SetStatus(v)
		}
	}

	if _, err = builder.Save(ctx); err != nil {
		if ent.IsConstraintError(err) {
			// 区分 code 冲突与 (cat,name) 冲突——简化处理：靠 message 匹配 OR 直接返回 CODE_DUPLICATED
			// 实际部署中可解析 PG error code 23505 details，本期保守返回 code 冲突。
			return thingmodelV1.ErrorProductCodeDuplicated("product code or (category, name) already exists")
		}
		r.log.Errorf("insert product failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("insert product failed")
	}
	return nil
}

// Update 更新（仅允许变更 name/manufacturer/modelNo/icon/description/sortOrder/isEnabled）
// code 与 category_id 由 ent.Schema 标记 Immutable，调用方不应在 update_mask 中放入。
func (r *ProductRepo) Update(ctx context.Context, req *thingmodelV1.UpdateProductRequest) (err error) {
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
			createReq := &thingmodelV1.CreateProductRequest{Data: req.Data}
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

	builder := tx.Product.UpdateOneID(req.GetId())
	_, err = r.repository.UpdateOne(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *thingmodelV1.Product) {
			b := builder.
				SetNillableName(req.Data.Name).
				SetNillableNameEn(req.Data.NameEn).
				SetNillableManufacturer(req.Data.Manufacturer).
				SetNillableModelNo(req.Data.ModelNo).
				SetNillableIcon(req.Data.Icon).
				SetNillableDescription(req.Data.Description).
				SetNillableSortOrder(req.Data.SortOrder).
				SetNillableIsEnabled(req.Data.IsEnabled).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())

			if req.Data.Status != nil {
				if v, ok := protoToEntProductStatus(req.Data.GetStatus()); ok {
					b.SetStatus(v)
				}
			}
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(product.FieldID, req.GetId()))
		},
	)
	if err != nil {
		r.log.Errorf("update product failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("update product failed")
	}
	return err
}

// UpdateStatus 仅更新 status（Publish/Unpublish 专用）
func (r *ProductRepo) UpdateStatus(ctx context.Context, id uint32, status product.Status) error {
	if id == 0 {
		return thingmodelV1.ErrorBadRequest("id is required")
	}
	_, err := r.entClient.Client().Product.UpdateOneID(id).
		SetStatus(status).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return thingmodelV1.ErrorProductNotFound("product not found")
		}
		r.log.Errorf("update product status failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("update status failed")
	}
	return nil
}

// Delete 删除（reference_count > 0 时 service 层应已拦截；这里仍兜底）
func (r *ProductRepo) Delete(ctx context.Context, id uint32) error {
	if id == 0 {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	// 先取实体检查 reference_count
	row, err := r.GetEntityByID(ctx, id)
	if err != nil {
		return err
	}
	if row.ReferenceCount != nil && *row.ReferenceCount > 0 {
		return thingmodelV1.ErrorProductInUseCannotDelete("product is referenced by device instances")
	}
	if err := r.entClient.Client().Product.DeleteOneID(id).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return thingmodelV1.ErrorProductNotFound("product not found")
		}
		r.log.Errorf("delete product failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("delete failed")
	}
	return nil
}

// DeleteBatch 批量删除（service 层负责事务与 reference_count 检查）
func (r *ProductRepo) DeleteBatch(ctx context.Context, ids []uint32) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	// 逐条 Delete 以走 reference_count 检查
	deleted := 0
	for _, id := range ids {
		if err := r.Delete(ctx, id); err != nil {
			return deleted, err
		}
		deleted++
	}
	return deleted, nil
}

// IncReferenceCount 在事务内 +N（未来 device_instance 引用产品时使用）
func (r *ProductRepo) IncReferenceCount(ctx context.Context, tx *ent.Tx, id uint32, delta int32) error {
	if tx == nil || id == 0 || delta == 0 {
		return nil
	}
	row, err := tx.Product.Query().Where(product.IDEQ(id)).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return thingmodelV1.ErrorProductNotFound("product not found")
		}
		return thingmodelV1.ErrorInternalServerError("get failed")
	}
	cur := uint32(0)
	if row.ReferenceCount != nil {
		cur = *row.ReferenceCount
	}
	next := int64(cur) + int64(delta)
	if next < 0 {
		next = 0
	}
	_, err = tx.Product.UpdateOneID(id).
		SetReferenceCount(uint32(next)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

// ToDTO mapper 包装
func (r *ProductRepo) ToDTO(e *ent.Product) *thingmodelV1.Product {
	if e == nil {
		return nil
	}
	return r.mapper.ToDTO(e)
}
