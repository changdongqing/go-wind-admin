// 物模型特征种子程序：upsert + 单位引用解析 + 关系 id 回填。
// Feature seed runner: upsert + unit ref resolve + relation id back-fill.

package seed

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"go-wind-admin/app/admin/service/internal/data/ent"
	"go-wind-admin/app/admin/service/internal/data/ent/feature"
	"go-wind-admin/app/admin/service/internal/data/ent/schema"
	"go-wind-admin/app/admin/service/internal/data/ent/unit"
	appViewer "go-wind-admin/pkg/entgo/viewer"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

const sysTenantFeature uint32 = 0

// SeedThingmodelFeatures 将 AllFeatureSeeds() 写入数据库（幂等 upsert）。
// SeedThingmodelFeatures writes all feature seeds idempotently.
//
// 执行策略 / Strategy:
//  1. 第一遍：upsert 全部非 RELATION 特征（property/event/service），构建 identifier → id 索引；
//  2. 第二遍：upsert RELATION，按 source/target.identifier 解析对应 feature.id 并写入 spec；
//  3. property spec.unit.unitCode 在 upsert 前解析为 unitId（依赖单位种子已先执行）；
//
// 错误处理：
//   - 单条 upsert 失败不中止整个 seed；汇总收集后在末尾报告。
//   - 这避免"一个特征 bug 导致 240+ 条全部不入库"的级联失败。
//   - 调用方（ent_client.go）若收到 error，应记 Errorf 让用户能立即看到。
func SeedThingmodelFeatures(ctx context.Context, client *ent.Client, logger *log.Helper) error {
	if client == nil {
		return fmt.Errorf("seed: ent client is nil")
	}
	if logger == nil {
		logger = log.NewHelper(log.With(log.GetLogger(), "module", "thingmodel-feature-seed"))
	}

	// 系统视图：绕过 TenantPrivacy 过滤，允许写入 tenant_id=0 系统预置数据
	ctx = appViewer.NewSystemViewerContext(ctx)

	now := time.Now()
	all := AllFeatureSeeds()
	logger.Infof("[feature-seed] starting: %d total seeds", len(all))

	// 单位 code → id 索引（property spec.unit.unitCode 解析用）
	unitMap, err := buildUnitCodeIndex(ctx, client)
	if err != nil {
		return fmt.Errorf("build unit index: %w", err)
	}
	logger.Infof("[feature-seed] unit code index loaded: %d entries", len(unitMap))

	// 累计错误（单条失败不中止）/ Accumulated errors (continue on individual failures)
	type failure struct {
		code string
		err  error
	}
	var failures []failure

	// 第一遍：非 RELATION
	totalP, totalE, totalS := 0, 0, 0
	for _, f := range all {
		if f.FeatureType == ftRelation {
			continue
		}
		// 对 property 解析 unit.unitCode → unitId（写入 spec）
		if f.FeatureType == ftProperty {
			resolveUnitID(f.Spec, unitMap)
		}
		if err := upsertFeature(ctx, client, f, now); err != nil {
			logger.Errorf("[feature-seed] upsert %s (%s) FAILED: %v", f.Code, f.Identifier, err)
			failures = append(failures, failure{code: f.Code, err: err})
			continue
		}
		switch f.FeatureType {
		case ftProperty:
			totalP++
		case ftEvent:
			totalE++
		case ftService:
			totalS++
		}
	}

	// 第二遍：RELATION，先解析 identifier → id 再 upsert
	idIndex, err := buildFeatureIdentifierIndex(ctx, client)
	if err != nil {
		return fmt.Errorf("build feature identifier index: %w", err)
	}
	totalR := 0
	for _, f := range all {
		if f.FeatureType != ftRelation {
			continue
		}
		resolveRelationRefs(f.Spec, idIndex)
		if err := upsertFeature(ctx, client, f, now); err != nil {
			logger.Errorf("[feature-seed] upsert relation %s FAILED: %v", f.Code, err)
			failures = append(failures, failure{code: f.Code, err: err})
			continue
		}
		totalR++
	}

	logger.Infof("[feature-seed] upserted: property=%d, event=%d, service=%d, relation=%d (total=%d, failures=%d)",
		totalP, totalE, totalS, totalR, totalP+totalE+totalS+totalR, len(failures))

	if len(failures) > 0 {
		// 截断到前 5 条避免日志爆炸
		preview := failures
		if len(preview) > 5 {
			preview = preview[:5]
		}
		var msgs []string
		for _, ff := range preview {
			msgs = append(msgs, fmt.Sprintf("%s: %v", ff.code, ff.err))
		}
		return fmt.Errorf("%d feature(s) failed to seed (showing first %d): %s",
			len(failures), len(preview), strings.Join(msgs, "; "))
	}
	return nil
}

