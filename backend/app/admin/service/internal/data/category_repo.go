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
	"go-wind-admin/app/admin/service/internal/data/ent/category"
	"go-wind-admin/app/admin/service/internal/data/ent/predicate"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// CategoryRepo 物模型分类仓库 / Thing-model category repository
//
// 设计依据 / Design ref: docs/thingmodel/sheji/分类管理/04-后端实现设计.md
// 镜像 unit_category_repo.go 实现风格，差异点：
//   - kind enum 字段（SYSTEM/SPACE/FACILITY，未来可扩展），需注册 EnumTypeConverter。
//   - 业务过滤（kind/level/parent_id/code 前缀/keyword）通过 PagingRequest.query 透传，
//     由 go-crud ListWithPaging 通用引擎解析，repo 内不重复实现。
//   - 新增两个领域辅助方法：GetByID（Service 校验父节点用）、HasChildren（Delete 前校验用）。
type CategoryRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper        *mapper.CopierMapper[thingmodelV1.Category, ent.Category]
	kindConverter *mapper.EnumTypeConverter[thingmodelV1.CategoryKind, category.Kind]

	repository *entCrud.Repository[
		ent.CategoryQuery, ent.CategorySelect,
		ent.CategoryCreate, ent.CategoryCreateBulk,
		ent.CategoryUpdate, ent.CategoryUpdateOne,
		ent.CategoryDelete,
		predicate.Category,
		thingmodelV1.Category, ent.Category,
	]
}

