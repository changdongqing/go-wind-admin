# 模型管理 · 后端实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在已落地的 `thingmodel` 域基础上，落地"分类默认模型 + 产品管理"后端能力（proto 契约 / ent schema / repo / service / 校验器 / wire / menu / seed / 测试），与设计文档 [`docs/thingmodel/sheji/模型管理/`](../../thingmodel/sheji/模型管理/) 1:1 对齐。

**Architecture:** 严格遵循项目"Contract-First Pipeline"——proto 源领域层定义 message 与全 gRPC service、BFF 层加 `google.api.http` 注解。三层架构（ent schema → repo → service handler），无 biz 中间层（与现有 unit/feature/category 一致）。所有 JSON 字段走 `FeatureSpecField` / `FeatureOverrideSpecField` 包装类型 + protojson，规避 oneof round-trip 陷阱（CLAUDE.md Step 13）。三类新表通过双向 edge 关联到 `categories` 与 `features`，应用层维护 `reference_count`，单事务保证一致性。

**Tech Stack:** Go 1.25+ / Kratos / Ent / buf / protojson / Wire DI / PostgreSQL ≥ 13 / Redis ≥ 8.0

**前置依赖**：单位管理 / 特征管理 / 分类管理 三模块已落地（thingmodel 域已存在）。

**范围**：仅后端。前端（React / Vue Element / Vue Vben）三份独立实施计划在后端 API 稳定（Phase 1.6 验收通过）后另起。

---

## 文件结构（落点全景）

新增：
```
backend/api/protos/thingmodel/service/v1/
  feature_override_spec.proto             ← Task 1
  category_default_feature.proto          ← Task 2
  product.proto                           ← Task 3
  product_feature.proto                   ← Task 4

backend/api/protos/admin/service/v1/
  thingmodel_category_default_feature.proto  ← Task 5
  thingmodel_product.proto                   ← Task 5
  thingmodel_product_feature.proto           ← Task 5

backend/app/admin/service/internal/data/ent/schema/
  featureoverridespec_jsonfield.go         ← Task 7（含 _test.go）
  category_default_feature.go              ← Task 8
  product.go                               ← Task 9
  product_feature.go                       ← Task 10

backend/app/admin/service/internal/data/
  category_default_feature_repo.go         ← Task 13（含 _test.go）
  product_repo.go                          ← Task 14
  product_feature_repo.go                  ← Task 15

backend/app/admin/service/internal/service/
  product_validator.go                     ← Task 16（覆写白名单 + effective_spec）
  category_default_feature_service.go      ← Task 17（含 _test.go）
  product_service.go                       ← Task 18
  product_feature_service.go               ← Task 19（含 PullFromDefault）

backend/app/admin/service/internal/data/seed/
  model_management_seed.go                 ← Task 22
  model_management_seed_data.go            ← Task 22
  model_management_seed_test.go            ← Task 22
```

修改：
```
backend/app/admin/service/internal/data/ent/schema/
  category.go                              ← Task 11（追加反向 edge）
  feature.go                               ← Task 11（追加反向 edge）

backend/app/admin/service/internal/data/data.go            ← Task 21（ProviderSet）
backend/app/admin/service/internal/service/service.go     ← Task 21
backend/app/admin/service/internal/server/grpc.go         ← Task 21
backend/app/admin/service/internal/server/http.go         ← Task 21
backend/pkg/constants/default_data.go                     ← Task 20（DefaultMenus）
```

---

## Task 1: 源领域 proto · FeatureOverrideSpec

**Files:**
- Create: `backend/api/protos/thingmodel/service/v1/feature_override_spec.proto`

- [ ] **Step 1: 写 proto 文件**

```protobuf
syntax = "proto3";

package thingmodel.service.v1;

option go_package = "go-wind-admin/api/gen/go/thingmodel/service/v1;v1";

import "thingmodel/service/v1/feature.proto";   // 复用 ValueConstraints / UnitRef
import "google/protobuf/struct.proto";
import "google/protobuf/wrappers.proto";

// 特征覆写规约（白名单字段）
// 用于：
//   thingmodel_category_default_features.override_spec
//   thingmodel_product_features.override_spec
message FeatureOverrideSpec {
  ValueConstraints     constraints   = 1;
  UnitRef              unit          = 2;
  google.protobuf.Value default_value = 3 [json_name = "defaultValue"];
  string               display_name  = 4 [json_name = "displayName"];
  string               description   = 5;
  google.protobuf.BoolValue required = 6;
}
```

- [ ] **Step 2: buf lint 通过**

Run: `cd backend && buf lint api/protos/thingmodel/service/v1/feature_override_spec.proto`
Expected: 无输出

- [ ] **Step 3: Commit**

```bash
git add backend/api/protos/thingmodel/service/v1/feature_override_spec.proto
git commit -m "feat(thingmodel): add FeatureOverrideSpec proto for model management"
```

---

## Task 2: 源领域 proto · CategoryDefaultFeature

**Files:**
- Create: `backend/api/protos/thingmodel/service/v1/category_default_feature.proto`

- [ ] **Step 1: 写 proto**

```protobuf
syntax = "proto3";

package thingmodel.service.v1;

option go_package = "go-wind-admin/api/gen/go/thingmodel/service/v1;v1";

import "thingmodel/service/v1/feature.proto";
import "thingmodel/service/v1/feature_override_spec.proto";
import "google/protobuf/timestamp.proto";

// 分类默认模型条目
message CategoryDefaultFeature {
  uint32 id            = 1;
  uint32 category_id   = 2 [json_name = "categoryId"];
  uint32 feature_id    = 3 [json_name = "featureId"];

  // 关联只读字段（List/Get 时由后端 join 填充）
  string feature_code        = 10 [json_name = "featureCode"];
  string feature_identifier  = 11 [json_name = "featureIdentifier"];
  string feature_name        = 12 [json_name = "featureName"];
  FeatureType feature_type   = 13 [json_name = "featureType"];
  FeatureSpec feature_snapshot_preview = 14 [json_name = "featureSnapshotPreview"];

  FeatureOverrideSpec override_spec = 20 [json_name = "overrideSpec"];
  string display_name  = 21 [json_name = "displayName"];
  bool   is_enabled    = 22 [json_name = "isEnabled"];
  uint32 sort_order    = 23 [json_name = "sortOrder"];
  uint32 tenant_id     = 24 [json_name = "tenantId"];

  google.protobuf.Timestamp created_at = 30 [json_name = "createdAt"];
  google.protobuf.Timestamp updated_at = 31 [json_name = "updatedAt"];
}
```

- [ ] **Step 2: buf lint 通过**

Run: `cd backend && buf lint api/protos/thingmodel/service/v1/category_default_feature.proto`

- [ ] **Step 3: Commit**

```bash
git add backend/api/protos/thingmodel/service/v1/category_default_feature.proto
git commit -m "feat(thingmodel): add CategoryDefaultFeature proto"
```

---

## Task 3: 源领域 proto · Product

**Files:**
- Create: `backend/api/protos/thingmodel/service/v1/product.proto`

- [ ] **Step 1: 写 proto**

```protobuf
syntax = "proto3";

package thingmodel.service.v1;

option go_package = "go-wind-admin/api/gen/go/thingmodel/service/v1;v1";

import "google/protobuf/timestamp.proto";

enum ProductStatus {
  PRODUCT_STATUS_UNSPECIFIED = 0;
  DRAFT     = 1;
  PUBLISHED = 2;
}

message Product {
  uint32 id            = 1;
  string code          = 2;
  string name          = 3;
  string name_en       = 4 [json_name = "nameEn"];
  uint32 category_id   = 5 [json_name = "categoryId"];

  string category_code = 10 [json_name = "categoryCode"];
  string category_name = 11 [json_name = "categoryName"];

  string manufacturer  = 20;
  string model_no      = 21 [json_name = "modelNo"];
  string icon          = 22;
  string description   = 23;

  ProductStatus status = 30;
  bool   is_enabled    = 31 [json_name = "isEnabled"];
  uint32 sort_order    = 32 [json_name = "sortOrder"];
  uint32 reference_count = 33 [json_name = "referenceCount"];
  uint32 tenant_id     = 34 [json_name = "tenantId"];

  uint32 feature_count = 40 [json_name = "featureCount"];

  google.protobuf.Timestamp created_at = 50 [json_name = "createdAt"];
  google.protobuf.Timestamp updated_at = 51 [json_name = "updatedAt"];
}
```

