package service

import (
	"context"
	"fmt"
	"math"

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

// UnitService 单位服务 / Unit service
type UnitService struct {
	adminV1.UnitServiceHTTPServer

	log  *log.Helper
	repo *data.UnitRepo
}

// NewUnitService 构造单位服务
func NewUnitService(
	ctx *bootstrap.Context,
	repo *data.UnitRepo,
) *UnitService {
	return &UnitService{
		log:  ctx.NewLoggerHelper("unit/service/admin-service"),
		repo: repo,
	}
}

// List 分页查询 / List
func (s *UnitService) List(ctx context.Context, req *paginationV1.PagingRequest) (*thingmodelV1.ListUnitResponse, error) {
	return s.repo.List(ctx, req)
}

// Get 查询详情 / Get
func (s *UnitService) Get(ctx context.Context, req *thingmodelV1.GetUnitRequest) (*thingmodelV1.Unit, error) {
	return s.repo.Get(ctx, req)
}

// ListByCategory 按物理量分类查询 / List by category
func (s *UnitService) ListByCategory(ctx context.Context, req *thingmodelV1.ListUnitByCategoryRequest) (*thingmodelV1.ListUnitResponse, error) {
	return s.repo.ListByCategory(ctx, req)
}

// Create 创建 / Create
func (s *UnitService) Create(ctx context.Context, req *thingmodelV1.CreateUnitRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	if err := validateUnit(req.Data); err != nil {
		return nil, err
	}

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
func (s *UnitService) Update(ctx context.Context, req *thingmodelV1.UpdateUnitRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, thingmodelV1.ErrorBadRequest("invalid parameter")
	}
	if err := validateUnit(req.Data); err != nil {
		return nil, err
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
// 被引用（reference_count > 0）的单位拒绝物理删除，请改为停用（is_enabled=false）。
// Units with reference_count > 0 cannot be physically deleted — disable them instead.
// 本期 reference_count 由预留字段恒为 0，待 thing_property 模块落地后由其维护增减。
// reference_count is currently reserved (always 0); will be maintained by the thing_property module.
func (s *UnitService) Delete(ctx context.Context, req *thingmodelV1.DeleteUnitRequest) (*emptypb.Empty, error) {
	ids := req.GetIds()
	if len(ids) == 0 {
		return nil, thingmodelV1.ErrorBadRequest("ids is required")
	}

	// 引用守卫 / Reference guard
	referenced, err := s.repo.ReferencedIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	if len(referenced) > 0 {
		return nil, thingmodelV1.ErrorUnitInUseCannotDelete(
			"%s", fmt.Sprintf("unit(s) referenced by properties, disable instead: %v", referenced))
	}

	if err := s.repo.BatchDelete(ctx, ids); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// ===== 单位换算 / Unit conversion =====

// Convert 单位换算：源单位 → 基准单位 → 目标单位
// Convert: source unit → base unit → target unit
func (s *UnitService) Convert(ctx context.Context, req *thingmodelV1.ConvertUnitRequest) (*thingmodelV1.ConvertUnitResponse, error) {
	src, err := s.getUnitForConvert(ctx, req.GetSourceUnitId(), req.GetSourceUnitCode())
	if err != nil || src == nil {
		return convFail(thingmodelV1.ConvertUnitStatus_CONVERT_NOT_FOUND, "source unit not found"), nil
	}
	dst, err := s.getUnitForConvert(ctx, req.GetTargetUnitId(), req.GetTargetUnitCode())
	if err != nil || dst == nil {
		return convFail(thingmodelV1.ConvertUnitStatus_CONVERT_NOT_FOUND, "target unit not found"), nil
	}

	precisionOverride := -1
	if req.Precision != nil {
		precisionOverride = int(req.GetPrecision())
	}
	return convertUnits(src, dst, req.GetValue(), precisionOverride)
}

// convertUnits 单位换算纯函数实现，便于单元测试（不依赖 DB/Service）。
// convertUnits is the pure conversion core for unit testing (no DB/Service deps).
// precisionOverride < 0 表示使用目标单位 precision；否则使用该值。
// precisionOverride < 0 means "use target unit precision"; otherwise use the override.
func convertUnits(src, dst *thingmodelV1.Unit, value float64, precisionOverride int) (*thingmodelV1.ConvertUnitResponse, error) {
	// 同物理量分类校验 / Same category check
	if src.GetCategoryId() != dst.GetCategoryId() {
		return convFail(thingmodelV1.ConvertUnitStatus_CONVERT_DIFFERENT_CATEGORY,
			fmt.Sprintf("%s 与 %s 不属于同一物理量分类 / different category", src.GetCode(), dst.GetCode())), nil
	}

	// 可换算校验（仅 LINEAR / AFFINE）/ Convertible check
	if !isConvertible(src.GetConversionType()) || !isConvertible(dst.GetConversionType()) {
		return convFail(thingmodelV1.ConvertUnitStatus_CONVERT_NOT_CONVERTIBLE,
			notConvertibleReason(src, dst)), nil
	}

	srcFactor, srcOffset := coef(src)
	dstFactor, dstOffset := coef(dst)
	if srcFactor == 0 || dstFactor == 0 {
		return nil, thingmodelV1.ErrorUnitInvalidFactor("factor must not be zero")
	}

	// base = value·srcFactor + srcOffset
	base := value*srcFactor + srcOffset
	// result = (base - dstOffset) / dstFactor
	result := (base - dstOffset) / dstFactor

	if math.IsNaN(result) || math.IsInf(result, 0) {
		return nil, thingmodelV1.ErrorUnitOverflow("convert result overflow")
	}

	precision := precisionOverride
	if precision < 0 {
		precision = int(dst.GetPrecision())
	}
	result = roundTo(result, precision)

	formula := fmt.Sprintf("%v %s = %v %s", value, src.GetSymbol(), result, dst.GetSymbol())

	return &thingmodelV1.ConvertUnitResponse{
		Result:    result,
		Formula:   formula,
		Status:    thingmodelV1.ConvertUnitStatus_CONVERT_OK,
		BaseValue: trans.Ptr(roundTo(base, precision)),
	}, nil
}

// getUnitForConvert 按 id 或 code 取单位 / Get unit by id or code
func (s *UnitService) getUnitForConvert(ctx context.Context, id uint32, code string) (*thingmodelV1.Unit, error) {
	getReq := &thingmodelV1.GetUnitRequest{}
	if id != 0 {
		getReq.QueryBy = &thingmodelV1.GetUnitRequest_Id{Id: id}
	} else if code != "" {
		getReq.QueryBy = &thingmodelV1.GetUnitRequest_Code{Code: code}
	} else {
		return nil, thingmodelV1.ErrorBadRequest("unit id or code required")
	}
	return s.repo.Get(ctx, getReq)
}

// ===== 校验与辅助 / Validation & helpers =====

// validateUnit 单位写入一致性校验 / Unit write consistency validation
func validateUnit(u *thingmodelV1.Unit) error {
	if u.GetCode() == "" || u.GetSymbol() == "" || u.GetName() == "" {
		return thingmodelV1.ErrorBadRequest("code/symbol/name required")
	}
	if u.GetCategoryId() == 0 {
		return thingmodelV1.ErrorUnitCategoryNotFound("category_id required")
	}
	f, o := u.GetFactor(), u.GetOffset()
	if u.GetIsBase() {
		// 基准单位系数必须为 1/0（factor 为 nil 时由 DB default 补 1，视为合法）
		if (f != 0 && f != 1) || o != 0 {
			return thingmodelV1.ErrorUnitBaseFactorInvalid("base unit requires factor=1 & offset=0")
		}
	} else if u.GetConversionType() == thingmodelV1.ConversionType_LINEAR && o != 0 {
		// 线性单位偏移必须为 0
		return thingmodelV1.ErrorUnitLinearOffsetMustBeZero("linear unit offset must be 0")
	}
	return nil
}

func isConvertible(t thingmodelV1.ConversionType) bool {
	return t == thingmodelV1.ConversionType_LINEAR || t == thingmodelV1.ConversionType_AFFINE
}

// coef 取换算系数（factor 缺省补 1，避免除零）
func coef(u *thingmodelV1.Unit) (factor, offset float64) {
	factor = u.GetFactor()
	if factor == 0 {
		factor = 1
	}
	offset = u.GetOffset()
	return
}

func notConvertibleReason(src, dst *thingmodelV1.Unit) string {
	t := src.GetConversionType()
	if isConvertible(t) {
		t = dst.GetConversionType()
	}
	switch t {
	case thingmodelV1.ConversionType_LOGARITHMIC:
		return "对数单位（dB/dBm）不可线性换算，参见公式说明 / logarithmic not linearly convertible"
	case thingmodelV1.ConversionType_CONDITIONAL:
		return "条件换算单位需外部参数（如分子量 M、温度 T），不支持自动换算 / conditional needs external params"
	default:
		return "单位不可换算（无量纲/计数/未指定）/ not convertible"
	}
}

func roundTo(v float64, p int) float64 {
	if p < 0 {
		p = 0
	}
	pow := math.Pow(10, float64(p))
	return math.Round(v*pow) / pow
}

func convFail(st thingmodelV1.ConvertUnitStatus, msg string) *thingmodelV1.ConvertUnitResponse {
	return &thingmodelV1.ConvertUnitResponse{Status: st, Message: msg}
}
