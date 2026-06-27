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
// to a global feature, with optional sparse override.
//
// 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/02-数据模型设计.md §2.1
//
// 关键约束（应用层兜底）/ Application-layer invariants:
//   - category_id 必须指向 level=4 节点（Service 校验）
//   - override_spec 仅允许白名单字段（由 FeatureOverrideSpec proto 结构本身收口）
//   - 写入/删除时维护 thing_features.reference_count（来自 feature_id）
//   - 写入/删除时维护 thingmodel_units.reference_count（来自 snapshot.unit 或 override.unit）
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
		schema.Comment("分类默认模型条目 / Category default model entry"),
	}
}

// Fields of the CategoryDefaultFeature.
func (CategoryDefaultFeature) Fields() []ent.Field {
	return []ent.Field{
		// 分类 ID（必须 level=4）
		field.Uint32("category_id").
			Comment("分类 ID（必须 level=4）/ Category id, must be level=4"),

		// 全局特征 ID
		field.Uint32("feature_id").
			Comment("全局特征 ID / Global feature id"),

		// 稀疏覆写 spec（白名单字段：constraints/unit/defaultValue/displayName/description/required）
		// 用 FeatureOverrideSpecField wrapper 走 protojson，规避 wkt wrapper round-trip 问题。
		// 详见 featureoverridespec_jsonfield.go 与 backend/CLAUDE.md「Step 13」。
		field.JSON("override_spec", &FeatureOverrideSpecField{}).
			Optional().
			Comment("稀疏覆写（白名单字段，protojson 编码）/ Sparse override (protojson)"),

		// 分类内展示别名
		field.String("display_name").
			MaxLen(128).
			Optional().
			Nillable().
			Comment("分类内展示别名 / Display alias within category"),
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

		// 常用筛选
		index.Fields("is_enabled").
			StorageKey("idx_tm_cat_default_features_enabled"),
		index.Fields("tenant_id").
			StorageKey("idx_tm_cat_default_features_tenant"),
	}
}