- [ ] **Step 2: buf lint 通过**

- [ ] **Step 3: Commit**

```bash
git add backend/api/protos/thingmodel/service/v1/product.proto
git commit -m "feat(thingmodel): add Product proto with DRAFT/PUBLISHED status"
```

---

## Task 4: 源领域 proto · ProductFeature

**Files:**
- Create: `backend/api/protos/thingmodel/service/v1/product_feature.proto`

- [ ] **Step 1: 写 proto**

```protobuf
syntax = "proto3";

package thingmodel.service.v1;

option go_package = "go-wind-admin/api/gen/go/thingmodel/service/v1;v1";

import "thingmodel/service/v1/feature.proto";
import "thingmodel/service/v1/feature_override_spec.proto";
import "google/protobuf/timestamp.proto";

enum ProductFeatureSource {
  PRODUCT_FEATURE_SOURCE_UNSPECIFIED = 0;
  DEFAULT = 1;   // 从分类默认模型拷贝
  GLOBAL  = 2;   // 从全局特征库新加
  LOCAL   = 3;   // 产品本地自定义
}

enum ConflictPolicy {
  CONFLICT_POLICY_UNSPECIFIED = 0;
  SKIP    = 1;
  REPLACE = 2;
}

message ProductFeature {
  uint32 id              = 1;
  uint32 product_id      = 2 [json_name = "productId"];
  ProductFeatureSource source = 3;
  uint32 ref_feature_id  = 4 [json_name = "refFeatureId"];

  FeatureType feature_type = 10 [json_name = "featureType"];
  string code            = 11;
  string identifier      = 12;
  string name            = 13;
  string name_en         = 14 [json_name = "nameEn"];
  string description     = 15;

  FeatureSpec         feature_snapshot = 20 [json_name = "featureSnapshot"];
  FeatureOverrideSpec override_spec    = 21 [json_name = "overrideSpec"];

  // 后端在 Get 时填充
  FeatureSpec effective_spec = 22 [json_name = "effectiveSpec"];

  DataType   data_type     = 30 [json_name = "dataType"];
  AccessMode access_mode   = 31 [json_name = "accessMode"];
  EventLevel event_level   = 32 [json_name = "eventLevel"];
  CallMode   call_mode     = 33 [json_name = "callMode"];
  string     relation_type = 34 [json_name = "relationType"];

  bool   is_enabled  = 40 [json_name = "isEnabled"];
  uint32 sort_order  = 41 [json_name = "sortOrder"];
  uint32 tenant_id   = 42 [json_name = "tenantId"];

  google.protobuf.Timestamp created_at = 50 [json_name = "createdAt"];
  google.protobuf.Timestamp updated_at = 51 [json_name = "updatedAt"];
}
```

- [ ] **Step 2: buf lint 通过**

- [ ] **Step 3: Commit**

```bash
git add backend/api/protos/thingmodel/service/v1/product_feature.proto
git commit -m "feat(thingmodel): add ProductFeature proto with source enum"
```

---

## Task 5: BFF 层 proto（3 个文件 + 错误码追加）

**Files:**
- Create: `backend/api/protos/admin/service/v1/thingmodel_category_default_feature.proto`
- Create: `backend/api/protos/admin/service/v1/thingmodel_product.proto`
- Create: `backend/api/protos/admin/service/v1/thingmodel_product_feature.proto`
- Modify: `backend/api/protos/thingmodel/service/v1/thingmodel_error.proto`（追加 reason）

参考：设计文档 [03 §3](../../thingmodel/sheji/模型管理/03-API与Proto设计.md)

- [ ] **Step 1: 追加错误码到 `thingmodel_error.proto`**

```protobuf
// ===== 模型管理 ===== (追加到现有 enum)
TM_CAT_DEFAULT_FEATURE_CATEGORY_NOT_LEAF = 4001;
TM_CAT_DEFAULT_FEATURE_DUPLICATE         = 4002;
TM_CAT_DEFAULT_FEATURE_OVERRIDE_INVALID  = 4003;
TM_CAT_DEFAULT_FEATURE_FEATURE_DISABLED  = 4004;
TM_PRODUCT_CATEGORY_NOT_LEAF             = 4101;
TM_PRODUCT_CODE_DUPLICATE                = 4102;
TM_PRODUCT_NAME_DUPLICATE                = 4103;
TM_PRODUCT_REFERENCED                    = 4104;
TM_PRODUCT_STATUS_NOT_DRAFT              = 4105;
TM_PRODUCT_STATUS_NOT_PUBLISHED          = 4106;
TM_PF_SOURCE_REF_MISMATCH                = 4201;
TM_PF_DUPLICATE_CODE                     = 4202;
TM_PF_DUPLICATE_IDENTIFIER               = 4203;
TM_PF_SPEC_TYPE_MISMATCH                 = 4204;
TM_PF_OVERRIDE_INVALID                   = 4205;
TM_PF_PRODUCT_PUBLISHED                  = 4206;
```

- [ ] **Step 2: 写 `thingmodel_category_default_feature.proto` BFF**

RPC 清单（详见设计文档 03 §3.1）：
```
ListCategoryDefaultFeatures    GET /v1/thingmodel/categories/{category_id}/default-features
GetCategoryDefaultFeature      GET /v1/thingmodel/category-default-features/{id}
CreateCategoryDefaultFeature   POST /v1/thingmodel/category-default-features
BatchAddCategoryDefaultFeatures POST /v1/thingmodel/categories/{category_id}/default-features:batch
UpdateCategoryDefaultFeature   PATCH /v1/thingmodel/category-default-features/{id}
DeleteCategoryDefaultFeature   DELETE /v1/thingmodel/category-default-features/{id}
ReorderCategoryDefaultFeatures POST /v1/thingmodel/categories/{category_id}/default-features:reorder
```

每条 RPC 写 Request/Response message + `google.api.http` 注解；FieldMask 走 google.protobuf.FieldMask（参照已有 `thingmodel_feature.proto` 风格）。

- [ ] **Step 3: 写 `thingmodel_product.proto` BFF**

RPC（03 §3.2）：
```
ListProducts / GetProduct / CreateProduct / UpdateProduct / DeleteProduct
PublishProduct   POST /v1/thingmodel/products/{id}:publish
UnpublishProduct POST /v1/thingmodel/products/{id}:unpublish
```

- [ ] **Step 4: 写 `thingmodel_product_feature.proto` BFF**

RPC（03 §3.3）：
```
ListProductFeatures / GetProductFeature / CreateProductFeature
PullFromDefault         POST /v1/thingmodel/products/{product_id}/features:pull-from-default
CloneFromProduct        POST /v1/thingmodel/products/{product_id}/features:clone-from
UpdateProductFeature / DeleteProductFeature
ReorderProductFeatures
```

- [ ] **Step 5: buf lint + buf breaking 检查**

Run: `cd backend && buf lint && buf breaking --against ".git#branch=origin/main,subdir=backend"`
Expected: 无输出（新增 message/service 不算 breaking）

- [ ] **Step 6: Commit**

```bash
git add backend/api/protos/admin/service/v1/thingmodel_*.proto \
        backend/api/protos/thingmodel/service/v1/thingmodel_error.proto
git commit -m "feat(thingmodel): add BFF protos & error codes for model management"
```

---

## Task 6: 生成代码 · make api + make ts

- [ ] **Step 1: 生成 Go API**

Run: `cd backend && make api`
Expected: `backend/api/gen/go/thingmodel/service/v1/` 与 `backend/api/gen/go/admin/service/v1/` 下出现新文件；编译无错

- [ ] **Step 2: 验证 Go 编译**

Run: `cd backend && go build ./...`
Expected: 无错（service handler 未实现导致 wire 报错可暂时忽略，下一阶段处理）

