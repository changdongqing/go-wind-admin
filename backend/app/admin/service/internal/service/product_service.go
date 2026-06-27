package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-admin/app/admin/service/internal/data"
	"go-wind-admin/app/admin/service/internal/data/ent/product"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"

	"go-wind-admin/pkg/middleware/auth"
)

// ProductService 产品管理服务 / Product admin service
//
// 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/04-后端实现设计.md §7
//
// 关键业务规则：
//   - Create 前校验 category.level == 4
//   - Update 禁止改 code/category_id（schema Immutable 兜底；mask 应不包含）
//   - Delete 前检查 reference_count == 0（Repo 已兜底）
//   - Publish: DRAFT -> PUBLISHED（幂等）
//   - Unpublish: PUBLISHED -> DRAFT（本期保留接口）
type ProductService struct {
	adminV1.ProductServiceHTTPServer

	log *log.Helper

	repo         *data.ProductRepo
	categoryRepo *data.CategoryRepo
}

// NewProductService 构造服务
func NewProductService(
	ctx *bootstrap.Context,
	repo *data.ProductRepo,
	categoryRepo *data.CategoryRepo,
) *ProductService {
	return &ProductService{
		log:          ctx.NewLoggerHelper("product/service/admin-service"),
		repo:         repo,
		categoryRepo: categoryRepo,
	}
}

// List 分页查询（支持 PagingRequest.query 过滤 category_id/status/manufacturer/keyword）
func (s *ProductService) List(ctx context.Context, req *paginationV1.PagingRequest) (*thingmodelV1.ListProductResponse, error) {
	return s.repo.List(ctx, req)
}

// Get 详情（支持 by id 或 by code）
func (s *ProductService) Get(ctx context.Context, req *thingmodelV1.GetProductRequest) (*thingmodelV1.Product, error) {
	return s.repo.Get(ctx, req)
}

// Create 创建（前置：category.level == 4）
func (s *ProductService) Create(ctx context.Context, req *thingmodelV1.CreateProductRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	if err := s.validateCategoryIsLeaf(ctx, req.Data.GetCategoryId()); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// Update 更新（FieldMask；code/category_id 在 schema 层 Immutable 兜底）
func (s *ProductService) Update(ctx context.Context, req *thingmodelV1.UpdateProductRequest) (*emptypb.Empty, error) {
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
		// 兜底拒绝试图修改不可变字段
		for _, p := range req.UpdateMask.GetPaths() {
			if p == "code" || p == "category_id" || p == "categoryId" {
				return nil, thingmodelV1.ErrorProductImmutableField("field %s is immutable", p)
			}
		}
	}
	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// Delete 批量删除（Repo 内已按 ref_count 兜底）
func (s *ProductService) Delete(ctx context.Context, req *thingmodelV1.DeleteProductRequest) (*emptypb.Empty, error) {
	if req == nil || len(req.GetIds()) == 0 {
		return &emptypb.Empty{}, nil
	}
	if _, err := s.repo.DeleteBatch(ctx, req.GetIds()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// Publish DRAFT -> PUBLISHED（幂等：已 PUBLISHED 时 no-op）
func (s *ProductService) Publish(ctx context.Context, req *thingmodelV1.PublishProductRequest) (*emptypb.Empty, error) {
	if req == nil || req.GetId() == 0 {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	row, err := s.repo.GetEntityByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if row.Status == product.StatusPublished {
		return &emptypb.Empty{}, nil // 幂等
	}
	if err := s.repo.UpdateStatus(ctx, req.GetId(), product.StatusPublished); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// Unpublish PUBLISHED -> DRAFT（本期保留；前端默认不暴露）
func (s *ProductService) Unpublish(ctx context.Context, req *thingmodelV1.UnpublishProductRequest) (*emptypb.Empty, error) {
	if req == nil || req.GetId() == 0 {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	row, err := s.repo.GetEntityByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if row.Status == product.StatusDraft {
		return &emptypb.Empty{}, nil // 幂等
	}
	if err := s.repo.UpdateStatus(ctx, req.GetId(), product.StatusDraft); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ===== 辅助 / Helpers =====

// validateCategoryIsLeaf 校验 category.level == 4
func (s *ProductService) validateCategoryIsLeaf(ctx context.Context, categoryID uint32) error {
	if categoryID == 0 {
		return thingmodelV1.ErrorBadRequest("category_id is required")
	}
	cat, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return thingmodelV1.ErrorCategoryNotFound("category not found")
	}
	level := uint8(0)
	if cat.Level != nil {
		level = *cat.Level
	}
	if level != 4 {
		return thingmodelV1.ErrorProductCategoryNotLeaf("category must be level=4 (leaf), got level=%d", level)
	}
	return nil
}
