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

// ProductFeature 产品下特征条目（产品物模型的真正载体）
// ProductFeature is the actual carrier of a product's thing-model — each row represents one
// feature instance under a product. Three flavors via `source`:
//   - DEFAULT: copied from a category default feature; ref_feature_id points to thing_features
//   - GLOBAL:  added directly from the global feature library; ref_feature_id non-zero
//   - LOCAL:   product-local custom feature; ref_feature_id is null/zero
//
// 设计依据 / Design ref:
//   - docs/thingmodel/sheji/模型管理/02-数据模型设计.md §2.3
//   - docs/thingmodel/sheji/修改记录/CR-001-结构化约束下沉到模型层.md
//
// CR-001（2026-06-29）后变更：
//   - feature_snapshot + override_spec 两字段合并为单一 spec(FeatureSpec)；
//   - effectiveSpec 合并函数及对应字段彻底删除（spec 即唯一来源）；
//   - PUBLISHED 状态可改字段白名单收紧为 name/description/sort_order/is_enabled。
//
// 关键约束 / Invariants:
//   - source=LOCAL ↔ ref_feature_id IS NULL；source∈{DEFAULT,GLOBAL} ↔ ref_feature_id IS NOT NULL（Service 校验）
//   - spec 必填，其 oneof 分支必须与 feature_type 一致（Service 校验）
//   - V1–V17 spec 完整性校验由 Service 在 Create/Update 时执行
//   - 产品 status=PUBLISHED 后禁止增删条目；禁止改 spec 字段（Service 校验器）
//   - 删除 source=GLOBAL/LOCAL 行 → 维护 thingmodel_units.reference_count；DEFAULT 不动
//   - (product_id, code)、(product_id, identifier) 各自唯一
type ProductFeature struct {
	ent.Schema
}

func (ProductFeature) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "thingmodel_product_features",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("产品下特征条目 / Product feature entry"),
	}
}

// Fields of the ProductFeature.
func (ProductFeature) Fields() []ent.Field {
	return []ent.Field{
		// 父产品（CASCADE 删除）
		field.Uint32("product_id").
			Comment("父产品 ID / Product id"),

		// 来源
		field.Enum("source").
			NamedValues(
				"Default", "DEFAULT", // 从分类默认模型拷贝
				"Global", "GLOBAL", // 从全局特征库新加
				"Local", "LOCAL", // 产品本地自定义
			).
			Comment("来源 / Source"),

		// 引用全局特征 ID（LOCAL 时为空）
		field.Uint32("ref_feature_id").
			Optional().
			Nillable().
			Comment("引用全局特征 ID（LOCAL 时为空）/ Referenced feature id"),

		// ===== 公共字段 =====
		field.Enum("feature_type").
			NamedValues(
				"Property", "PROPERTY",
				"Event", "EVENT",
				"Service", "SERVICE",
				"Relation", "RELATION",
			).
			Comment("特征类型 / Feature type"),

		field.String("code").
			MaxLen(128).
			NotEmpty().
			Optional().
			Nillable().
			Comment("产品内编码 / Code within product"),
		field.String("identifier").
			MaxLen(128).
			NotEmpty().
			Optional().
			Nillable().
			Comment("产品内程序标识符 / Identifier within product"),
		field.String("name").
			MaxLen(128).
			NotEmpty().
			Optional().
			Nillable().
			Comment("名称 / Name (zh)"),
		field.String("name_en").
			MaxLen(128).
			Optional().
			Nillable().
			Comment("英文名 / Name (en)"),
		field.String("description").
			Optional().
			Nillable().
			Comment("描述 / Description"),

		// ===== CR-001：单一完整 FeatureSpec（产品最终物模型）/ Single full FeatureSpec =====
		// 用 FeatureSpecField wrapper 走 protojson 编码 oneof。
		field.JSON("spec", &FeatureSpecField{}).
			Optional().
			Comment("完整 FeatureSpec（CR-001 合并 snapshot+override+effective，protojson）/ Full feature spec"),

		// ===== 冗余特化列（从 spec 派生，便于筛选）=====
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
			Comment("property 数据类型（冗余）/ Property data type"),
		field.Enum("access_mode").
			NamedValues(
				"R", "R",
				"RW", "RW",
			).
			Optional().
			Nillable().
			Comment("property 访问模式 R/RW / Access mode"),
		field.Enum("event_level").
			NamedValues(
				"Info", "INFO",
				"Alert", "ALERT",
				"Error", "ERROR",
			).
			Optional().
			Nillable().
			Comment("event 级别 INFO/ALERT/ERROR / Event level"),
		field.Enum("call_mode").
			NamedValues(
				"Async", "ASYNC",
				"Sync", "SYNC",
			).
			Optional().
			Nillable().
			Comment("service 调用模式 ASYNC/SYNC / Call mode"),
		field.String("relation_type").
			Optional().
			Nillable().
			Comment("relation 关系类型 / Relation type"),
	}
}

// Mixin of the ProductFeature.
func (ProductFeature) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.IsEnabled{},
		mixin.SortOrder{},
		mixin.TenantID[uint32]{},
	}
}

// Edges of the ProductFeature.
// 上游：Product（反向引用 "features"，已在 Product schema 定义）
// CASCADE：产品删除 → 本表条目随之删除
func (ProductFeature) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).
			Ref("features").
			Field("product_id").
			Required().
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

// Indexes of the ProductFeature.
func (ProductFeature) Indexes() []ent.Index {
	return []ent.Index{
		// 产品内 code 唯一
		index.Fields("product_id", "code").
			Unique().
			StorageKey("uix_tm_pf_product_code"),
		// 产品内 identifier 唯一
		index.Fields("product_id", "identifier").
			Unique().
			StorageKey("uix_tm_pf_product_identifier"),

		// 主查询：按产品 + 类型列条目，按 sort_order 排序
		index.Fields("product_id", "feature_type", "sort_order").
			StorageKey("idx_tm_pf_product_type_sort"),

		// 反向：某全局特征被哪些产品引用
		index.Fields("ref_feature_id").
			StorageKey("idx_tm_pf_ref_feature"),

		// 按来源筛选（DEFAULT/GLOBAL/LOCAL）
		index.Fields("source").
			StorageKey("idx_tm_pf_source"),

		// 租户级查询
		index.Fields("tenant_id", "product_id").
			StorageKey("idx_tm_pf_tenant_product"),

		// 特化列筛选（property 列出 data_type）
		index.Fields("feature_type", "data_type").
			StorageKey("idx_tm_pf_type_datatype"),

		// 常用筛选
		index.Fields("is_enabled").
			StorageKey("idx_tm_pf_enabled"),
		index.Fields("tenant_id").
			StorageKey("idx_tm_pf_tenant"),
		index.Fields("sort_order").
			StorageKey("idx_tm_pf_sort_order"),
	}
}