- [ ] **Step 3: 生成 TS 客户端**

Run: `cd backend && make ts`
Expected: 三个前端的 `api/generated/` 出现新 service client

- [ ] **Step 4: Commit 生成产物**

```bash
git add backend/api/gen/ frontend/admin/*/src/**/api/generated/
git commit -m "chore(gen): regenerate Go API & TS clients for model management"
```

---

## Task 7: JSON 包装类型 · FeatureOverrideSpecField

**Files:**
- Create: `backend/app/admin/service/internal/data/ent/schema/featureoverridespec_jsonfield.go`
- Create: `backend/app/admin/service/internal/data/ent/schema/featureoverridespec_jsonfield_test.go`

参考：现有 `featurespec_jsonfield.go`（同目录）模板。

- [ ] **Step 1: 写 round-trip 测试（TDD 红）**

```go
// featureoverridespec_jsonfield_test.go
package schema

import (
    "encoding/json"
    "testing"

    thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

func TestFeatureOverrideSpecField_RoundTrip(t *testing.T) {
    src := &thingmodelV1.FeatureOverrideSpec{
        Constraints: &thingmodelV1.ValueConstraints{Min: 5, Max: 12, Step: 0.5},
        DisplayName: "出水温度",
        Description: "冷冻水侧",
    }
    f := WrapFeatureOverrideSpec(src)

    b, err := json.Marshal(f)
    if err != nil { t.Fatal(err) }

    f2 := &FeatureOverrideSpecField{}
    if err := json.Unmarshal(b, f2); err != nil { t.Fatal(err) }

    got := UnwrapFeatureOverrideSpec(f2)
    if got.GetDisplayName() != "出水温度" { t.Fatal("displayName lost") }
    if got.GetConstraints().GetMax() != 12 { t.Fatal("constraints.max lost") }
}

func TestFeatureOverrideSpecField_NilSafety(t *testing.T) {
    if WrapFeatureOverrideSpec(nil) != nil { t.Fatal("nil wrap should return nil") }
    if UnwrapFeatureOverrideSpec(nil) != nil { t.Fatal("nil unwrap should return nil") }

    f := &FeatureOverrideSpecField{}
    b, err := json.Marshal(f)
    if err != nil { t.Fatal(err) }
    if string(b) != "null" { t.Fatalf("expected null, got %s", b) }
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./app/admin/service/internal/data/ent/schema/ -run TestFeatureOverrideSpecField -v`
Expected: FAIL（类型未定义）

- [ ] **Step 3: 实现包装类型（镜像 FeatureSpecField）**

```go
// featureoverridespec_jsonfield.go
package schema

import (
    "encoding/json"

    "google.golang.org/protobuf/encoding/protojson"

    thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// FeatureOverrideSpecField 包装 *thingmodelV1.FeatureOverrideSpec，
// 通过 protojson 处理（无 oneof 但保持风格统一）。详见 CLAUDE.md Step 13。
type FeatureOverrideSpecField struct {
    *thingmodelV1.FeatureOverrideSpec
}

var (
    _ json.Marshaler   = (*FeatureOverrideSpecField)(nil)
    _ json.Unmarshaler = (*FeatureOverrideSpecField)(nil)
)

func (f *FeatureOverrideSpecField) MarshalJSON() ([]byte, error) {
    if f == nil || f.FeatureOverrideSpec == nil {
        return []byte("null"), nil
    }
    return protojson.MarshalOptions{
        UseProtoNames: false, EmitUnpopulated: false,
    }.Marshal(f.FeatureOverrideSpec)
}

func (f *FeatureOverrideSpecField) UnmarshalJSON(data []byte) error {
    if len(data) == 0 || string(data) == "null" {
        f.FeatureOverrideSpec = nil
        return nil
    }
    if f.FeatureOverrideSpec == nil {
        f.FeatureOverrideSpec = &thingmodelV1.FeatureOverrideSpec{}
    }
    return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(data, f.FeatureOverrideSpec)
}

func WrapFeatureOverrideSpec(s *thingmodelV1.FeatureOverrideSpec) *FeatureOverrideSpecField {
    if s == nil { return nil }
    return &FeatureOverrideSpecField{FeatureOverrideSpec: s}
}

func UnwrapFeatureOverrideSpec(f *FeatureOverrideSpecField) *thingmodelV1.FeatureOverrideSpec {
    if f == nil { return nil }
    return f.FeatureOverrideSpec
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./app/admin/service/internal/data/ent/schema/ -run TestFeatureOverrideSpecField -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/app/admin/service/internal/data/ent/schema/featureoverridespec_jsonfield*.go
git commit -m "feat(thingmodel): add FeatureOverrideSpecField JSON wrapper"
```

---

## Task 8: Ent Schema · CategoryDefaultFeature

**Files:**
- Create: `backend/app/admin/service/internal/data/ent/schema/category_default_feature.go`

参考：设计文档 [02 §2.1](../../thingmodel/sheji/模型管理/02-数据模型设计.md)。

- [ ] **Step 1: 写 schema**

```go
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

type CategoryDefaultFeature struct{ ent.Schema }

func (CategoryDefaultFeature) Annotations() []schema.Annotation {
    return []schema.Annotation{
        entsql.Annotation{
            Table: "thingmodel_category_default_features",
            Charset: "utf8mb4", Collation: "utf8mb4_bin",
        },
        entsql.WithComments(true),
        schema.Comment("分类默认模型条目 / Category default model entry"),
    }
}

func (CategoryDefaultFeature) Fields() []ent.Field {
    return []ent.Field{
        field.Uint32("category_id").
            Comment("分类 ID（必须 level=4）/ Category id, must be level=4"),
        field.Uint32("feature_id").
            Comment("全局特征 ID / Global feature id"),
        field.JSON("override_spec", &FeatureOverrideSpecField{}).
            Optional().Nillable().
            Comment("稀疏覆写 / Sparse override spec"),
        field.String("display_name").MaxLen(128).Optional().Nillable().
            Comment("分类内展示别名 / Display alias within category"),
    }
}

func (CategoryDefaultFeature) Mixin() []ent.Mixin {
    return []ent.Mixin{
        mixin.AutoIncrementId{}, mixin.TimeAt{}, mixin.OperatorID{},
        mixin.IsEnabled{}, mixin.SortOrder{}, mixin.TenantID[uint32]{},
    }
}

func (CategoryDefaultFeature) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("category", Category.Type).
            Ref("default_features").
            Field("category_id").
            Required().Unique().
            Annotations(entsql.Annotation{OnDelete: entsql.Restrict}),
        edge.From("feature", Feature.Type).
            Ref("category_default_entries").
            Field("feature_id").
            Required().Unique().
            Annotations(entsql.Annotation{OnDelete: entsql.Restrict}),
    }
}

func (CategoryDefaultFeature) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("tenant_id", "category_id", "feature_id").Unique().
            StorageKey("uix_tm_cat_default_features_tenant_cat_feat"),
        index.Fields("tenant_id", "category_id", "sort_order").
            StorageKey("idx_tm_cat_default_features_tenant_cat_sort"),
        index.Fields("feature_id").StorageKey("idx_tm_cat_default_features_feature"),
        index.Fields("is_enabled").StorageKey("idx_tm_cat_default_features_enabled"),
        index.Fields("tenant_id").StorageKey("idx_tm_cat_default_features_tenant"),
    }
}
```

- [ ] **Step 2: Commit（暂时无法 `make ent`，留到 Task 11 一起跑）**

```bash
git add backend/app/admin/service/internal/data/ent/schema/category_default_feature.go
git commit -m "feat(thingmodel): add CategoryDefaultFeature ent schema"
```

---

## Task 9: Ent Schema · Product

**Files:**
- Create: `backend/app/admin/service/internal/data/ent/schema/product.go`

- [ ] **Step 1: 写 schema**