// ===========================================================================
// upsert 实现 / Upsert implementation
// ===========================================================================

// upsertFeature 按 (tenant_id=0, code) upsert 单条特征。
// reference_count 字段不属于 feature；spec/特化列每次覆盖。
func upsertFeature(ctx context.Context, client *ent.Client, f SeedFeature, now time.Time) error {
	specProto := buildFeatureSpecProto(f)

	builder := client.Feature.Create().
		SetTenantID(sysTenantFeature).
		SetFeatureType(feature.FeatureType(f.FeatureType.String())).
		SetCode(f.Code).
		SetIdentifier(f.Identifier).
		SetName(f.Name).
		SetNameEn(f.NameEn).
		SetDescription(f.Description).
		SetApplicableScope(f.ApplicableScope).
		SetSortOrder(f.SortOrder).
		SetIsEnabled(true).
		SetCreatedAt(now)

	// 特化列同步（与 service.syncSpecializedColumns 行为一致）
	syncSpecCols(builder, f)

	if specProto != nil {
		builder.SetSpec(schema.WrapFeatureSpec(specProto))
	}

	return builder.
		OnConflictColumns(feature.FieldTenantID, feature.FieldCode).
		Update(func(up *ent.FeatureUpsert) {
			up.UpdateFeatureType().
				UpdateIdentifier().
				UpdateName().
				UpdateNameEn().
				UpdateDescription().
				UpdateApplicableScope().
				UpdateSortOrder().
				UpdateIsEnabled().
				SetUpdatedAt(now)
			// 特化列与 spec 强制覆盖（保持种子数据为权威源）
			upsertSpecCols(up, f)
			if specProto != nil {
				up.SetSpec(schema.WrapFeatureSpec(specProto))
			}
		}).
		Exec(ctx)
}

// syncSpecCols 同步特化列 / Sync specialized columns onto create builder
func syncSpecCols(b *ent.FeatureCreate, f SeedFeature) {
	switch f.FeatureType {
	case ftProperty:
		if dt, ok := f.Spec["dataType"].(string); ok {
			b.SetDataType(feature.DataType(dt))
		}
		if am, ok := f.Spec["accessMode"].(string); ok {
			b.SetAccessMode(feature.AccessMode(am))
		}
	case ftEvent:
		if lv, ok := f.Spec["level"].(string); ok {
			b.SetEventLevel(feature.EventLevel(lv))
		}
	case ftService:
		if cm, ok := f.Spec["callMode"].(string); ok {
			b.SetCallMode(feature.CallMode(cm))
		}
	case ftRelation:
		if rt, ok := f.Spec["relationType"].(string); ok {
			b.SetRelationType(rt)
		}
	}
}

// upsertSpecCols 在 upsert update 分支中同步特化列
func upsertSpecCols(up *ent.FeatureUpsert, f SeedFeature) {
	switch f.FeatureType {
	case ftProperty:
		if dt, ok := f.Spec["dataType"].(string); ok {
			up.SetDataType(feature.DataType(dt))
		}
		if am, ok := f.Spec["accessMode"].(string); ok {
			up.SetAccessMode(feature.AccessMode(am))
		}
	case ftEvent:
		if lv, ok := f.Spec["level"].(string); ok {
			up.SetEventLevel(feature.EventLevel(lv))
		}
	case ftService:
		if cm, ok := f.Spec["callMode"].(string); ok {
			up.SetCallMode(feature.CallMode(cm))
		}
	case ftRelation:
		if rt, ok := f.Spec["relationType"].(string); ok {
			up.SetRelationType(rt)
		}
	}
}

// ===========================================================================
// 单位引用解析 / Unit reference resolution
// ===========================================================================

// buildUnitCodeIndex 构建 unit.code → unit.id 索引
func buildUnitCodeIndex(ctx context.Context, client *ent.Client) (map[string]uint32, error) {
	entities, err := client.Unit.Query().
		Where(unit.TenantIDEQ(sysTenantFeature)).
		Select(unit.FieldID, unit.FieldCode).
		All(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string]uint32, len(entities))
	for _, e := range entities {
		if e.Code != nil {
			m[*e.Code] = e.ID
		}
	}
	return m, nil
}

