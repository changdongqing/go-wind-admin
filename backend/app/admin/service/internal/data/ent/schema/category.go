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

// Category 物模型-分类表（智能系统/空间/设备设施 统一承载，未来可扩展 kind）。
// Category — unified thing-model category table covering SYSTEM / SPACE / FACILITY (kind extensible).
//
// 设计要点 / Design notes:
//   - 单表 + kind 枚举承载多业务域；未来追加 kind 枚举值即可承载文档/证书/表单等分类。
//   - 层级固定 4 层（level=1..4），code 长度按 level 严格分级：2/4/6/8 位纯数字，绝不补 0。
//   - 不存 path / is_leaf 字段：父子关系由 code 截断推导，叶子由 level==4 推导。
//   - 唯一键退化为 (tenant_id, kind, code) 3 元组，无需 level 参与。
type Category struct {
	ent.Schema
}

func (Category) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "thingmodel_categories",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("物模型-分类表 / Thing model category"),
	}
}

// Fields of the Category.
func (Category) Fields() []ent.Field {
	return []ent.Field{
		// 分类种类（不可变）；未来扩展只需在 NamedValues 追加新枚举值
		// Optional+Nillable：让 ent 生成 *Kind 指针字段，匹配 mapper.EnumTypeConverter 的 *ENTITY↔*DTO 转换；
		// 应用层永远 SetKind 非空（Service 层 V1 校验保证）。
		field.Enum("kind").
			NamedValues(
				"System", "SYSTEM",
				"Space", "SPACE",
				"Facility", "FACILITY",
				// 未来扩展示例 / Future placeholders:
				// "Document",    "DOCUMENT",
				// "Certificate", "CERTIFICATE",
				// "Form",        "FORM",
			).
			Immutable().
			Optional().
			Nillable().
			Comment("分类种类（不可变）/ Category kind, immutable"),

		// 变长 code（2/4/6/8 位纯数字）
		field.String("code").
			MaxLen(8).
			NotEmpty().
			Immutable().
			Optional().
			Nillable().
			Comment("分类编码（变长 2/4/6/8 位纯数字，按 level 决定长度，不可变）/ Variable-length numeric code, immutable"),

		// 层级 1..4
		// Optional+Nillable：让 ent 生成 *uint8 指针字段，让 copier 在 ent→DTO 转换时正确填充 proto 的 *uint32 字段；
		// 应用层永远 SetLevel 非空（Service 层 V3 校验保证）。
		field.Uint8("level").
			Min(1).Max(4).
			Optional().
			Nillable().
			Comment("层级：1=大类 2=中类 3=小类 4=细类 / Hierarchy level (1..4)"),

		// 父节点 ID（自引用）
		field.Uint32("parent_id").
			Optional().
			Nillable().
			Comment("父节点 ID（level=1 时为空）/ Parent category id (nullable when level=1)"),

		// 中/英文名
		field.String("name").
			NotEmpty().Optional().Nillable().
			Comment("中文名 / Category name (zh)"),
		field.String("name_en").
			Optional().Nillable().
			Comment("英文名 / Category name (en)"),

		// 图标/描述
		field.String("icon").Optional().Nillable().
			Comment("Iconify 图标名 / Iconify icon name"),
		field.String("description").Optional().Nillable().
			Comment("描述 / Description"),

		// 被引用次数（预留，本期恒 0）
		field.Uint32("reference_count").
			Default(0).
			Optional().
			Nillable().
			Comment("被物模型/实例引用次数（预留，本期恒 0）/ Reference count (reserved)"),
	}
}

// Mixin of the Category.
func (Category) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.IsEnabled{},
		mixin.SortOrder{},
		mixin.TenantID[uint32]{},
	}
}

// Edges of the Category. 自引用（父 → 多子）；删除受限（RESTRICT）。
func (Category) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", Category.Type).
			From("parent").
			Field("parent_id").
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Restrict,
			}),
	}
}

// Indexes of the Category.
func (Category) Indexes() []ent.Index {
	return []ent.Index{
		// ===== 唯一约束（核心）/ Unique constraint =====
		// (tenant_id, kind, code) 全局唯一 —— code 长度不同自然不冲突，3 元组足够
		index.Fields("tenant_id", "kind", "code").
			Unique().
			StorageKey("uix_thingmodel_cat_tenant_kind_code"),

		// ===== 树/查询索引 =====
		// 按 kind+level 查询（"取所有大类"/"取所有细类"）
		index.Fields("tenant_id", "kind", "level").
			StorageKey("idx_thingmodel_cat_tenant_kind_level"),

		// 按父节点取直接子节点（编辑/删除时校验）
		index.Fields("tenant_id", "kind", "parent_id").
			StorageKey("idx_thingmodel_cat_tenant_kind_parent"),

		// ===== 常用筛选 =====
		index.Fields("is_enabled").
			StorageKey("idx_thingmodel_cat_is_enabled"),
		index.Fields("sort_order").
			StorageKey("idx_thingmodel_cat_sort_order"),
		index.Fields("tenant_id").
			StorageKey("idx_thingmodel_cat_tenant_id"),

		// 名称模糊搜索
		index.Fields("name").
			StorageKey("idx_thingmodel_cat_name"),

		// 注：子树前缀查询（code LIKE 'xxx%'）由 (tenant_id, kind, code) 唯一索引天然支持
	}
}