```go
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

type Product struct{ ent.Schema }

func (Product) Annotations() []schema.Annotation {
    return []schema.Annotation{
        entsql.Annotation{
            Table: "thingmodel_products",
            Charset: "utf8mb4", Collation: "utf8mb4_bin",
        },
        entsql.WithComments(true),
        schema.Comment("物模型-产品表 / Thing model product"),
    }
}

func (Product) Fields() []ent.Field {
    return []ent.Field{
        field.String("code").MaxLen(64).NotEmpty().Immutable().Optional().Nillable().
            Comment("产品编码（程序标识符，租户内唯一，不可变）/ Product code"),
        field.String("name").MaxLen(128).NotEmpty().Optional().Nillable().
            Comment("产品中文名 / Name (zh)"),
        field.String("name_en").MaxLen(128).Optional().Nillable().
            Comment("产品英文名 / Name (en)"),
        field.Uint32("category_id").Immutable().
            Comment("分类 ID（必须 level=4，不可变）/ Category id"),
        field.String("manufacturer").MaxLen(128).Optional().Nillable().
            Comment("制造商 / Manufacturer"),
        field.String("model_no").MaxLen(64).Optional().Nillable().
            Comment("型号 / Model number"),
        field.String("icon").Optional().Nillable().Comment("图标 / Icon"),
        field.String("description").Optional().Nillable().Comment("描述 / Description"),

        field.Enum("status").
            NamedValues("Draft", "DRAFT", "Published", "PUBLISHED").
            Default("DRAFT").
            Comment("发布状态 / Lifecycle status"),

        field.Uint32("reference_count").Default(0).Nillable().
            Comment("被设备实例引用次数（预留）/ Reference count (reserved)"),
    }
}

func (Product) Mixin() []ent.Mixin {
    return []ent.Mixin{
        mixin.AutoIncrementId{}, mixin.TimeAt{}, mixin.OperatorID{},
        mixin.IsEnabled{}, mixin.SortOrder{}, mixin.TenantID[uint32]{},
    }
}

func (Product) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("category", Category.Type).
            Ref("products").
            Field("category_id").
            Required().Unique().
            Annotations(entsql.Annotation{OnDelete: entsql.Restrict}),
        edge.To("features", ProductFeature.Type).
            Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
    }
}

func (Product) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("tenant_id", "code").Unique().
            StorageKey("uix_tm_products_tenant_code"),
        index.Fields("tenant_id", "category_id", "name").Unique().
            StorageKey("uix_tm_products_tenant_cat_name"),
        index.Fields("tenant_id", "category_id").StorageKey("idx_tm_products_tenant_cat"),
        index.Fields("tenant_id", "status").StorageKey("idx_tm_products_tenant_status"),
        index.Fields("manufacturer").StorageKey("idx_tm_products_manufacturer"),
        index.Fields("is_enabled").StorageKey("idx_tm_products_enabled"),
        index.Fields("tenant_id").StorageKey("idx_tm_products_tenant"),
    }
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/app/admin/service/internal/data/ent/schema/product.go
git commit -m "feat(thingmodel): add Product ent schema"
```

---

## Task 10: Ent Schema · ProductFeature

**Files:**
- Create: `backend/app/admin/service/internal/data/ent/schema/product_feature.go`

- [ ] **Step 1: 写 schema（含 FeatureSpecField/FeatureOverrideSpecField 引用 + 冗余特化列）**

参考设计文档 [02 §2.3](../../thingmodel/sheji/模型管理/02-数据模型设计.md) 完整骨架：

```go
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

type ProductFeature struct{ ent.Schema }

func (ProductFeature) Annotations() []schema.Annotation {
    return []schema.Annotation{
        entsql.Annotation{
            Table: "thingmodel_product_features",
            Charset: "utf8mb4", Collation: "utf8mb4_bin",
        },
        entsql.WithComments(true),
        schema.Comment("产品下特征条目 / Product feature entry"),
    }
}

func (ProductFeature) Fields() []ent.Field {
    return []ent.Field{
        field.Uint32("product_id").Comment("产品 ID / Product id"),

        field.Enum("source").
            NamedValues("Default", "DEFAULT", "Global", "GLOBAL", "Local", "LOCAL").
            Comment("来源 / Source"),

        field.Uint32("ref_feature_id").Optional().Nillable().
            Comment("引用全局特征 ID（LOCAL 时为空）/ Referenced feature id"),

        field.Enum("feature_type").
            NamedValues(
                "Property", "PROPERTY", "Event", "EVENT",
                "Service", "SERVICE", "Relation", "RELATION",
            ).
            Comment("特征类型（冗余列）/ Feature type"),

        field.String("code").MaxLen(128).NotEmpty().Optional().Nillable().
            Comment("产品内编码 / Code within product"),
        field.String("identifier").MaxLen(128).NotEmpty().Optional().Nillable().
            Comment("产品内程序标识符 / Identifier within product"),
        field.String("name").MaxLen(128).NotEmpty().Optional().Nillable().
            Comment("名称 / Name (zh)"),
        field.String("name_en").MaxLen(128).Optional().Nillable().
            Comment("英文名 / Name (en)"),
        field.String("description").Optional().Nillable().
            Comment("描述 / Description"),

        field.JSON("feature_snapshot", &FeatureSpecField{}).
            Comment("完整 spec 快照 / Full feature spec snapshot"),
        field.JSON("override_spec", &FeatureOverrideSpecField{}).
            Optional().Nillable().
            Comment("稀疏覆写 / Sparse override"),

        field.Enum("data_type").
            NamedValues(
                "Int", "INT", "Float", "FLOAT", "Double", "DOUBLE",
                "Bool", "BOOL", "Enum", "ENUM", "Text", "TEXT",
                "Date", "DATE", "Struct", "STRUCT", "Array", "ARRAY",
            ).Optional().Nillable().Comment("property 数据类型"),
        field.Enum("access_mode").
            NamedValues("R", "R", "RW", "RW").Optional().Nillable().
            Comment("property 访问模式"),
        field.Enum("event_level").
            NamedValues("Info", "INFO", "Alert", "ALERT", "Error", "ERROR").
            Optional().Nillable().Comment("event 级别"),
        field.Enum("call_mode").
            NamedValues("Async", "ASYNC", "Sync", "SYNC").Optional().Nillable().
            Comment("service 调用模式"),
        field.String("relation_type").Optional().Nillable().
            Comment("relation 关系类型"),
    }
}

func (ProductFeature) Mixin() []ent.Mixin {
    return []ent.Mixin{
        mixin.AutoIncrementId{}, mixin.TimeAt{}, mixin.OperatorID{},
        mixin.IsEnabled{}, mixin.SortOrder{}, mixin.TenantID[uint32]{},
    }
}

func (ProductFeature) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("product", Product.Type).
            Ref("features").
            Field("product_id").
            Required().Unique().
            Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
    }
}

func (ProductFeature) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("product_id", "code").Unique().StorageKey("uix_tm_pf_product_code"),
        index.Fields("product_id", "identifier").Unique().StorageKey("uix_tm_pf_product_identifier"),
        index.Fields("product_id", "feature_type", "sort_order").StorageKey("idx_tm_pf_product_type_sort"),
        index.Fields("ref_feature_id").StorageKey("idx_tm_pf_ref_feature"),
        index.Fields("source").StorageKey("idx_tm_pf_source"),
        index.Fields("tenant_id", "product_id").StorageKey("idx_tm_pf_tenant_product"),
        index.Fields("feature_type", "data_type").StorageKey("idx_tm_pf_type_datatype"),
    }
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/app/admin/service/internal/data/ent/schema/product_feature.go
git commit -m "feat(thingmodel): add ProductFeature ent schema"
```

---

## Task 11: 修改 Category / Feature 反向 edge + 生成 ent

**Files:**
- Modify: `backend/app/admin/service/internal/data/ent/schema/category.go`
- Modify: `backend/app/admin/service/internal/data/ent/schema/feature.go`

- [ ] **Step 1: Category.Edges 追加反向 edge**

在现有 Edges 中追加：

```go
edge.To("default_features", CategoryDefaultFeature.Type).
    Annotations(entsql.Annotation{OnDelete: entsql.Restrict}),
edge.To("products", Product.Type).
    Annotations(entsql.Annotation{OnDelete: entsql.Restrict}),
```

