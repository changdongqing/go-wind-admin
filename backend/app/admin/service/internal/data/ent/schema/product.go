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

// Product 物模型-产品主表
// Product is the main entity of the model management module — a concrete, instantiable prototype
// belonging to exactly one level=4 Category. Its features live in thingmodel_product_features
// (CASCADE on delete).
//
// 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/02-数据模型设计.md §2.2
//
// 关键约束 / Invariants:
//   - code 与 category_id 均不可变（Immutable）
//   - (tenant, code) 唯一；(tenant, category_id, name) 唯一
//   - category 必须 level=4（Service 校验）
//   - reference_count > 0 时禁止物理删除（Service 拦截）
//   - status=PUBLISHED 后特征结构冻结（仅 override 可改；Service 校验器）
type Product struct {
	ent.Schema
}

func (Product) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "thingmodel_products",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("物模型-产品表 / Thing model product"),
	}
}

// Fields of the Product.
func (Product) Fields() []ent.Field {
	return []ent.Field{
		// 产品编码（程序标识符，租户内唯一，不可变）
		field.String("code").
			MaxLen(64).
			NotEmpty().
			Immutable().
			Optional().
			Nillable().
			Comment("产品编码（程序标识符，租户内唯一，不可变）/ Product code, immutable"),

		// 中/英文名
		field.String("name").
			MaxLen(128).
			NotEmpty().
			Optional().
			Nillable().
			Comment("产品中文名 / Name (zh)"),
		field.String("name_en").
			MaxLen(128).
			Optional().
			Nillable().
			Comment("产品英文名 / Name (en)"),

		// 分类（必须 level=4，不可变）
		field.Uint32("category_id").
			Immutable().
			Comment("分类 ID（必须 level=4，不可变）/ Category id, must be level=4"),

		// 制造商 / 型号（optional）
		field.String("manufacturer").
			MaxLen(128).
			Optional().
			Nillable().
			Comment("制造商/品牌 / Manufacturer"),
		field.String("model_no").
			MaxLen(64).
			Optional().
			Nillable().
			Comment("型号 / Model number"),

		field.String("icon").
			Optional().
			Nillable().
			Comment("Iconify 图标名 / Icon"),
		field.String("description").
			Optional().
			Nillable().
			Comment("描述 / Description"),

		// 发布状态机 DRAFT/PUBLISHED
		field.Enum("status").
			NamedValues(
				"Draft", "DRAFT",
				"Published", "PUBLISHED",
			).
			Default("DRAFT").
			Comment("发布状态 / Lifecycle status"),

		// 被设备实例引用次数（预留，本期恒 0）
		field.Uint32("reference_count").
			Default(0).
			Nillable().
			Comment("被设备实例引用次数（预留）/ Reference count (reserved)"),
	}
}

// Mixin of the Product.
func (Product) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.IsEnabled{},
		mixin.SortOrder{},
		mixin.TenantID[uint32]{},
	}
}

// Edges of the Product.
// 上游：Category（反向引用 "products"，由 Category 在 Task 11 增补）
// 下游：ProductFeature（一对多，CASCADE 删除）
func (Product) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("category", Category.Type).
			Ref("products").
			Field("category_id").
			Required().
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Restrict,
			}),

		edge.To("features", ProductFeature.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

// Indexes of the Product.
func (Product) Indexes() []ent.Index {
	return []ent.Index{
		// 唯一约束 #1：(tenant, code) — 产品程序标识符
		index.Fields("tenant_id", "code").
			Unique().
			StorageKey("uix_tm_products_tenant_code"),

		// 唯一约束 #2：(tenant, category_id, name) — 同分类下产品名不重
		index.Fields("tenant_id", "category_id", "name").
			Unique().
			StorageKey("uix_tm_products_tenant_cat_name"),

		// 查询路径
		index.Fields("tenant_id", "category_id").
			StorageKey("idx_tm_products_tenant_cat"),
		index.Fields("tenant_id", "status").
			StorageKey("idx_tm_products_tenant_status"),
		index.Fields("manufacturer").
			StorageKey("idx_tm_products_manufacturer"),

		// 常用筛选
		index.Fields("is_enabled").
			StorageKey("idx_tm_products_enabled"),
		index.Fields("tenant_id").
			StorageKey("idx_tm_products_tenant"),
		index.Fields("sort_order").
			StorageKey("idx_tm_products_sort_order"),
	}
}