// resolveUnitID 在 property spec 的 unit/structFields/arraySpec.element/参数 unit 内
// 把 unitCode 解析为 unitId（递归）。
func resolveUnitID(spec map[string]any, idx map[string]uint32) {
	if spec == nil {
		return
	}
	// 顶层 unit
	if u, ok := spec["unit"].(map[string]any); ok {
		if code, ok2 := u["unitCode"].(string); ok2 && code != "" {
			if id, found := idx[code]; found {
				u["unitId"] = id
			}
		}
	}
	// structFields 递归
	if fields, ok := spec["structFields"].([]map[string]any); ok {
		for _, f := range fields {
			resolveUnitID(f, idx)
		}
	}
	// arraySpec.element 递归
	if as, ok := spec["arraySpec"].(map[string]any); ok {
		if el, ok2 := as["element"].(map[string]any); ok2 {
			resolveUnitID(el, idx)
		}
	}
	// 事件/服务输出输入参数（仅 ParamSpec 内部 unit，事件/服务种子主要走顶层）
	for _, k := range []string{"outputParams", "inputParams"} {
		if pl, ok := spec[k].([]map[string]any); ok {
			for _, p := range pl {
				resolveUnitID(p, idx)
			}
		}
	}
}

// ===========================================================================
// 关系 source/target 识别符 → id 回填
// ===========================================================================

// buildFeatureIdentifierIndex 构建 feature.identifier → id 索引
func buildFeatureIdentifierIndex(ctx context.Context, client *ent.Client) (map[string]uint32, error) {
	entities, err := client.Feature.Query().
		Where(feature.TenantIDEQ(sysTenantFeature)).
		Select(feature.FieldID, feature.FieldIdentifier, feature.FieldCode).
		All(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string]uint32, len(entities)*2)
	for _, e := range entities {
		if e.Identifier != nil {
			m[*e.Identifier] = e.ID
		}
		if e.Code != nil {
			m["code:"+*e.Code] = e.ID
		}
	}
	return m, nil
}

// resolveRelationRefs 在 relation spec 的 source/target 中按 identifier 回填 id。
// 同时也补 code（若空）以便调试展示。
func resolveRelationRefs(spec map[string]any, idx map[string]uint32) {
	if spec == nil {
		return
	}
	for _, k := range []string{"source", "target"} {
		ref, ok := spec[k].(map[string]any)
		if !ok {
			continue
		}
		if ref["kind"] != "feature" {
			continue
		}
		if id, hasID := ref["id"].(uint32); hasID && id > 0 {
			continue
		}
		if ident, ok2 := ref["identifier"].(string); ok2 && ident != "" {
			if id, found := idx[ident]; found {
				ref["id"] = id
			}
		} else if code, ok2 := ref["code"].(string); ok2 && code != "" {
			if id, found := idx["code:"+code]; found {
				ref["id"] = id
			}
		}
	}
}

// ===========================================================================
// 把 map spec → proto FeatureSpec（oneof）
// ===========================================================================

// buildFeatureSpecProto 把 SeedFeature.Spec (map) 构造为 proto FeatureSpec oneof。
// 注：DB 字段是强类型 *thingmodelV1.FeatureSpec；JSON 编码由 Ent 自动处理。
func buildFeatureSpecProto(f SeedFeature) *thingmodelV1.FeatureSpec {
	if f.Spec == nil {
		return nil
	}
	out := &thingmodelV1.FeatureSpec{}
	switch f.FeatureType {
	case ftProperty:
		p := buildPropertySpec(f.Spec)
		out.Spec = &thingmodelV1.FeatureSpec_Property{Property: p}
	case ftEvent:
		e := buildEventSpec(f.Spec)
		out.Spec = &thingmodelV1.FeatureSpec_Event{Event: e}
	case ftService:
		s := buildServiceSpec(f.Spec)
		out.Spec = &thingmodelV1.FeatureSpec_Service{Service: s}
	case ftRelation:
		r := buildRelationSpec(f.Spec)
		out.Spec = &thingmodelV1.FeatureSpec_Relation{Relation: r}
	default:
		return nil
	}
	return out
}