- [ ] **Step 2: Feature.Edges 追加反向 edge**

```go
edge.To("category_default_entries", CategoryDefaultFeature.Type).
    Annotations(entsql.Annotation{OnDelete: entsql.Restrict}),
```

- [ ] **Step 3: 运行 make ent**

Run: `cd backend && make ent`
Expected: `internal/data/ent/` 下新出 `category_default_feature.go` / `product.go` / `product_feature.go` 等；编译无错

- [ ] **Step 4: 运行 schema migration（本地 PostgreSQL）**

Run: `cd backend/app/admin/service && go run ./cmd/server -c ./configs` 起一次（自动 migrate）；或专用 migrate 命令
Expected: PostgreSQL 出现 `thingmodel_products` / `thingmodel_product_features` / `thingmodel_category_default_features` 三张表 + 全部索引

- [ ] **Step 5: 编译检查**

Run: `cd backend && go build ./...`
Expected: 无错

- [ ] **Step 6: Commit**

```bash
git add backend/app/admin/service/internal/data/ent/
git commit -m "feat(thingmodel): regenerate ent with model management schemas & reverse edges"
```

---

## Task 12: Repo · CategoryDefaultFeature

**Files:**
- Create: `backend/app/admin/service/internal/data/category_default_feature_repo.go`
- Create: `backend/app/admin/service/internal/data/category_default_feature_repo_test.go`

参考：现有 `feature_repo.go` 范式（同目录）。

- [ ] **Step 1: 写 Repo round-trip 测试**

```go
// category_default_feature_repo_test.go (主要测试 JSON 字段 round-trip)
func TestCategoryDefaultFeatureRepo_CreateAndGet_RoundTrip(t *testing.T) {
    repo, ctx := setupTestRepo(t)  // 复用既有测试 helper

    override := &thingmodelV1.FeatureOverrideSpec{
        Constraints: &thingmodelV1.ValueConstraints{Min: 4, Max: 15},
    }
    bo := &CategoryDefaultFeatureBO{
        CategoryId: 1, FeatureId: 1,
        OverrideSpec: override,
        DisplayName:  "出水温度",
    }
    got, err := repo.Create(ctx, bo)
    require.NoError(t, err)

    got2, err := repo.Get(ctx, got.Id)
    require.NoError(t, err)
    require.Equal(t, float64(15), got2.OverrideSpec.GetConstraints().GetMax())
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./app/admin/service/internal/data/ -run TestCategoryDefaultFeatureRepo -v`
Expected: FAIL（类型未定义）

- [ ] **Step 3: 实现 Repo（含 copier TypeConverter 注册）**

要点：
- 方法：`Create / Get / List / Update / Delete / BatchCreate / Reorder / Exists / IncSortOrder`
- copier 注册 `*schema.FeatureOverrideSpecField` ↔ `*thingmodelV1.FeatureOverrideSpec` 双向 TypeConverter
- 列表查询支持 `category_id` / `feature_type`（join feature）/ `is_enabled` 过滤
- 写入时校验 `category.level == 4`：调用 `categoryRepo.Get(ctx, categoryId).Level`

- [ ] **Step 4: 运行测试通过**

Run: `cd backend && go test ./app/admin/service/internal/data/ -run TestCategoryDefaultFeatureRepo -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/app/admin/service/internal/data/category_default_feature_repo*.go
git commit -m "feat(thingmodel): add CategoryDefaultFeature repo with JSON round-trip"
```

---

## Task 13: Repo · Product

**Files:**
- Create: `backend/app/admin/service/internal/data/product_repo.go`
- Create: `backend/app/admin/service/internal/data/product_repo_test.go`

- [ ] **Step 1: 写测试**

```go
func TestProductRepo_UniqueConstraints(t *testing.T) {
    repo, ctx := setupTestRepo(t)
    _, err := repo.Create(ctx, &ProductBO{Code: "X", Name: "x", CategoryId: 1, TenantId: 0})
    require.NoError(t, err)

    // 同 code 应冲突
    _, err = repo.Create(ctx, &ProductBO{Code: "X", Name: "y", CategoryId: 1, TenantId: 0})
    require.Error(t, err)
}

func TestProductRepo_ListWithFilters(t *testing.T) {
    // category_id / status / manufacturer / keyword 过滤
}
```

- [ ] **Step 2: 运行测试确认失败**

- [ ] **Step 3: 实现 Repo**

要点：
- 方法：`Create / Get / GetByCode / List(paged) / Update(mask) / Delete / UpdateStatus / IncRefCount / DecRefCount / Exists`
- `Get` 返回时 join Category 填充 category_code / category_name
- `List` 支持 `category_id` / `status` / `manufacturer` / `keyword`（name+code+model_no LIKE） / `is_enabled`；可选 `include_descendants` 在 service 层展开为 code LIKE 'xxx%' 后再传 IDs

- [ ] **Step 4: 运行测试通过**

- [ ] **Step 5: Commit**

```bash
git add backend/app/admin/service/internal/data/product_repo*.go
git commit -m "feat(thingmodel): add Product repo with unique constraints & filters"
```

---

## Task 14: Repo · ProductFeature

**Files:**
- Create: `backend/app/admin/service/internal/data/product_feature_repo.go`
- Create: `backend/app/admin/service/internal/data/product_feature_repo_test.go`

- [ ] **Step 1: 写测试**

```go
func TestProductFeatureRepo_JSONRoundTrip(t *testing.T) {
    // 写入含 oneof PropertySpec 的 snapshot，读回应 == 原 spec
}

func TestProductFeatureRepo_BatchCreate_TransactionRollback(t *testing.T) {
    // 第二条触发唯一冲突时，整批应回滚
}

func TestProductFeatureRepo_ListByProductFeatureType(t *testing.T) {
    // 按 (product_id, feature_type) 过滤、按 sort_order 排序
}
```

- [ ] **Step 2: 运行测试确认失败**

- [ ] **Step 3: 实现 Repo**

要点：
- 方法：`Create / Get / List(filters) / BatchCreate / Update(mask) / Delete / DeleteByProductId / ListByProductFeatureIds / ExistsByProductIdAndCode`
- copier 注册两个 wrapper TypeConverter
- BatchCreate 走 `client.ProductFeature.CreateBulk` 单事务

- [ ] **Step 4: 运行测试通过**

- [ ] **Step 5: Commit**

```bash
git add backend/app/admin/service/internal/data/product_feature_repo*.go
git commit -m "feat(thingmodel): add ProductFeature repo with JSON round-trip"
```

---

## Task 15: Service · Override 校验器 + EffectiveSpec

**Files:**
- Create: `backend/app/admin/service/internal/service/product_validator.go`
- Create: `backend/app/admin/service/internal/service/product_validator_test.go`

- [ ] **Step 1: 写测试**

```go
func TestValidateOverrideSpec_ConstraintsRange(t *testing.T) {
    v := NewProductValidator(nil)  // unitRepo 可空，本用例不涉及
    propertyFeat := &thingmodelV1.FeatureSpec{
        Spec: &thingmodelV1.FeatureSpec_Property{Property: &thingmodelV1.PropertySpec{
            DataType: thingmodelV1.DataType_DOUBLE,
        }},
    }
    over := &thingmodelV1.FeatureOverrideSpec{
        Constraints: &thingmodelV1.ValueConstraints{Min: 10, Max: 5},
    }
    err := v.ValidateOverrideSpec(propertyFeat, over)
    require.Error(t, err, "max<min should fail")
}

func TestValidateOverrideSpec_NonPropertyRejectsConstraints(t *testing.T) {
    eventFeat := &thingmodelV1.FeatureSpec{
        Spec: &thingmodelV1.FeatureSpec_Event{Event: &thingmodelV1.EventSpec{}},
    }
    over := &thingmodelV1.FeatureOverrideSpec{
        Constraints: &thingmodelV1.ValueConstraints{Min: 0, Max: 10},
    }
    err := NewProductValidator(nil).ValidateOverrideSpec(eventFeat, over)
    require.Error(t, err)
}

func TestEffectiveSpec_MergesProperty(t *testing.T) {
    snap := buildPropertySpec(0, 100)
    over := &thingmodelV1.FeatureOverrideSpec{
        Constraints: &thingmodelV1.ValueConstraints{Min: 10, Max: 50},
    }
    eff := EffectiveSpec(snap, over)
    require.Equal(t, float64(10), eff.GetProperty().GetConstraints().GetMin())
}
```

