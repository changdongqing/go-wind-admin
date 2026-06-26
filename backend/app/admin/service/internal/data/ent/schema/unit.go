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

// Unit 单位表 / Thing model unit
type Unit struct {
	ent.Schema
}

func (Unit) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "thingmodel_units",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("物模型-单位表 / Thing model unit"),
	}
}

// Fields of the Unit.
func (Unit) Fields() []ent.Field {
	return []ent.Field{
		// 所属物理量分类（必填）
		field.Uint32("category_id").
			Comment("所属物理量分类ID / Category ID").
			Nillable(),

		// 单位编码（不可变，租户内唯一）
		field.String("code").
			Comment("单位编码，如 celsius（不可变）/ Unit code, immutable").
			NotEmpty().
			Immutable().
			Optional().
			Nillable(),

		// 单位符号
		field.String("symbol").
			Comment("单位符号，如 ℃ / Unit symbol").
			NotEmpty().
			Optional().
			Nillable(),

		// 中文名
		field.String("name").
			Comment("中文名 / Unit name (zh)").
			NotEmpty().
			Optional().
			Nillable(),

		// 英文名
		field.String("name_en").
			Comment("英文名 / Unit name (en)").
			Optional().
			Nillable(),

		// 是否基准单位（每分类唯一，由 partial unique index 保证）
		field.Bool("is_base").
			Comment("是否基准单位（每分类唯一）/ Is base unit").
			Default(false).
			Nillable(),

		// 换算类型
		field.Enum("conversion_type").
			Comment("换算类型 / Conversion type").
			NamedValues(
				"Linear", "LINEAR",
				"Affine", "AFFINE",
				"Logarithmic", "LOGARITHMIC",
				"Conditional", "CONDITIONAL",
				"None", "NONE",
			).
			Default("LINEAR").
			Optional().
			Nillable(),

		// 线性系数 k（正向：基准 = 原值·factor + offset）
		field.Float("factor").
			Comment("线性系数 k（基准=原值·k+offset）/ Linear factor k").
			Default(1.0).
			Nillable(),

		// 偏移量 b（仅仿射类型非0，如温度）
		field.Float("offset").
			Comment("偏移量 b（仅仿射非0）/ Offset b (affine only)").
			Default(0.0).
			Nillable(),

		// 公式说明（仅展示，不参与计算）
		field.String("formula_expr").
			Comment("公式说明（仅展示）/ Formula description (display only)").
			Optional().
			Nillable(),

		// 建议显示精度（小数位）
		field.Int32("precision").
			Comment("建议显示精度（小数位）/ Display precision").
			Default(2).
			Optional().
			Nillable(),

		// 是否 SI 单位（含 SI 词头组合）
		field.Bool("is_si_unit").
			Comment("是否 SI 单位 / Is SI unit").
			Default(false).
			Nillable(),

		// 是否中国法定计量单位（GB 3100-1993）
		field.Bool("is_legal_unit").
			Comment("是否中国法定计量单位 / Is PRC legal unit").
			Default(false).
			Nillable(),

		// 被物模型属性引用次数（预留，本期恒 0）
		field.Uint32("reference_count").
			Comment("被物模型属性引用次数（预留）/ Reference count (reserved)").
			Default(0).
			Nillable(),
	}
}

// Mixin of the Unit.
func (Unit) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.SortOrder{},
		mixin.IsEnabled{},
		mixin.TenantID[uint32]{},
	}
}

// Edges of the Unit.
func (Unit) Edges() []ent.Edge {
	return []ent.Edge{
		// 多对一：归属物理量分类（category_id 为显式 field 且 NOT NULL，故 edge Required）
		edge.From("category", UnitCategory.Type).
			Ref("units").
			Field("category_id").
			Unique().
			Required(),
	}
}

// Indexes of the Unit.
func (Unit) Indexes() []ent.Index {
	return []ent.Index{
		// 租户级唯一：同一租户下 code 唯一（程序引用的稳定标识）
		index.Fields("tenant_id", "code").
			Unique().
			StorageKey("uix_thingmodel_unit_tenant_code"),

		// 按分类查询单位（最高频查询：属性选单位时拉某分类下全部单位）
		index.Fields("tenant_id", "category_id").
			StorageKey("idx_thingmodel_unit_tenant_category"),
		index.Fields("category_id").
			StorageKey("idx_thingmodel_unit_category_id"),

		// 基准单位定位（每分类查 is_base=true）
		index.Fields("category_id", "is_base").
			StorageKey("idx_thingmodel_unit_category_base"),

		// 常用筛选
		index.Fields("is_enabled").
			StorageKey("idx_thingmodel_unit_is_enabled"),
		index.Fields("conversion_type").
			StorageKey("idx_thingmodel_unit_conversion_type"),
		index.Fields("tenant_id").
			StorageKey("idx_thingmodel_unit_tenant_id"),

		// 符号模糊搜索
		index.Fields("symbol").
			StorageKey("idx_thingmodel_unit_symbol"),
	}
}
