package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go-wind-admin/app/admin/service/internal/data"
	"go-wind-admin/app/admin/service/internal/data/ent"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"

	"go-wind-admin/pkg/middleware/auth"
)

// digitsOnly 纯数字 code 校验正则 / Digits-only code regex
var digitsOnly = regexp.MustCompile(`^\d+$`)

// CategoryService 物模型分类服务 / Thing-model category service
//
// 设计依据 / Design ref: docs/thingmodel/sheji/分类管理/04-后端实现设计.md §3
//
// 12 项校验（V1–V12）：
//   V1  kind 必填且属于允许集合
//   V2  code 必须为纯数字
//   V3  level 必须 1..4
//   V4  code 长度 = level × 2
//   V5  name 必填
//   V6  level=1 不可有 parent；level>1 必须有 parent
//   V7  parent.level == self.level - 1 且 parent.kind == self.kind 且 parent.tenant_id == self.tenant_id
//   V8  code 必须以 parent.code 为严格前缀，且 len(code) = len(parent.code) + 2
//   V9  (tenant, kind, code) 唯一（DB 唯一索引兜底）
//   V10 Update 不可修改 kind/code/parent_id/level
//   V11 Delete 时 children 不存在
//   V12 Delete 时 reference_count == 0（本期恒成立，预留逻辑）
type CategoryService struct {
	adminV1.CategoryServiceHTTPServer

	log *log.Helper

	repo *data.CategoryRepo
}

// NewCategoryService 构造分类服务
func NewCategoryService(
	ctx *bootstrap.Context,
	repo *data.CategoryRepo,
) *CategoryService {
	return &CategoryService{
		log:  ctx.NewLoggerHelper("category/service/admin-service"),
		repo: repo,
	}
}

// List 分页查询 / List
func (s *CategoryService) List(ctx context.Context, req *paginationV1.PagingRequest) (*thingmodelV1.ListCategoryResponse, error) {
	return s.repo.List(ctx, req)
}

// Get 查询详情 / Get
func (s *CategoryService) Get(ctx context.Context, req *thingmodelV1.GetCategoryRequest) (*thingmodelV1.Category, error) {
	return s.repo.Get(ctx, req)
}