- [ ] **Step 2: 运行测试确认失败**

- [ ] **Step 3: 实现校验器与合并工具**

参考设计文档 [04 §3 + §5](../../thingmodel/sheji/模型管理/04-后端实现设计.md)：

```go
// product_validator.go
package service

import (
    "context"

    "google.golang.org/protobuf/proto"

    thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

type ProductValidator struct {
    unitRepo UnitRepoMin   // 仅需 Exists(id)
}
func NewProductValidator(unitRepo UnitRepoMin) *ProductValidator { return &ProductValidator{unitRepo} }

func (v *ProductValidator) ValidateOverrideSpec(target *thingmodelV1.FeatureSpec, over *thingmodelV1.FeatureOverrideSpec) error {
    if over == nil { return nil }
    p := target.GetProperty()
    if p == nil {
        if over.Constraints != nil || over.Unit != nil || over.DefaultValue != nil {
            return errorOverrideInvalid("non-property cannot override constraints/unit/defaultValue")
        }
        return nil
    }
    if over.Constraints != nil && over.Constraints.Max < over.Constraints.Min {
        return errorOverrideInvalid("constraints.max < min")
    }
    if over.Unit != nil && over.Unit.UnitId > 0 && v.unitRepo != nil {
        if exists, _ := v.unitRepo.Exists(context.TODO(), over.Unit.UnitId); !exists {
            return errorOverrideInvalid("unit not found")
        }
    }
    return nil
}

// EffectiveSpec = snapshot deepmerged with override
func EffectiveSpec(snap *thingmodelV1.FeatureSpec, over *thingmodelV1.FeatureOverrideSpec) *thingmodelV1.FeatureSpec {
    if over == nil { return snap }
    out := proto.Clone(snap).(*thingmodelV1.FeatureSpec)
    if p := out.GetProperty(); p != nil {
        if over.Constraints != nil  { p.Constraints = over.Constraints }
        if over.Unit != nil         { p.Unit = over.Unit }
        if over.DefaultValue != nil { p.DefaultValue = over.DefaultValue }
    }
    return out
}
```

- [ ] **Step 4: 运行测试通过**

- [ ] **Step 5: Commit**

```bash
git add backend/app/admin/service/internal/service/product_validator*.go
git commit -m "feat(thingmodel): add override validator & effective spec merger"
```

---

## Task 16: Service · CategoryDefaultFeature handler

**Files:**
- Create: `backend/app/admin/service/internal/service/category_default_feature_service.go`
- Create: `backend/app/admin/service/internal/service/category_default_feature_service_test.go`

- [ ] **Step 1: 写测试覆盖关键路径**

```go
func TestCreateCategoryDefaultFeature_RequiresLevel4(t *testing.T) {
    svc, _ := setupService(t)
    // 用 level=3 的 category 创建
    _, err := svc.CreateCategoryDefaultFeature(ctx, &adminV1.CreateCategoryDefaultFeatureRequest{
        Data: &adminV1.CreateCategoryDefaultFeatureRequest_Data{
            CategoryId: levelThreeCategoryId, FeatureId: existingFeatureId,
        },
    })
    require.ErrorIs(t, err, thingmodelV1.ErrorTmCatDefaultFeatureCategoryNotLeaf(""))
}

func TestCreateCategoryDefaultFeature_IncRefCount(t *testing.T) {
    // create 后 feature.reference_count 应 +1
}

func TestDeleteCategoryDefaultFeature_DecRefCount(t *testing.T) {
    // delete 后应 -1
}

func TestBatchAdd_SkipsDuplicates(t *testing.T) {
    // 同 (category, feature) 已存在的项应在 skipped_duplicate_feature_codes 返回
}
```

- [ ] **Step 2: 运行测试确认失败**

- [ ] **Step 3: 实现 service handler**

要点：
- 实现 7 个 RPC（List/Get/Create/BatchAdd/Update/Delete/Reorder）
- 每个写入操作开启事务（`data.Tx(ctx, func(tx) error {...})`）
- Create 流程：校验 level=4 → 校验 feature.is_enabled → 校验 override_spec → 创建行 → feature.reference_count +1 → unit.reference_count +1（若 override 或 snapshot 含 unitId）
- Delete 流程：取行 → 删行 → feature ref -1 → unit ref -1
- Update（FieldMask）：override.unit 变更 → 旧 unit -1 / 新 unit +1
- BatchAdd：循环 Create 收集 created / skipped；遇唯一冲突归入 skipped

- [ ] **Step 4: 运行测试通过**

- [ ] **Step 5: Commit**

```bash
git add backend/app/admin/service/internal/service/category_default_feature_service*.go
git commit -m "feat(thingmodel): impl CategoryDefaultFeature service with ref_count maintenance"
```

---

## Task 17: Service · Product handler

**Files:**
- Create: `backend/app/admin/service/internal/service/product_service.go`
- Create: `backend/app/admin/service/internal/service/product_service_test.go`

- [ ] **Step 1: 写测试**

```go
func TestCreateProduct_RequiresLevel4Category(t *testing.T) {}
func TestCreateProduct_CodeDuplicate(t *testing.T) {}
func TestCreateProduct_NameDuplicateInSameCategory(t *testing.T) {}
func TestPublishProduct_Idempotent(t *testing.T) {}
func TestDeleteProduct_BlockedByRefCount(t *testing.T) {
    // 手动 inc product.reference_count 后再 Delete 应报错
}
```

- [ ] **Step 2: 运行测试确认失败**

- [ ] **Step 3: 实现 service handler（7 RPC）**

要点：
- Create：校验 category.level==4；status 默认 DRAFT；DB 唯一冲突映射为 TM_PRODUCT_CODE_DUPLICATE / NAME_DUPLICATE
- Update（FieldMask）：禁止改 code/category_id（Ent 已 Immutable 兜底）
- Delete：reference_count > 0 报错
- Publish/Unpublish：仅切 status；Publish 后由 ProductFeature service 校验冻结

- [ ] **Step 4: 运行测试通过**

- [ ] **Step 5: Commit**

```bash
git add backend/app/admin/service/internal/service/product_service*.go
git commit -m "feat(thingmodel): impl Product service with status lifecycle"
```

---

## Task 18: Service · ProductFeature handler（含 PullFromDefault）

**Files:**
- Create: `backend/app/admin/service/internal/service/product_feature_service.go`
- Create: `backend/app/admin/service/internal/service/product_feature_service_test.go`

参考：设计文档 [04 §4](../../thingmodel/sheji/模型管理/04-后端实现设计.md)。

- [ ] **Step 1: 写 PullFromDefault 测试（关键路径）**

```go
func TestPullFromDefault_AllFeaturesCopied(t *testing.T) {
    // category 配 3 个 default_features → 新建 product → PullFromDefault({}) 全部拉取
    // 期望 product_features 出现 3 条 source=DEFAULT；snapshot 等于全局 spec merge default override
}

func TestPullFromDefault_SkipConflict(t *testing.T) {
    // 先 Pull 一次产生 1 条 → 再 Pull 同 default_id（SKIP）应跳过、skipped 数组含此项
}

func TestPullFromDefault_ReplaceConflict(t *testing.T) {
    // REPLACE：先删旧 → 重建；原 override_spec 丢失
}

func TestPullFromDefault_PublishedProductRejected(t *testing.T) {
    // product.status=PUBLISHED → 拒绝
}

func TestPullFromDefault_RefCountUnchanged(t *testing.T) {
    // source=DEFAULT 拉取后 thing_features.reference_count 不变
}

func TestCreateLocalFeature_RefCountUntouched(t *testing.T) {
    // source=LOCAL 不动 thing_features.reference_count
}

func TestUpdate_PublishedProductFreezesStructure(t *testing.T) {
    // status=PUBLISHED 时改 override_spec 允许；改 code/identifier/dataType 拒绝
}
```