func buildPropertySpec(m map[string]any) *thingmodelV1.PropertySpec {
	p := &thingmodelV1.PropertySpec{}
	if v, ok := m["dataType"].(string); ok {
		dt := protoDataType(v)
		p.DataType = &dt
	}
	if v, ok := m["accessMode"].(string); ok {
		am := protoAccessMode(v)
		p.AccessMode = &am
	}
	if v, ok := m["category"].(string); ok {
		p.Category = strPtr(v)
	}
	if u, ok := m["unit"].(map[string]any); ok {
		p.Unit = buildUnitRef(u)
	}
	if c, ok := m["constraints"].(map[string]any); ok {
		p.Constraints = buildConstraints(c)
	}
	if items, ok := m["enumItems"].([]map[string]any); ok {
		for _, it := range items {
			p.EnumItems = append(p.EnumItems, buildEnumItem(it))
		}
	}
	if bl, ok := m["boolLabels"].(map[string]any); ok {
		p.BoolLabels = &thingmodelV1.BoolLabels{
			FalseLabel: anyStr(bl["false"]),
			TrueLabel:  anyStr(bl["true"]),
		}
	}
	if n, ok := m["textMaxLength"].(int); ok {
		p.TextMaxLength = int32Ptr(int32(n))
	}
	if fields, ok := m["structFields"].([]map[string]any); ok {
		for _, f := range fields {
			p.StructFields = append(p.StructFields, buildParamSpec(f))
		}
	}
	if as, ok := m["arraySpec"].(map[string]any); ok {
		p.ArraySpec = buildArraySpec(as)
	}
	if v, ok := m["isRated"].(bool); ok {
		p.IsRated = boolPtr(v)
	}
	return p
}

func buildEventSpec(m map[string]any) *thingmodelV1.EventSpec {
	e := &thingmodelV1.EventSpec{}
	if v, ok := m["level"].(string); ok {
		lv := protoEventLevel(v)
		e.Level = &lv
	}
	if params, ok := m["outputParams"].([]map[string]any); ok {
		for _, p := range params {
			e.OutputParams = append(e.OutputParams, buildParamSpec(p))
		}
	}
	if v, ok := m["triggerCondition"].(string); ok {
		e.TriggerCondition = strPtr(v)
	}
	if n, ok := m["severity"].(int); ok {
		e.Severity = int32Ptr(int32(n))
	}
	return e
}

func buildServiceSpec(m map[string]any) *thingmodelV1.ServiceSpec {
	s := &thingmodelV1.ServiceSpec{}
	if v, ok := m["callMode"].(string); ok {
		cm := protoCallMode(v)
		s.CallMode = &cm
	}
	if params, ok := m["inputParams"].([]map[string]any); ok {
		for _, p := range params {
			s.InputParams = append(s.InputParams, buildParamSpec(p))
		}
	}
	if params, ok := m["outputParams"].([]map[string]any); ok {
		for _, p := range params {
			s.OutputParams = append(s.OutputParams, buildParamSpec(p))
		}
	}
	if n, ok := m["timeout"].(int); ok {
		s.Timeout = int32Ptr(int32(n))
	}
	if v, ok := m["description"].(string); ok {
		s.Description = strPtr(v)
	}
	return s
}

func buildRelationSpec(m map[string]any) *thingmodelV1.RelationSpec {
	r := &thingmodelV1.RelationSpec{}
	if v, ok := m["relationType"].(string); ok {
		r.RelationType = strPtr(v)
	}
	if v, ok := m["cardinality"].(string); ok {
		r.Cardinality = strPtr(v)
	}
	if v, ok := m["directional"].(bool); ok {
		r.Directional = boolPtr(v)
	}
	if src, ok := m["source"].(map[string]any); ok {
		r.Source = buildEntityRef(src)
	}
	if tgt, ok := m["target"].(map[string]any); ok {
		r.Target = buildEntityRef(tgt)
	}
	return r
}

func buildParamSpec(m map[string]any) *thingmodelV1.ParamSpec {
	if m == nil {
		return nil
	}
	p := &thingmodelV1.ParamSpec{}
	if v, ok := m["key"].(string); ok {
		p.Key = strPtr(v)
	}
	if v, ok := m["name"].(string); ok {
		p.Name = strPtr(v)
	}
	if v, ok := m["dataType"].(string); ok {
		dt := protoDataType(v)
		p.DataType = &dt
	}
	if u, ok := m["unit"].(map[string]any); ok {
		p.Unit = buildUnitRef(u)
	}
	if c, ok := m["constraints"].(map[string]any); ok {
		p.Constraints = buildConstraints(c)
	}
	if items, ok := m["enumItems"].([]map[string]any); ok {
		for _, it := range items {
			p.EnumItems = append(p.EnumItems, buildEnumItem(it))
		}
	}
	if fields, ok := m["structFields"].([]map[string]any); ok {
		for _, f := range fields {
			p.StructFields = append(p.StructFields, buildParamSpec(f))
		}
	}
	if as, ok := m["arraySpec"].(map[string]any); ok {
		p.ArraySpec = buildArraySpec(as)
	}
	if v, ok := m["required"].(bool); ok {
		p.Required = boolPtr(v)
	}
	if v, ok := m["defaultValue"].(string); ok {
		p.DefaultValue = strPtr(v)
	}
	return p
}

