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

// CategoryDefaultFeature 分类默认模型条目（关联表）
// CategoryDefaultFeature is one entry in a category's default model — binding a category (level=4)
// to a global feature skeleton, plus the FULL structured spec at this category's scope.
//
// 设计依据 / Design ref:
//   - docs/thingmodel/sheji/模型管理/02-数据模型设计.md §2.1
//   - docs/thingmodel/sheji/修改记录/CR-001-结构化约束下沉到模型层.md
//
// CR-001（2026-06-29）后变更：
//   - override_spec(FeatureOverrideSpec) → spec(FeatureSpec)：从"白名单稀疏覆写"升级为"完整约束承载"；
//   - 新增 5 个冗余特化抽取列（data_type/access_mode/event_level/call_mode/relation_type），从 spec 派生，
//     用于列表筛选与排序，避免 JSON 内字段索引开销。
//
// 关键约束（应用层兜底）/ Application-layer invariants:
//   - category_id 必须指向 level=4 节点（Service 校验）
//   - spec 完整性由 V1–V17 校验器在 Create/Update 时执行
//   - 写入/删除时维护 thing_features.reference_count（来自 feature_id；本期未启用计数列，预留）
//   - 写入/删除时维护 thingmodel_units.reference_count（来自 spec.property.unit）
type CategoryDefaultFeature struct {
	ent.Schema
}

func (CategoryDefaultFeature) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "thingmodel_category_default_features",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("分类默认模型条目（CR-001：承载完整 spec）/ Category default model entry (full spec)"),
	}
}

// Fields of the CategoryDefaultFeature.
func (CategoryDefaultFeature) Fields() []ent.Field {
	return []ent.Field{
		// 分类 ID（必须 level=4）
		field.Uint32("category_id").
			Comment("分类 ID（必须 level=4）/ Category id, must be level=4"),

		// 全局特征 ID（骨架来源）
		field.Uint32("feature_id").
			Comment("全局特征 ID（骨架来源）/ Global feature id (skeleton source)"),

		// 完整 FeatureSpec（CR-001 承载所有结构化约束）
		// 用 FeatureSpecField wrapper 走 protojson 编解码。
		field.JSON("spec", &FeatureSpecField{}).
			Optional().
			Comment("完整 FeatureSpec（按 feature_type 解读，protojson）/ Full feature spec (protojson)"),

		// 分类内展示别名
		field.String("display_name").
			MaxLen(128).
			Optional().
			Nillable().
			Comment("分类内展示别名 / Display alias within category"),

		// ===== CR-001 新增：冗余特化抽取列（从 spec 派生，便于列表筛选）/ Specialized columns =====
		field.Enum("data_type").
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
			Nillable().
			Comment("property 数据类型（从 spec 派生）/ Property data type"),

		field.Enum("access_mode").
			NamedValues(
				"R", "R",
				"RW", "RW",
			).
			Optional().
			Nillable().
			Comment("property 访问模式 R/RW / Property access mode"),

		field.Enum("event_level").
			NamedValues(
				"Info", "INFO",
				"Alert", "ALERT",
				"Error", "ERROR",
			).
			Optional().
			Nillable().
			Comment("event 级别 / Event level"),

		field.Enum("call_mode").
			NamedValues(
				"Async", "ASYNC",
				"Sync", "SYNC",
			).
			Optional().
			Nillable().
			Comment("service 调用模式 / Service call mode"),

		field.String("relation_type").
			Optional().
			Nillable().
			Comment("relation 关系类型 / Relation type"),
	}
}

// Mixin of the CategoryDefaultFeature.
func (CategoryDefaultFeature) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.IsEnabled{},
		mixin.SortOrder{},
		mixin.TenantID[uint32]{},
	}
}

// Edges of the CategoryDefaultFeature.
// 双向：上游 Category 与 Feature 通过 Ref 引用本表的反向 edge（"default_features" 与 "category_default_entries"）。
func (CategoryDefaultFeature) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("category", Category.Type).
			Ref("default_features").
			Field("category_id").
			Required().
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Restrict,
			}),

		edge.From("feature", Feature.Type).
			Ref("category_default_entries").
			Field("feature_id").
			Required().
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Restrict,
			}),
	}
}

// Indexes of the CategoryDefaultFeature.
func (CategoryDefaultFeature) Indexes() []ent.Index {
	return []ent.Index{
		// 同一分类下同一特征不可重复（核心唯一约束）
		index.Fields("tenant_id", "category_id", "feature_id").
			Unique().
			StorageKey("uix_tm_cat_default_features_tenant_cat_feat"),

		// 列表主查询：按分类列条目，按 sort_order 排序
		index.Fields("tenant_id", "category_id", "sort_order").
			StorageKey("idx_tm_cat_default_features_tenant_cat_sort"),

		// 反向查询：某特征被哪些分类引用（治理用）
		index.Fields("feature_id").
			StorageKey("idx_tm_cat_default_features_feature"),

		// CR-001 新增：冗余列筛选
		index.Fields("data_type").
			StorageKey("idx_tm_cat_default_features_data_type"),
		index.Fields("event_level").
			StorageKey("idx_tm_cat_default_features_event_level"),
		index.Fields("call_mode").
			StorageKey("idx_tm_cat_default_features_call_mode"),

		// 常用筛选
		index.Fields("is_enabled").
			StorageKey("idx_tm_cat_default_features_enabled"),
		index.Fields("tenant_id").
			StorageKey("idx_tm_cat_default_features_tenant"),
	}
}