- [ ] **Step 2: 运行测试确认失败**

- [ ] **Step 3: 实现 service handler**

实现 9 个 RPC：List/Get/Create/PullFromDefault/CloneFromProduct/Update/Delete/Reorder。

PullFromDefault 核心逻辑（参考设计文档 [04 §4](../../thingmodel/sheji/模型管理/04-后端实现设计.md) 完整代码）：
```go
func (s *ProductFeatureService) PullFromDefault(ctx, req) (resp, error) {
    // 1. Get product, check status != PUBLISHED
    // 2. List default_features by category_id + filter ids
    // 3. List existing product_features by feature_ids
    // 4. Tx:
    //    for each cdf:
    //      if duplicate:
    //        SKIP → append skipped; continue
    //        REPLACE → delete existing first
    //      feature := featureRepo.Get(cdf.feature_id)
    //      if !feature.is_enabled: skip; continue
    //      snapshot := EffectiveSpec(feature.spec, cdf.override_spec)  // 合并 default 层
    //      Create product_feature{
    //        source=DEFAULT, ref_feature_id=feature.id,
    //        feature_type, code, identifier, name, ...
    //        feature_snapshot=snapshot, override_spec=nil,
    //        denormalized: data_type/access_mode/event_level/call_mode/relation_type,
    //      }
    //      // 注意：source=DEFAULT 不动 reference_count
    // 5. Return created + skipped
}
```

Create（POST .../features）的三类来源分支：
- source=GLOBAL：require ref_feature_id；从 feature 拷 snapshot；feature.reference_count +1；unit.reference_count +1
- source=LOCAL：require feature_snapshot in request；不动 reference_count
- source=DEFAULT：拒绝（仅 PullFromDefault 创建此来源）

Update 的"PUBLISHED 冻结"检查：通过 FieldMask 看哪些 paths 在请求里，结构字段（feature_type/code/identifier/dataType/feature_snapshot 非 override 部分）禁改。

- [ ] **Step 4: 运行测试通过**

Run: `cd backend && go test ./app/admin/service/internal/service/ -run TestPullFromDefault -v`
Expected: PASS（5 个 PullFromDefault 用例 + 其它）

- [ ] **Step 5: Commit**

```bash
git add backend/app/admin/service/internal/service/product_feature_service*.go
git commit -m "feat(thingmodel): impl ProductFeature service with PullFromDefault batch"
```

---

## Task 19: Wire + Server 注册

**Files:**
- Modify: `backend/app/admin/service/internal/data/data.go`
- Modify: `backend/app/admin/service/internal/service/service.go`
- Modify: `backend/app/admin/service/internal/server/grpc.go`
- Modify: `backend/app/admin/service/internal/server/http.go`

- [ ] **Step 1: data.ProviderSet 追加 3 个 repo**

```go
var ProviderSet = wire.NewSet(
    // ... existing ...
    NewCategoryDefaultFeatureRepo,
    NewProductRepo,
    NewProductFeatureRepo,
)
```

- [ ] **Step 2: service.ProviderSet 追加 3 个 service + validator**

```go
var ProviderSet = wire.NewSet(
    // ... existing ...
    NewProductValidator,
    NewCategoryDefaultFeatureService,
    NewProductService,
    NewProductFeatureService,
)
```

- [ ] **Step 3: server/grpc.go + http.go 注册**

在 NewGRPCServer / NewHTTPServer 里追加 `RegisterCategoryDefaultFeatureServiceServer(srv, svc)`、`RegisterProductServiceServer`、`RegisterProductFeatureServiceServer`，并把 service 加进 `NewXxxServer` 的入参。

- [ ] **Step 4: 跑 make wire**

Run: `cd backend && make wire`
Expected: `wire_gen.go` 更新；编译无错

- [ ] **Step 5: 起服务验证**

Run: `cd backend/app/admin/service && make run`
Expected: 启动成功；`http://localhost:7788/docs/` 可见 12+ 个新接口

- [ ] **Step 6: Commit**

```bash
git add backend/app/admin/service/internal/{data,service,server}/*.go \
        backend/app/admin/service/internal/**/wire_gen.go
git commit -m "feat(thingmodel): wire model management services into admin server"
```

---

## Task 20: 菜单 + 权限码

**Files:**
- Modify: `backend/pkg/constants/default_data.go`

- [ ] **Step 1: DefaultMenus 追加产品管理菜单**

定位 `DefaultMenus` 数组中 thingmodel 父菜单的子项后追加：

```go
{
    Path: "/thingmodel/product",
    Name: "ThingModelProduct",
    Title: "产品管理",
    TitleI18n: "menu.thingmodel.product",
    Icon: "carbon:product",
    ParentPath: "/thingmodel",
    Authority: []string{"sys:platform_admin", "sys:tenant_manager"},
    Sort: 40,
},
```

> 注意：分类默认模型不单独建菜单，复用 `thingmodel:category:edit` 权限码。

- [ ] **Step 2: 启动服务，验证 SyncPermissions 派生权限码**

Run: `cd backend/app/admin/service && make run`
Expected：日志中可见 `thingmodel:product:view/create/edit/delete` 4 个权限码插入；菜单挂到 `sys:platform_admin` / `sys:tenant_manager`

- [ ] **Step 3: Commit**

```bash
git add backend/pkg/constants/default_data.go
git commit -m "feat(thingmodel): add 产品管理 menu entry & permission codes"
```

---

## Task 21: 种子 · 暖通冷水机组最小演示链路

**Files:**
- Create: `backend/app/admin/service/internal/data/seed/model_management_seed.go`
- Create: `backend/app/admin/service/internal/data/seed/model_management_seed_data.go`
- Create: `backend/app/admin/service/internal/data/seed/model_management_seed_test.go`

参考：设计文档 [06 §2 + §3](../../thingmodel/sheji/模型管理/06-种子数据与实施计划.md)。

- [ ] **Step 1: 写种子数据常量（10 特征 codes + 2 overrides + 1 产品定义）**

```go
// model_management_seed_data.go
package seed

var modelMgmtFeatureCodes = []string{
    "P-RUN-0001","P-RUN-0002","P-MEAS-0010","P-MEAS-0011","P-MEAS-0030",
    "P-RATED-0050","E-FAULT-0001","E-GEN-0001","S-CTRL-0001","R-PART-0001",
}

var modelMgmtOverrides = map[string]*thingmodelV1.FeatureOverrideSpec{
    "P-MEAS-0010": {Constraints: &thingmodelV1.ValueConstraints{Min: 4, Max: 15}},
    "P-MEAS-0011": {Constraints: &thingmodelV1.ValueConstraints{Min: 6, Max: 18}},
}

var modelMgmtSkipDefaultPull = map[string]bool{
    "P-MEAS-0030": true, "R-PART-0001": true,
}

var modelMgmtProductOverrides = map[string]*thingmodelV1.FeatureOverrideSpec{
    "P-MEAS-0010": {Constraints: &thingmodelV1.ValueConstraints{Min: 5, Max: 12}},
}
```

- [ ] **Step 2: 写 idempotent 测试**

```go
func TestModelManagementSeed_Idempotent(t *testing.T) {
    s := setupSeeder(t)
    require.NoError(t, s.Run(ctx))
    cnt1 := countAllSeeded(t)
    require.NoError(t, s.Run(ctx))   // 第二次跑
    cnt2 := countAllSeeded(t)
    require.Equal(t, cnt1, cnt2, "seed should be idempotent")
}

func TestModelManagementSeed_RefCountCorrect(t *testing.T) {
    // 跑完种子后，thing_features.reference_count
    // = 默认条目数（10）+ 该产品额外 source=GLOBAL 数（0）
}
```

- [ ] **Step 3: 运行测试确认失败**

- [ ] **Step 4: 实现 Seeder**

