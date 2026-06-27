package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

// Feature 物模型特征表（统一承载属性/事件/服务/关系）
// Feature is the unified thing-model feature schema covering property/event/service/relation.
//
// 设计依据 / Design ref: docs/thingmodel/sheji/09-特征数据模型设计.md §2
//
// 统一表策略 / Unified-table strategy:
//   - 公共字段（code/identifier/name/...）抽为独立列；
//   - 4 类特征差异化约束收敛到 spec JSON（FeatureSpec oneof 强类型）；
//   - 5 个特化列（data_type/access_mode/event_level/call_mode/relation_type）从 spec 提升，
//     用于列表筛选与校验前置检查。
type Feature struct {
	ent.Schema
}

func (Feature) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "thingmodel_features",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("物模型-特征表（属性/事件/服务/关系统一）/ Thing model feature"),
	}
}

// Fields of the Feature.
func (Feature) Fields() []ent.Field {
	return []ent.Field{
		// ===== 公共字段 / Common fields =====
		field.Enum("feature_type").
			Comment("特征类型 PROPERTY/EVENT/SERVICE/RELATION / Feature type").
			NamedValues(
				"Property", "PROPERTY",
				"Event", "EVENT",
				"Service", "SERVICE",
				"Relation", "RELATION",
			).
			Optional().
			Nillable(),

		// 清单编码（不可变，租户内唯一）
		field.String("code").
			Comment("清单编码，如 P-RUN-0001（不可变）/ Code, immutable").
			NotEmpty().
			Immutable().
			Optional().
			Nillable(),

		// 程序标识符（租户内唯一）
		field.String("identifier").
			Comment("程序标识符，如 powerSwitch / Program identifier").
			NotEmpty().
			Optional().
			Nillable(),

		field.String("name").
			Comment("中文名 / Name (zh)").
			NotEmpty().
			Optional().
			Nillable(),

		field.String("name_en").
			Comment("英文名 / Name (en)").
			Optional().
			Nillable(),

		field.String("description").
			Comment("说明 / Description").
			Optional().
			Nillable(),

		field.String("applicable_scope").
			Comment("适用设备范围，如 冷机/锅炉 / Applicable device scope").
			Optional().
			Nillable(),

		// ===== 特化抽取列（高频筛选，从 spec 提升）/ Specialized columns =====
		field.Enum("data_type").
			Comment("property 数据类型 / Property data type").
			NamedValues(
				"Int", "INT",
				"Float", "FLOAT",
				"Double", "DOUBLE",
				"Bool", "BOOL",
				"Enum", "ENUM",
				"Text", "TEXT",
				"Date", "DATE",
				"Struct", "STRUCT",
				"Array", "ARRAY",
			).
			Optional().
			Nillable(),

		field.Enum("access_mode").
			Comment("property 访问模式 R/RW / Property access mode").
			NamedValues(
				"R", "R",
				"RW", "RW",
			).
			Optional().
			Nillable(),

		field.Enum("event_level").
			Comment("event 级别 INFO/ALERT/ERROR / Event level").
			NamedValues(
				"Info", "INFO",
				"Alert", "ALERT",
				"Error", "ERROR",
			).
			Optional().
			Nillable(),

		field.Enum("call_mode").
			Comment("service 调用模式 ASYNC/SYNC / Service call mode").
			NamedValues(
				"Async", "ASYNC",
				"Sync", "SYNC",
			).
			Optional().
			Nillable(),

		field.String("relation_type").
			Comment("relation 关系类型，如 derivedFrom/partOf / Relation type").
			Optional().
			Nillable(),

		// ===== 差异容器（四类 spec 的 JSON 序列化）/ Spec container =====
		// 注意：直接用 *thingmodelV1.FeatureSpec 会失败，因为 FeatureSpec 含 oneof，
		// encoding/json 不能识别 protobuf oneof 接口字段。
		// 必须用 FeatureSpecField 包装，内部用 protojson 序列化/反序列化。
		// 详见 featurespec_jsonfield.go 与 backend/CLAUDE.md 「Step 13」。
		field.JSON("spec", &FeatureSpecField{}).
			Comment("特征结构化约束（按 feature_type 解读，protojson 编码）/ Structured spec by feature_type (protojson)").
			Optional(),
	}
}

// Mixin of the Feature.
func (Feature) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.IsEnabled{},
		mixin.SortOrder{},
		mixin.TenantID[uint32]{},
	}
}

// Edges of the Feature.
// 反向 edges:
//   - category_default_entries ← CategoryDefaultFeature.feature
//     （一对多；某分类把该特征列为默认。删除特征时受限于条目存在，避免悬挂引用）
//
// 说明：product_features 不在此建反向 edge——它通过 ref_feature_id 弱关联，
// 应用层维护一致性（reference_count），避免 schema 循环依赖。
func (Feature) Edges() []ent.Edge {
	return []ent.Edge{
		// 多对一：property 引用单位（弱关联，spec.unit_id 指向 unit.id，不建强外键避免循环）
		// 关系：relation 自引用不建 edge（source/target 在 spec 内，应用层解析）

		// ===== 模型管理新增的反向 edge / Reverse edge for model management =====
		edge.To("category_default_entries", CategoryDefaultFeature.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Restrict,
			}),
	}
}

// Indexes of the Feature.
func (Feature) Indexes() []ent.Index {
	return []ent.Index{
		// 租户级唯一：code 与 identifier 各自租户内唯一
		index.Fields("tenant_id", "code").
			Unique().
			StorageKey("uix_thingmodel_feature_tenant_code"),
		index.Fields("tenant_id", "identifier").
			Unique().
			StorageKey("uix_thingmodel_feature_tenant_identifier"),

		// 按类型查询（最高频：左侧树选类型 → 右侧列该类型特征）
		index.Fields("tenant_id", "feature_type").
			StorageKey("idx_thingmodel_feature_tenant_type"),
		index.Fields("feature_type").
			StorageKey("idx_thingmodel_feature_type"),

		// 特化列筛选（property 按 data_type、event 按 level 等）
		index.Fields("feature_type", "data_type").
			StorageKey("idx_thingmodel_feature_type_datatype"),
		index.Fields("feature_type", "event_level").
			StorageKey("idx_thingmodel_feature_type_level"),
		index.Fields("feature_type", "call_mode").
			StorageKey("idx_thingmodel_feature_type_callmode"),

		// 适用设备范围筛选
		index.Fields("applicable_scope").
			StorageKey("idx_thingmodel_feature_scope"),

		// 常用筛选
		index.Fields("is_enabled").
			StorageKey("idx_thingmodel_feature_is_enabled"),
		index.Fields("tenant_id").
			StorageKey("idx_thingmodel_feature_tenant_id"),
		index.Fields("sort_order").
			StorageKey("idx_thingmodel_feature_sort_order"),
	}
}