// NewCategoryRepo 构造分类仓库 / Construct category repository
func NewCategoryRepo(
	ctx *bootstrap.Context,
	entClient *entCrud.EntClient[*ent.Client],
) *CategoryRepo {
	repo := &CategoryRepo{
		log:       ctx.NewLoggerHelper("category/repo/admin-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[thingmodelV1.Category, ent.Category](),
		kindConverter: mapper.NewEnumTypeConverter[
			thingmodelV1.CategoryKind, category.Kind,
		](thingmodelV1.CategoryKind_name, thingmodelV1.CategoryKind_value),
	}

	repo.init()

	return repo
}

func (r *CategoryRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.CategoryQuery, ent.CategorySelect,
		ent.CategoryCreate, ent.CategoryCreateBulk,
		ent.CategoryUpdate, ent.CategoryUpdateOne,
		ent.CategoryDelete,
		predicate.Category,
		thingmodelV1.Category, ent.Category,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	// 注册 kind 枚举转换器 / Register kind enum converter
	r.mapper.AppendConverters(r.kindConverter.NewConverterPair())
}

// protoToEntKind 把 proto kind 转为 ent kind；UNSPECIFIED 视为未提供。
// Convert proto CategoryKind to ent kind; UNSPECIFIED returns ok=false.
func protoToEntCategoryKind(t thingmodelV1.CategoryKind) (category.Kind, bool) {
	s := t.String()
	if s == "CATEGORY_KIND_UNSPECIFIED" {
		return "", false
	}
	return category.Kind(s), true
}

// ===== CRUD =====

// Count 统计数量 / Count
func (r *CategoryRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().Category.Query()
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
//
// 业务过滤参数（kind / level / parent_id / code___starts_with / name___icontains / is_enabled）
// 由前端塞进 PagingRequest.query (JSON)，go-crud ListWithPaging 自动解析。
// 默认排序按 code 升序（前端通过 orderBy 传入）。
func (r *CategoryRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*thingmodelV1.ListCategoryResponse, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Category.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &thingmodelV1.ListCategoryResponse{Total: 0, Items: nil}, nil
	}

	return &thingmodelV1.ListCategoryResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

// IsExist 是否存在 / Is exist
func (r *CategoryRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().Category.Query().
		Where(category.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, thingmodelV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

// Get 查询详情 / Get
//
// 支持按 id 或 (kind, code) 查询。按 code 查时必须同时带 kind 才能定位唯一。
func (r *CategoryRepo) Get(ctx context.Context, req *thingmodelV1.GetCategoryRequest) (*thingmodelV1.Category, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Category.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *thingmodelV1.GetCategoryRequest_Id:
		whereCond = append(whereCond, category.IDEQ(req.GetId()))
	case *thingmodelV1.GetCategoryRequest_Code:
		builder.Where(category.CodeEQ(req.GetCode()))
		if k, ok := protoToEntCategoryKind(req.GetKind()); ok {
			builder.Where(category.KindEQ(k))
		}
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

// GetByID 取 ent 实体（Service 层用来校验父节点的 kind / level / tenant 一致性）。
// GetByID returns the raw ent entity for service-layer validation (parent kind/level/tenant checks).
func (r *CategoryRepo) GetByID(ctx context.Context, id uint32) (*ent.Category, error) {
	return r.entClient.Client().Category.Query().
		Where(category.IDEQ(id)).
		Only(ctx)
}

// HasChildren 是否存在子节点（Delete 前校验用）。
// HasChildren reports whether the category has direct children (used in delete pre-check).
func (r *CategoryRepo) HasChildren(ctx context.Context, id uint32) (bool, error) {
	n, err := r.entClient.Client().Category.Query().
		Where(category.ParentIDEQ(id)).
		Count(ctx)
	if err != nil {
		r.log.Errorf("query has children failed: %s", err.Error())
		return false, thingmodelV1.ErrorInternalServerError("query has children failed")
	}
	return n > 0, nil
}

// Create 创建分类 / Create
//
// kind / code / level 在 Create 时必须落库（Immutable，无法 Update 修改）。
func (r *CategoryRepo) Create(ctx context.Context, req *thingmodelV1.CreateCategoryRequest) (err error) {
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

	builder := tx.Category.Create().
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCode(req.Data.Code).
		SetNillableParentID(req.Data.ParentId).
		SetNillableName(req.Data.Name).
		SetNillableNameEn(req.Data.NameEn).
		SetNillableIcon(req.Data.Icon).
		SetNillableDescription(req.Data.Description).
		SetNillableSortOrder(req.Data.SortOrder).
		SetNillableIsEnabled(req.Data.IsEnabled).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	// kind / level 必填（Service 层已校验）
	if k, ok := protoToEntCategoryKind(req.Data.GetKind()); ok {
		builder.SetKind(k)
	}
	if req.Data.Level != nil {
		builder.SetLevel(uint8(req.Data.GetLevel()))
	}

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if _, err = builder.Save(ctx); err != nil {
		r.log.Errorf("insert category failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("insert category failed")
	}

	return nil
}

// Update 更新分类 / Update
//
// kind / code / parent_id / level 是不可变字段；Service 层会拒绝带这些字段的 update_mask。
// 这里仅维护可变字段（name / nameEn / icon / description / sortOrder / isEnabled / updatedBy）。
func (r *CategoryRepo) Update(ctx context.Context, req *thingmodelV1.UpdateCategoryRequest) (err error) {
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
			createReq := &thingmodelV1.CreateCategoryRequest{Data: req.Data}
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

	// kind / code / parent_id / level 为 Immutable，更新时不设置
	builder := tx.Category.UpdateOneID(req.GetId())
	_, err = r.repository.UpdateOne(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *thingmodelV1.Category) {
			builder.
				SetNillableName(req.Data.Name).
				SetNillableNameEn(req.Data.NameEn).
				SetNillableIcon(req.Data.Icon).
				SetNillableDescription(req.Data.Description).
				SetNillableSortOrder(req.Data.SortOrder).
				SetNillableIsEnabled(req.Data.IsEnabled).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(category.FieldID, req.GetId()))
		},
	)
	if err != nil {
		r.log.Errorf("update category failed: %s", err.Error())
		return thingmodelV1.ErrorInternalServerError("update category failed")
	}

	return err
}

// Delete 删除单个 / Delete one
func (r *CategoryRepo) Delete(ctx context.Context, id uint32) error {
	if id == 0 {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	if err := r.entClient.Client().Category.DeleteOneID(id).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return thingmodelV1.ErrorCategoryNotFound("category not found")
		}

		r.log.Errorf("delete one data failed: %s", err.Error())

		return thingmodelV1.ErrorInternalServerError("delete failed")
	}

	return nil
}

// BatchDelete 批量删除 / Batch delete
//
// 注意：DB 层 (parent_id) 上是 ON DELETE RESTRICT，因此如果带子节点的父行被删除会失败。
// Service 层会在调用前先用 HasChildren 校验并返回 CATEGORY_HAS_CHILDREN 错误码。
func (r *CategoryRepo) BatchDelete(ctx context.Context, ids []uint32) error {
	if len(ids) == 0 {
		return thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	if _, err := r.entClient.Client().Category.Delete().
		Where(category.IDIn(ids...)).
		Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return thingmodelV1.ErrorCategoryNotFound("category not found")
		}

		r.log.Errorf("delete data failed: %s", err.Error())

		return thingmodelV1.ErrorInternalServerError("delete failed")
	}

	return nil
}
