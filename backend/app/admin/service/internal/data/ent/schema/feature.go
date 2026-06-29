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

// Feature 物模型特征表（统一承载属性/事件/服务/关系的"骨架"）
// Feature is the unified thing-model feature schema — skeleton-only since CR-001.
//
// 设计依据 / Design ref:
//   - docs/thingmodel/sheji/09-特征数据模型设计.md §2
//   - docs/thingmodel/sheji/修改记录/CR-001-结构化约束下沉到模型层.md
//
// CR-001（2026-06-29）后变更：
//   - spec 列移除；
//   - data_type/access_mode/event_level/call_mode/relation_type 5 个特化抽取列移除；
//   - 新增 recommended_unit_category_id / semantic_tag 两个推荐元信息字段（不参与约束计算）。
//
// 本表现在只承载特征的身份/语义骨架；结构化约束完整下沉到
// thingmodel_category_default_features.spec 与 thingmodel_product_features.spec。
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
		schema.Comment("物模型-特征骨架表（CR-001 后不再承载 spec）/ Thing model feature skeleton"),
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

		// ===== CR-001 新增推荐元信息（仅用于 UI 提示与检索，不参与约束计算）=====
		field.Uint32("recommended_unit_category_id").
			Comment("推荐单位物理量分类 ID（不指定具体单位，仅 UI 预过滤）/ Recommended unit category").
			Optional().
			Nillable(),

		field.String("semantic_tag").
			Comment("语义标签，如 pressure/temperature/runMode / Semantic tag").
			Optional().
			Nillable(),
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

		// CR-001 新增：按语义标签检索
		index.Fields("semantic_tag").
			StorageKey("idx_thingmodel_feature_semantic_tag"),

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