func buildUnitRef(m map[string]any) *thingmodelV1.UnitRef {
	u := &thingmodelV1.UnitRef{}
	if v, ok := m["unitId"].(uint32); ok {
		u.UnitId = uint32Ptr(v)
	}
	if v, ok := m["unitCode"].(string); ok {
		u.UnitCode = strPtr(v)
	}
	if v, ok := m["unitSymbol"].(string); ok {
		u.UnitSymbol = strPtr(v)
	}
	return u
}

func buildConstraints(m map[string]any) *thingmodelV1.ValueConstraints {
	c := &thingmodelV1.ValueConstraints{}
	if v, ok := m["min"].(float64); ok {
		c.Min = float64Ptr(v)
	}
	if v, ok := m["max"].(float64); ok {
		c.Max = float64Ptr(v)
	}
	if v, ok := m["step"].(float64); ok {
		c.Step = float64Ptr(v)
	}
	if v, ok := m["defaultValue"].(string); ok {
		c.DefaultValue = strPtr(v)
	}
	return c
}

func buildEnumItem(m map[string]any) *thingmodelV1.EnumItem {
	it := &thingmodelV1.EnumItem{}
	switch v := m["value"].(type) {
	case int:
		it.Value = int32(v)
	case int32:
		it.Value = v
	case float64:
		it.Value = int32(v)
	}
	if v, ok := m["label"].(string); ok {
		it.Label = v
	}
	return it
}

func buildArraySpec(m map[string]any) *thingmodelV1.ArraySpec {
	as := &thingmodelV1.ArraySpec{}
	if v, ok := m["size"].(int); ok {
		as.Size = int32Ptr(int32(v))
	}
	if el, ok := m["element"].(map[string]any); ok {
		as.Element = buildParamSpec(el)
	}
	return as
}

func buildEntityRef(m map[string]any) *thingmodelV1.EntityRef {
	ref := &thingmodelV1.EntityRef{}
	if v, ok := m["kind"].(string); ok {
		ref.Kind = strPtr(v)
	}
	if v, ok := m["id"].(uint32); ok {
		ref.Id = uint32Ptr(v)
	}
	if v, ok := m["code"].(string); ok {
		ref.Code = strPtr(v)
	}
	if v, ok := m["identifier"].(string); ok {
		ref.Identifier = strPtr(v)
	}
	if v, ok := m["type"].(string); ok {
		ref.Type = strPtr(v)
	}
	return ref
}

// ===========================================================================
// 枚举 / 指针辅助 / Enum and pointer helpers
// ===========================================================================

func protoDataType(s string) thingmodelV1.DataType {
	if v, ok := thingmodelV1.DataType_value[s]; ok {
		return thingmodelV1.DataType(v)
	}
	return thingmodelV1.DataType_DATA_TYPE_UNSPECIFIED
}

func protoAccessMode(s string) thingmodelV1.AccessMode {
	if v, ok := thingmodelV1.AccessMode_value[s]; ok {
		return thingmodelV1.AccessMode(v)
	}
	return thingmodelV1.AccessMode_ACCESS_MODE_UNSPECIFIED
}

func protoEventLevel(s string) thingmodelV1.EventLevel {
	if v, ok := thingmodelV1.EventLevel_value[s]; ok {
		return thingmodelV1.EventLevel(v)
	}
	return thingmodelV1.EventLevel_EVENT_LEVEL_UNSPECIFIED
}

func protoCallMode(s string) thingmodelV1.CallMode {
	if v, ok := thingmodelV1.CallMode_value[s]; ok {
		return thingmodelV1.CallMode(v)
	}
	return thingmodelV1.CallMode_CALL_MODE_UNSPECIFIED
}

func strPtr(s string) *string       { return &s }
func boolPtr(b bool) *bool          { return &b }
func int32Ptr(n int32) *int32       { return &n }
func uint32Ptr(n uint32) *uint32    { return &n }
func float64Ptr(f float64) *float64 { return &f }

func anyStr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