参考设计文档 [06 §3.2](../../thingmodel/sheji/模型管理/06-种子数据与实施计划.md) 完整骨架。要点：
- Run(ctx)：查 category 20010100 → 查 10 个 features → upsert 10 条 default_feature → upsert 产品 → upsert 8 条 DEFAULT product_features + 1 条 LOCAL `L-NIGHT-MUTE`
- 全部 `OnConflict(...).Ignore()` 保证 idempotent

- [ ] **Step 5: 注册到 seed runner（与现有 unit_seed / feature_seed / category_seed 同级）**

```go
// internal/data/seed/runner.go（或现有 seed 注册入口）
seeders := []Seeder{
    // ... existing
    NewModelManagementSeeder(c),  // 必须在 unit/feature/category 之后
}
```

- [ ] **Step 6: 运行 make run 自动跑种子**

Expected：DB 出现 10 条 category_default_features + 1 个 product + 9 条 product_features

- [ ] **Step 7: 运行集成测试通过**

Run: `cd backend && go test ./app/admin/service/internal/data/seed/ -run TestModelManagementSeed -v`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add backend/app/admin/service/internal/data/seed/model_management*.go
git commit -m "feat(thingmodel): add seed for chiller model management demo link"
```

---

## Task 22: 端到端集成测试 · API 级

**Files:**
- Create: `backend/app/admin/service/internal/service/model_management_e2e_test.go`

- [ ] **Step 1: 写场景化 e2e 测试**

```go
func TestE2E_FullFlow(t *testing.T) {
    ts, cli := setupHTTPServer(t)
    defer ts.Close()

    // 1. 在某 level=4 分类配置默认模型（BatchAdd 3 个特征）
    cdfs := cli.BatchAddDefaultFeatures(catId, []uint32{f1, f2, f3})
    require.Len(t, cdfs, 3)

    // 2. 创建产品
    p := cli.CreateProduct(adminV1.CreateProductRequest{Code:"E2E-001", Name:"测试产品", CategoryId: catId})

    // 3. PullFromDefault（全部）
    pulled := cli.PullFromDefault(p.Id, nil)
    require.Len(t, pulled.Created, 3)

    // 4. 加一条 GLOBAL feature；feature.reference_count 应 +1
    initRef := cli.GetFeature(otherFeatureId).ReferenceCount
    cli.CreateProductFeature(p.Id, adminV1.CreateProductFeatureRequest{
        Source: thingmodelV1.ProductFeatureSource_GLOBAL,
        RefFeatureId: otherFeatureId,
    })
    require.Equal(t, initRef+1, cli.GetFeature(otherFeatureId).ReferenceCount)

    // 5. 加一条 LOCAL；reference_count 不变
    cli.CreateProductFeature(p.Id, adminV1.CreateProductFeatureRequest{
        Source: thingmodelV1.ProductFeatureSource_LOCAL,
        Code: "L-TEST", Identifier: "test", Name: "本地测试",
        FeatureType: thingmodelV1.FeatureType_PROPERTY,
        FeatureSnapshot: buildBoolPropertySpec(),
    })

    // 6. Publish 产品
    cli.PublishProduct(p.Id)

    // 7. PUBLISHED 后试图加新特征 → 失败 4206
    _, err := cli.CreateProductFeatureRaw(p.Id, ...)
    require.ErrorContains(t, err, "4206")

    // 8. PUBLISHED 后改 override → 成功
    cli.UpdateProductFeature(...)  // override_spec.constraints

    // 9. 删除产品 → 失败（先去掉 reference_count 阻挡？本期无 device 引用，应直接成功）
    cli.DeleteProduct(p.Id)

    // 10. 验证 thing_features.reference_count 全部归位
    require.Equal(t, initRef, cli.GetFeature(otherFeatureId).ReferenceCount)
}
```

- [ ] **Step 2: 运行测试确认失败**

- [ ] **Step 3: 修补任何回归（多数情况下应直接通过，凡失败者修代码不修测试）**

- [ ] **Step 4: 运行测试通过**

Run: `cd backend && go test ./app/admin/service/internal/service/ -run TestE2E_FullFlow -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/app/admin/service/internal/service/model_management_e2e_test.go
git commit -m "test(thingmodel): add e2e flow test for model management"
```

---

## Task 23: 全量验收 · make test + make lint + 接口巡检

- [ ] **Step 1: make test 全部通过**

Run: `cd backend && make test`
Expected: 全绿（含新加的 25+ 用例）

- [ ] **Step 2: make lint 全部通过**

Run: `cd backend && make lint`
Expected: 无 warning（按已有项目 golangci-lint 配置）

- [ ] **Step 3: 接口巡检 · Swagger UI**

打开 `http://localhost:7788/docs/`，在搜索框输 `thingmodel`，确认以下接口全部可见：

```
分类默认模型 7 个：
  POST /v1/thingmodel/category-default-features
  POST /v1/thingmodel/categories/{category_id}/default-features:batch
  GET  /v1/thingmodel/categories/{category_id}/default-features
  POST /v1/thingmodel/categories/{category_id}/default-features:reorder
  GET  /v1/thingmodel/category-default-features/{id}
  PATCH /v1/thingmodel/category-default-features/{id}
  DELETE /v1/thingmodel/category-default-features/{id}

产品 7 个：
  GET/POST/PATCH/DELETE /v1/thingmodel/products[/{id}]
  POST /v1/thingmodel/products/{id}:publish
  POST /v1/thingmodel/products/{id}:unpublish
  GET /v1/thingmodel/products/{id}

产品特征 8 个：
  GET/POST /v1/thingmodel/products/{product_id}/features
  POST /v1/thingmodel/products/{product_id}/features:pull-from-default
  POST /v1/thingmodel/products/{product_id}/features:clone-from
  POST /v1/thingmodel/products/{product_id}/features:reorder
  GET/PATCH/DELETE /v1/thingmodel/product-features/{id}
```

- [ ] **Step 4: 验收单（设计文档 06 §5.1 AC1~AC15）**

逐条手测（Swagger UI 或 curl）：
- AC1~AC15：参考设计文档 [06 §5.1](../../thingmodel/sheji/模型管理/06-种子数据与实施计划.md)
- 期望全部 ✓

- [ ] **Step 5: 接口性能 sanity（可选）**

`ListProducts` 100 条数据响应 < 200ms（本地环境）
`PullFromDefault` 10 条特征响应 < 500ms（本地环境）

- [ ] **Step 6: 收尾 commit（如有修补）**

```bash
git status   # 应 clean，或 lint 修补
git push origin chdq
```

---

## 完成标准

实施完成 = 满足以下全部：

1. ✅ Tasks 1-23 全部 checkbox 完成
2. ✅ `make test` + `make lint` 全绿
3. ✅ Swagger UI 可见 22 个新接口
4. ✅ 种子跑完后 DB 出现 "电动压缩式冷水机组" 默认模型 10 条 + 示范产品 1 个 + 产品特征 9 条
5. ✅ 设计文档 [06 §5.1 AC1~AC15](../../thingmodel/sheji/模型管理/06-种子数据与实施计划.md) 全部 ✓
6. ✅ `thing_features.reference_count` / `thingmodel_units.reference_count` 在端到端测试中正确增减

---

## 风险与缓解

| 风险 | 缓解 |
|------|------|
| `FeatureSpec` oneof JSON 写入侥幸成功、读取崩 | Task 7/14 强制走 wrapper + round-trip 测试 |
| PullFromDefault 大批量性能差 | Task 14 用 `CreateBulk` 单事务；后续 100+ 默认特征再优化 |
| `reference_count` 漏更新或重复更新 | Task 16/18 在事务内同步更新；e2e 测试验证归零 |
| make wire 没识别新 ProviderSet | Task 19 Step 4 验收依赖；漏注册会编译失败 |
| 种子在分类/特征未就绪时跑导致空查询 | Task 21 Seeder 内 require category/feature 存在，否则跳过并打 WARN |

---

## 不在本计划内（按设计文档"非目标"）

- 产品模型版本号 / 发布历史
- 设备实例 / 数据上报
- 批量 Excel 导入
- 跨租户克隆
- 物模型 JSON 出码
- 前端实现（React / Vue Element / Vue Vben）— 三份独立计划在本计划 Task 23 验收通过后另起
