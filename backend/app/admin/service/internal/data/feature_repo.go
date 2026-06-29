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
	"go-wind-admin/app/admin/service/internal/data/ent/feature"
	"go-wind-admin/app/admin/service/internal/data/ent/predicate"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// FeatureRepo 特征仓库 / Feature repository
//
// 设计依据 / Design ref:
//   - docs/thingmodel/sheji/12-特征后端实现设计.md §2
//   - docs/thingmodel/sheji/修改记录/CR-001-结构化约束下沉到模型层.md
//
// CR-001（2026-06-29）后变更：
//   - 删除 spec 字段读写、5 个特化列读写、protojson converter；
//   - 仅承载特征骨架 CRUD；
//   - ReferencedByRelation 暂留接口，但实现改为 no-op（spec 已迁移到 CDF/PF，
//     完整性校验由 CDF/PF service 在写入时执行）。
type FeatureRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[thingmodelV1.Feature, ent.Feature]

	// FeatureType 仍保留（用于 ListByType 路径）
	featureTypeConverter *mapper.EnumTypeConverter[thingmodelV1.FeatureType, feature.FeatureType]

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

	// FeatureType 枚举转换器（其它 4 个枚举与 spec 转换器已在 CR-001 删除）
	r.mapper.AppendConverters(r.featureTypeConverter.NewConverterPair())
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

// Create 创建（仅骨架）/ Create skeleton
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
		SetNillableRecommendedUnitCategoryID(req.Data.RecommendedUnitCategoryId).
		SetNillableSemanticTag(req.Data.SemanticTag).
		SetNillableSortOrder(req.Data.SortOrder).
		SetNillableIsEnabled(req.Data.IsEnabled).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.FeatureType != nil {
		if v, ok := protoToEntFeatureType(req.Data.GetFeatureType()); ok {
			builder.SetFeatureType(v)
		}
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

// UpsertByCode 按 (tenant_id, code) 幂等 upsert（导入专用，仅骨架）。
// CR-001 后不再写 spec / 5 个特化列。
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
		SetNillableRecommendedUnitCategoryID(f.RecommendedUnitCategoryId).
		SetNillableSemanticTag(f.SemanticTag).
		SetNillableSortOrder(f.SortOrder).
		SetNillableIsEnabled(f.IsEnabled).
		SetNillableCreatedBy(f.CreatedBy).
		SetCreatedAt(time.Now())

	if f.FeatureType != nil {
		if v, ok := protoToEntFeatureType(f.GetFeatureType()); ok {
			builder.SetFeatureType(v)
		}
	}

	return builder.
		OnConflictColumns(feature.FieldTenantID, feature.FieldCode).
		Update(func(up *ent.FeatureUpsert) {
			up.UpdateIdentifier().
				UpdateName().
				UpdateNameEn().
				UpdateDescription().
				UpdateApplicableScope().
				UpdateRecommendedUnitCategoryID().
				UpdateSemanticTag().
				UpdateSortOrder().
				UpdateIsEnabled().
				SetUpdatedAt(time.Now())

			if f.FeatureType != nil {
				if v, ok := protoToEntFeatureType(f.GetFeatureType()); ok {
					up.SetFeatureType(v)
				}
			}
		}).
		Exec(ctx)
}

// Update 更新（仅骨架）/ Update skeleton
func (r *FeatureRepo) Update(ctx context.Context, req *thingmodelV1.UpdateFeatureRequest) (err error) {
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
				SetNillableRecommendedUnitCategoryID(req.Data.RecommendedUnitCategoryId).
				SetNillableSemanticTag(req.Data.SemanticTag).
				SetNillableSortOrder(req.Data.SortOrder).
				SetNillableIsEnabled(req.Data.IsEnabled).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())

			if req.Data.FeatureType != nil {
				if v, ok := protoToEntFeatureType(req.Data.GetFeatureType()); ok {
					b.SetFeatureType(v)
				}
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

// ReferencedByRelation 判断指定 feature 是否被 RELATION 引用。
//
// CR-001：thing_features 不再含 spec 列，relation 的 source/target 已迁移到 CDF.spec / PF.spec。
// 删除前完整性校验改为：检查是否有任何 CDF/PF 的 spec.relation.source/target.id 等于本 id。
// 为避免在 feature_repo 引入对 CDF/PF 表的耦合，这里返回 false（不阻断删除）。
// CDF/PF 表上有 OnDelete: Restrict 的外键引用 feature_id；若 feature 被 CDF.feature_id
// 引用，DB 层会直接拒绝删除——这是更可靠的完整性保障。
//
// TODO: 如需 relation 内部 source/target 引用的细粒度校验，可在 CDF/PF service
// 的删除流程上调用专门的 helper（不属于 feature_repo 职责）。
func (r *FeatureRepo) ReferencedByRelation(ctx context.Context, id uint32) (bool, error) {
	_ = ctx
	_ = id
	return false, nil
}
