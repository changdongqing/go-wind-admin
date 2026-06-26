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

// UnitCategory 物理量分类表 / Thing model unit category
type UnitCategory struct {
	ent.Schema
}

func (UnitCategory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "thingmodel_unit_categories",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("物模型-物理量分类表 / Thing model unit category"),
	}
}

// Fields of the UnitCategory.
func (UnitCategory) Fields() []ent.Field {
	return []ent.Field{
		// 物理量编码（不可变，租户内唯一）
		field.String("code").
			Comment("物理量编码，如 temperature/pressure（不可变）/ Category code, immutable").
			NotEmpty().
			Immutable().
			Optional().
			Nillable(),

		// 中文名
		field.String("name").
			Comment("中文名，如 温度 / Category name (zh)").
			NotEmpty().
			Optional().
			Nillable(),

		// 英文名
		field.String("name_en").
			Comment("英文名，如 Temperature / Category name (en)").
			Optional().
			Nillable(),

		// 量纲/物理量名
		field.String("quantity").
			Comment("量纲/物理量名，如 热力学温度 / Quantity").
			Optional().
			Nillable(),

		// 基准单位符号（冗余展示）
		field.String("base_unit_symbol").
			Comment("基准单位符号（冗余展示，如 K）/ Base unit symbol (display only)").
			Optional().
			Nillable(),

		// Iconify 图标
		field.String("icon").
			Comment("Iconify 图标名 / Iconify icon name").
			Optional().
			Nillable(),

		// 描述
		field.String("description").
			Comment("描述 / Description").
			Optional().
			Nillable(),
	}
}

// Mixin of the UnitCategory.
func (UnitCategory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.IsEnabled{},
		mixin.SortOrder{},
		mixin.TenantID[uint32]{},
	}
}

// Edges of the UnitCategory.
func (UnitCategory) Edges() []ent.Edge {
	return []ent.Edge{
		// 一个分类下挂多个单位，删除分类时级联删除单位
		// 外键列 category_id 在 Unit 端作为显式 field 定义；Required 仅在 Unit 端（多端）声明
		edge.To("units", Unit.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

// Indexes of the UnitCategory.
func (UnitCategory) Indexes() []ent.Index {
	return []ent.Index{
		// 租户级唯一：同一租户下 code 唯一
		index.Fields("tenant_id", "code").
			Unique().
			StorageKey("uix_thingmodel_unit_cat_tenant_code"),

		// 按租户快速筛选
		index.Fields("tenant_id").
			StorageKey("idx_thingmodel_unit_cat_tenant_id"),

		// 按启用状态过滤
		index.Fields("is_enabled").
			StorageKey("idx_thingmodel_unit_cat_is_enabled"),

		// 按排序值查询/排序
		index.Fields("sort_order").
			StorageKey("idx_thingmodel_unit_cat_sort_order"),
	}
}