// Create 创建分类 / Create
func (s *CategoryService) Create(ctx context.Context, req *thingmodelV1.CreateCategoryRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.validateCreate(ctx, req.Data); err != nil {
		return nil, err
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

// Update 更新分类 / Update
//
// V10: kind / code / parent_id / level 是不可变字段；带这些字段的 update_mask 一律拒绝。
func (s *CategoryService) Update(ctx context.Context, req *thingmodelV1.UpdateCategoryRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	if hasMaskedField(req.UpdateMask, "kind", "code", "parent_id", "level") {
		return nil, thingmodelV1.ErrorCategoryImmutableField("kind/code/parent_id/level are immutable")
	}

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
//
// V11: 任何一个 id 有子节点 → 拒绝（返回 CATEGORY_HAS_CHILDREN）。
// V12: 引用计数本期恒 0，预留校验位（注释中给出未来形态）。
func (s *CategoryService) Delete(ctx context.Context, req *thingmodelV1.DeleteCategoryRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}

	for _, id := range req.GetIds() {
		if id == 0 {
			continue
		}

		hasChildren, err := s.repo.HasChildren(ctx, id)
		if err != nil {
			return nil, err
		}
		if hasChildren {
			return nil, thingmodelV1.ErrorCategoryHasChildren(
				fmt.Sprintf("category id=%d has children", id))
		}

		// V12 预留：未来若 reference_count > 0 则禁止物理删除。
		// V12 placeholder for future reference_count enforcement:
		//   entity, err := s.repo.GetByID(ctx, id)
		//   if err == nil && entity != nil && entity.ReferenceCount != nil && *entity.ReferenceCount > 0 {
		//       return nil, thingmodelV1.ErrorCategoryInUseCannotDelete(...)
		//   }
	}

	if err := s.repo.BatchDelete(ctx, req.GetIds()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// ===== 校验逻辑 / Validation =====

func (s *CategoryService) validateCreate(ctx context.Context, c *thingmodelV1.Category) error {
	// V1: kind 必填
	if c.GetKind() == thingmodelV1.CategoryKind_CATEGORY_KIND_UNSPECIFIED {
		return thingmodelV1.ErrorBadRequest("kind is required")
	}

	// V2: code 必须为纯数字
	code := c.GetCode()
	if code == "" {
		return thingmodelV1.ErrorCategoryCodeFormatInvalid("code is required")
	}
	if !digitsOnly.MatchString(code) {
		return thingmodelV1.ErrorCategoryCodeFormatInvalid("code must be all digits")
	}

	// V3: level 1..4
	lvl := c.GetLevel()
	if lvl < 1 || lvl > 4 {
		return thingmodelV1.ErrorCategoryLevelInvalid("level must be 1..4")
	}

	// V4: code 长度 = level × 2
	if uint32(len(code)) != lvl*2 {
		return thingmodelV1.ErrorCategoryCodeLengthMismatch(
			fmt.Sprintf("code length must be %d for level %d, got %d", lvl*2, lvl, len(code)))
	}

	// V5: name 必填
	if c.GetName() == "" {
		return thingmodelV1.ErrorBadRequest("name is required")
	}

	// V6 / V7 / V8: 父子关系
	if lvl == 1 {
		if c.GetParentId() != 0 {
			return thingmodelV1.ErrorCategoryParentForbidden("level=1 must not have parent")
		}
	} else {
		if c.GetParentId() == 0 {
			return thingmodelV1.ErrorCategoryParentRequired("level>1 requires parent")
		}

		parent, err := s.repo.GetByID(ctx, c.GetParentId())
		if err != nil {
			if ent.IsNotFound(err) {
				return thingmodelV1.ErrorCategoryParentNotFound("parent not found")
			}
			s.log.Errorf("load parent category failed: %v", err)
			return thingmodelV1.ErrorInternalServerError("load parent failed")
		}
		if parent == nil {
			return thingmodelV1.ErrorCategoryParentNotFound("parent not found")
		}

		// V7-a: parent.level == self.level - 1
		var parentLevel uint32
		if parent.Level != nil {
			parentLevel = uint32(*parent.Level)
		}
		if parentLevel != lvl-1 {
			return thingmodelV1.ErrorCategoryLevelParentMismatch(
				fmt.Sprintf("parent level=%d, self level=%d", parentLevel, lvl))
		}
		// V7-b: parent.kind == self.kind
		var parentKindStr string
		if parent.Kind != nil {
			parentKindStr = string(*parent.Kind)
		}
		if parentKindStr != c.GetKind().String() {
			return thingmodelV1.ErrorCategoryKindParentMismatch(
				fmt.Sprintf("parent kind=%s, self kind=%s", parentKindStr, c.GetKind()))
		}
		// V7-c: parent.tenant_id == self.tenant_id（避免跨租户挂载）
		parentTenant := uint32(0)
		if parent.TenantID != nil {
			parentTenant = *parent.TenantID
		}
		if parentTenant != c.GetTenantId() {
			return thingmodelV1.ErrorCategoryKindParentMismatch(
				fmt.Sprintf("parent tenant_id=%d, self tenant_id=%d", parentTenant, c.GetTenantId()))
		}

		// V8: code 以 parent.code 为严格前缀，且 len(code) == len(parent.code) + 2
		parentCode := ""
		if parent.Code != nil {
			parentCode = *parent.Code
		}
		if !strings.HasPrefix(code, parentCode) || len(code) != len(parentCode)+2 {
			return thingmodelV1.ErrorCategoryCodePrefixMismatch(
				fmt.Sprintf("code %s must start with parent code %s and be 2 digits longer", code, parentCode))
		}
	}

	return nil
}

// hasMaskedField 检查 FieldMask 是否命中给定字段集合。
// hasMaskedField reports whether mask paths intersect with the given fields.
func hasMaskedField(mask *fieldmaskpb.FieldMask, fields ...string) bool {
	if mask == nil {
		return false
	}
	set := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		set[f] = struct{}{}
	}
	for _, p := range mask.Paths {
		if _, ok := set[p]; ok {
			return true
		}
	}
	return false
}
