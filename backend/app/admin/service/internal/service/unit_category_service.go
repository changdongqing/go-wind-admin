package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-admin/app/admin/service/internal/data"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"

	"go-wind-admin/pkg/middleware/auth"
)

// UnitCategoryService 物理量分类服务 / Unit category service
type UnitCategoryService struct {
	adminV1.UnitCategoryServiceHTTPServer

	log *log.Helper

	repo *data.UnitCategoryRepo
}

// NewUnitCategoryService 构造物理量分类服务
func NewUnitCategoryService(
	ctx *bootstrap.Context,
	repo *data.UnitCategoryRepo,
) *UnitCategoryService {
	return &UnitCategoryService{
		log:  ctx.NewLoggerHelper("unit-category/service/admin-service"),
		repo: repo,
	}
}

// List 分页查询 / List
func (s *UnitCategoryService) List(ctx context.Context, req *paginationV1.PagingRequest) (*thingmodelV1.ListUnitCategoryResponse, error) {
	return s.repo.List(ctx, req)
}

// Get 查询详情 / Get
func (s *UnitCategoryService) Get(ctx context.Context, req *thingmodelV1.GetUnitCategoryRequest) (*thingmodelV1.UnitCategory, error) {
	return s.repo.Get(ctx, req)
}

// Create 创建 / Create
func (s *UnitCategoryService) Create(ctx context.Context, req *thingmodelV1.CreateUnitCategoryRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	// 获取操作人信息 / Get operator
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	if err = s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Update 更新 / Update
func (s *UnitCategoryService) Update(ctx context.Context, req *thingmodelV1.UpdateUnitCategoryRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	// 获取操作人信息 / Get operator
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.Id = trans.Ptr(req.GetId())

	req.Data.UpdatedBy = trans.Ptr(operator.UserId)
	if req.UpdateMask != nil {
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "updated_by")
	}

	if err = s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Delete 批量删除 / Batch delete
func (s *UnitCategoryService) Delete(ctx context.Context, req *thingmodelV1.DeleteUnitCategoryRequest) (*emptypb.Empty, error) {
	if err := s.repo.BatchDelete(ctx, req.GetIds()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
