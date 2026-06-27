# 03 · API 与 Proto 设计

> 本章定义模型管理的 proto 契约（源领域层 + BFF 层）、RPC 清单、批量"拉取默认模型"原子操作、错误码。
> 严格遵循项目"Contract-First Pipeline"（[CLAUDE.md §The Contract-First Pipeline](../../../../CLAUDE.md)）。

---

## 1. proto 文件布局

```
backend/api/protos/
├── thingmodel/service/v1/
│   ├── category_default_feature.proto      ← 源领域层（新增）
│   ├── product.proto                       ← 源领域层（新增）
│   ├── product_feature.proto               ← 源领域层（新增）
│   ├── feature_override_spec.proto         ← 源领域层（新增，覆写 spec 强类型契约）
│   ├── ... (已有 category/feature/unit 等)
│
└── admin/service/v1/
    ├── category_default_feature.proto      ← BFF 层（新增，带 google.api.http）
    ├── product.proto                       ← BFF 层（新增）
    └── product_feature.proto               ← BFF 层（新增）
```

> 与 CLAUDE.md 两层 proto 架构对齐：源领域定义 messages + 完整 gRPC service（无 http 注解），BFF 层定义 REST 表面（带 google.api.http）。

---

## 2. 源领域层 — 共享 messages

### 2.1 `feature_override_spec.proto`

```protobuf
syntax = "proto3";

package thingmodel.service.v1;

import "thingmodel/service/v1/feature.proto";  // 复用 ValueConstraints / UnitRef
import "google/protobuf/struct.proto";
import "google/protobuf/wrappers.proto";

// 特征覆写规约（仅白名单字段）
// 用于：thingmodel_category_default_features.override_spec
//      thingmodel_product_features.override_spec
message FeatureOverrideSpec {
  ValueConstraints     constraints   = 1;  // 收紧/放宽范围
  UnitRef              unit          = 2;  // 覆写单位
  google.protobuf.Value default_value = 3 [json_name="defaultValue"];
  string               display_name  = 4 [json_name="displayName"];
  string               description   = 5;
  google.protobuf.BoolValue required = 6;  // 仅 service.inputParams 子项有效
}
```

### 2.2 `category_default_feature.proto`（源领域）

```protobuf
syntax = "proto3";

package thingmodel.service.v1;

import "thingmodel/service/v1/feature_override_spec.proto";
import "thingmodel/service/v1/feature.proto";   // 复用 FeatureSpec 等
import "google/protobuf/timestamp.proto";

// 单条分类默认模型条目
message CategoryDefaultFeature {
  uint32 id            = 1;
  uint32 category_id   = 2 [json_name="categoryId"];
  uint32 feature_id    = 3 [json_name="featureId"];

  // 关联字段（list/get 时由后端 join 填充）
  string feature_code        = 10 [json_name="featureCode"];        // 全局特征 code
  string feature_identifier  = 11 [json_name="featureIdentifier"];
  string feature_name        = 12 [json_name="featureName"];
  FeatureType feature_type   = 13 [json_name="featureType"];
  FeatureSpec feature_snapshot_preview = 14 [json_name="featureSnapshotPreview"];  // 来自全局特征的 spec 预览

  // 自有字段
  FeatureOverrideSpec override_spec = 20 [json_name="overrideSpec"];
  string display_name  = 21 [json_name="displayName"];
  bool   is_enabled    = 22 [json_name="isEnabled"];
  uint32 sort_order    = 23 [json_name="sortOrder"];
  uint32 tenant_id     = 24 [json_name="tenantId"];

  google.protobuf.Timestamp created_at = 30 [json_name="createdAt"];
  google.protobuf.Timestamp updated_at = 31 [json_name="updatedAt"];
}
```

### 2.3 `product.proto`（源领域）

```protobuf
syntax = "proto3";

package thingmodel.service.v1;

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
  string name_en       = 4 [json_name="nameEn"];
  uint32 category_id   = 5 [json_name="categoryId"];

  // 关联字段（list/get 时填充）
  string category_code = 10 [json_name="categoryCode"];
  string category_name = 11 [json_name="categoryName"];

  string manufacturer  = 20;
  string model_no      = 21 [json_name="modelNo"];
  string icon          = 22;
  string description   = 23;

  ProductStatus status = 30;
  bool   is_enabled    = 31 [json_name="isEnabled"];
  uint32 sort_order    = 32 [json_name="sortOrder"];
  uint32 reference_count = 33 [json_name="referenceCount"];
  uint32 tenant_id     = 34 [json_name="tenantId"];

  // 统计字段（list 时返回，详细计数后端聚合）
  uint32 feature_count = 40 [json_name="featureCount"];

  google.protobuf.Timestamp created_at = 50 [json_name="createdAt"];
  google.protobuf.Timestamp updated_at = 51 [json_name="updatedAt"];
}
```

### 2.4 `product_feature.proto`（源领域）

```protobuf
syntax = "proto3";

package thingmodel.service.v1;

import "thingmodel/service/v1/feature.proto";
import "thingmodel/service/v1/feature_override_spec.proto";
import "google/protobuf/timestamp.proto";

enum ProductFeatureSource {
  PRODUCT_FEATURE_SOURCE_UNSPECIFIED = 0;
  DEFAULT = 1;  // 从分类默认模型拷贝
  GLOBAL  = 2;  // 从全局特征库新加
  LOCAL   = 3;  // 产品本地自定义
}

message ProductFeature {
  uint32 id              = 1;
  uint32 product_id      = 2 [json_name="productId"];
  ProductFeatureSource source = 3;
  uint32 ref_feature_id  = 4 [json_name="refFeatureId"];   // LOCAL 时为 0

  FeatureType feature_type = 10 [json_name="featureType"];
  string code            = 11;
  string identifier      = 12;
  string name            = 13;
  string name_en         = 14 [json_name="nameEn"];
  string description     = 15;

  FeatureSpec          feature_snapshot = 20 [json_name="featureSnapshot"];
  FeatureOverrideSpec  override_spec    = 21 [json_name="overrideSpec"];

  // 冗余特化列（READ ONLY，由后端按 snapshot 同步）
  DataType   data_type    = 30 [json_name="dataType"];
  AccessMode access_mode  = 31 [json_name="accessMode"];
  EventLevel event_level  = 32 [json_name="eventLevel"];
  CallMode   call_mode    = 33 [json_name="callMode"];
  string     relation_type = 34 [json_name="relationType"];

  bool   is_enabled  = 40 [json_name="isEnabled"];
  uint32 sort_order  = 41 [json_name="sortOrder"];
  uint32 tenant_id   = 42 [json_name="tenantId"];

  google.protobuf.Timestamp created_at = 50 [json_name="createdAt"];
  google.protobuf.Timestamp updated_at = 51 [json_name="updatedAt"];
}
```

---

## 3. BFF 层 — REST 表面

### 3.1 `admin/service/v1/category_default_feature.proto`

| RPC | HTTP | 用途 |
|-----|------|------|
| `ListCategoryDefaultFeatures` | `GET /v1/thingmodel/categories/{category_id}/default-features` | 取某 level=4 分类的默认模型条目（默认按 sort_order） |
| `GetCategoryDefaultFeature`   | `GET /v1/thingmodel/category-default-features/{id}` | 取单条 |
| `CreateCategoryDefaultFeature`| `POST /v1/thingmodel/category-default-features` | 新增一条（手动加而非批量） |
| `BatchAddCategoryDefaultFeatures` | `POST /v1/thingmodel/categories/{category_id}/default-features:batch` | 批量加多条（场景：分类管理 Drawer 内一次选 N 个特征） |
| `UpdateCategoryDefaultFeature` | `PATCH /v1/thingmodel/category-default-features/{id}` | 更新 override_spec / display_name / sort / enabled，走 FieldMask |
| `DeleteCategoryDefaultFeature` | `DELETE /v1/thingmodel/category-default-features/{id}` | 删除单条；DB 层 RESTRICT 保护 |
| `ReorderCategoryDefaultFeatures` | `POST /v1/thingmodel/categories/{category_id}/default-features:reorder` | 拖拽排序 |

#### 关键 request/response 形态

```protobuf
// ListCategoryDefaultFeatures
message ListCategoryDefaultFeaturesRequest {
  uint32 category_id = 1 [json_name="categoryId"];
  optional thingmodel.service.v1.FeatureType feature_type = 2 [json_name="featureType"];  // 按类型筛选
  optional bool is_enabled = 3 [json_name="isEnabled"];
}
message ListCategoryDefaultFeaturesResponse {
  repeated thingmodel.service.v1.CategoryDefaultFeature items = 1;
  uint32 total = 2;
}

// BatchAddCategoryDefaultFeatures（用户在分类 Drawer 中一次勾多特征）
message BatchAddCategoryDefaultFeaturesRequest {
  uint32 category_id = 1 [json_name="categoryId"];
  message Item {
    uint32 feature_id = 1 [json_name="featureId"];
    optional thingmodel.service.v1.FeatureOverrideSpec override_spec = 2 [json_name="overrideSpec"];
    optional string display_name = 3 [json_name="displayName"];
    optional uint32 sort_order = 4 [json_name="sortOrder"];
  }
  repeated Item items = 2;
}
message BatchAddCategoryDefaultFeaturesResponse {
  repeated thingmodel.service.v1.CategoryDefaultFeature created = 1;
  repeated string skipped_duplicate_feature_codes = 2 [json_name="skippedDuplicateFeatureCodes"];
}
```

### 3.2 `admin/service/v1/product.proto`

| RPC | HTTP | 用途 |
|-----|------|------|
| `ListProducts` | `GET /v1/thingmodel/products` | 分页列产品，支持 category_id / status / manufacturer / keyword |
| `GetProduct` | `GET /v1/thingmodel/products/{id}` | 取单产品（不含特征列表，特征单独取） |
| `CreateProduct` | `POST /v1/thingmodel/products` | 创建产品（两步向导第一步落库；第二步走"拉取默认模型"另一个 RPC） |
| `UpdateProduct` | `PATCH /v1/thingmodel/products/{id}` | 更新 name/manufacturer/model_no/desc/status/is_enabled/sort，FieldMask |
| `DeleteProduct` | `DELETE /v1/thingmodel/products/{id}` | 删除（reference_count>0 拦截） |
| `PublishProduct` | `POST /v1/thingmodel/products/{id}:publish` | 状态机变更 DRAFT→PUBLISHED |
| `UnpublishProduct` | `POST /v1/thingmodel/products/{id}:unpublish` | PUBLISHED→DRAFT（本期保留接口，UI 不必暴露） |

### 3.3 `admin/service/v1/product_feature.proto`

| RPC | HTTP | 用途 |
|-----|------|------|
| `ListProductFeatures` | `GET /v1/thingmodel/products/{product_id}/features` | 列产品下特征，支持 feature_type / source 过滤 |
| `GetProductFeature` | `GET /v1/thingmodel/product-features/{id}` | 取单条（带 effective_spec 字段） |
| `CreateProductFeature` | `POST /v1/thingmodel/products/{product_id}/features` | 手动加一个（GLOBAL 或 LOCAL；DEFAULT 走 PullFromDefault） |
| `PullFromDefault` ⭐ | `POST /v1/thingmodel/products/{product_id}/features:pull-from-default` | **批量从分类默认模型拉取**（产品创建后立即调用，或在产品页面追加拉取） |
| `UpdateProductFeature` | `PATCH /v1/thingmodel/product-features/{id}` | 更新 override_spec/sort/enabled/displayName/desc，FieldMask |
| `DeleteProductFeature` | `DELETE /v1/thingmodel/product-features/{id}` | 删除单条 |
| `ReorderProductFeatures` | `POST /v1/thingmodel/products/{product_id}/features:reorder` | 拖拽排序 |
| `CloneProductFeatures` | `POST /v1/thingmodel/products/{product_id}/features:clone-from` | 从另一产品克隆全部特征（含 LOCAL）— 本期实现 |

#### PullFromDefault — 关键设计

```protobuf
message PullFromDefaultRequest {
  uint32 product_id = 1 [json_name="productId"];
  // 用户勾选的 default_feature 条目 ID 列表；空表示"全部拉取"
  repeated uint32 default_feature_ids = 2 [json_name="defaultFeatureIds"];
  // 冲突策略：产品已存在同 feature_id 的条目时
  ConflictPolicy on_conflict = 3 [json_name="onConflict"];
}

enum ConflictPolicy {
  CONFLICT_POLICY_UNSPECIFIED = 0;
  SKIP    = 1;  // 跳过冲突项（默认）
  REPLACE = 2;  // 覆盖（先删后建，丢失原 override_spec）
}

message PullFromDefaultResponse {
  repeated thingmodel.service.v1.ProductFeature created = 1;
  repeated PullSkipped skipped = 2;
  message PullSkipped {
    uint32 default_feature_id = 1 [json_name="defaultFeatureId"];
    string reason = 2;  // "duplicate" / "feature_disabled" / "category_mismatch"
  }
}
```

**实现要点**（详见 [04 §4](./04-后端实现设计.md)）：

1. 读取产品的 `category_id`，再取该 category 下的 default_features 列表（受 `default_feature_ids` 过滤）。
2. 对每条 default_feature：
   - 读 `thing_features` 拿原 spec。
   - 合并 default_feature.override_spec → 得到 spec_for_default。
   - 复制到新 ProductFeature 行：`feature_snapshot = spec_for_default`、`override_spec = NULL`（产品层暂未覆写）、`source = DEFAULT`、`ref_feature_id = feature_id`、code/identifier/name/冗余特化列从 thing_features 拷过来。
3. 单事务批量插入；`thing_features.reference_count` **不变**（DEFAULT 不增计数）。

> 重要：拉取时合并的是**默认模型层 override**，不是 thing_features 原 spec。这样产品上看到的范围就是分类约定的范围（如冷水机组的温度被分类层收紧到 4~15）。产品自己再覆写则进一步收紧/调整。

---

## 4. RPC 清单总表

| 资源 | RPC | HTTP |
|------|-----|------|
| 分类默认模型条目 | List | `GET /v1/thingmodel/categories/{category_id}/default-features` |
|  | Get | `GET /v1/thingmodel/category-default-features/{id}` |
|  | Create | `POST /v1/thingmodel/category-default-features` |
|  | BatchAdd | `POST /v1/thingmodel/categories/{category_id}/default-features:batch` |
|  | Update | `PATCH /v1/thingmodel/category-default-features/{id}` |
|  | Delete | `DELETE /v1/thingmodel/category-default-features/{id}` |
|  | Reorder | `POST /v1/thingmodel/categories/{category_id}/default-features:reorder` |
| 产品 | List | `GET /v1/thingmodel/products` |
|  | Get | `GET /v1/thingmodel/products/{id}` |
|  | Create | `POST /v1/thingmodel/products` |
|  | Update | `PATCH /v1/thingmodel/products/{id}` |
|  | Delete | `DELETE /v1/thingmodel/products/{id}` |
|  | Publish | `POST /v1/thingmodel/products/{id}:publish` |
|  | Unpublish | `POST /v1/thingmodel/products/{id}:unpublish` |
| 产品特征 | List | `GET /v1/thingmodel/products/{product_id}/features` |
|  | Get | `GET /v1/thingmodel/product-features/{id}` |
|  | Create | `POST /v1/thingmodel/products/{product_id}/features` |
|  | PullFromDefault ⭐ | `POST /v1/thingmodel/products/{product_id}/features:pull-from-default` |
|  | CloneFromProduct | `POST /v1/thingmodel/products/{product_id}/features:clone-from` |
|  | Update | `PATCH /v1/thingmodel/product-features/{id}` |
|  | Delete | `DELETE /v1/thingmodel/product-features/{id}` |
|  | Reorder | `POST /v1/thingmodel/products/{product_id}/features:reorder` |

---

## 5. 错误码（沿用 thingmodel 域错误命名空间）

```protobuf
enum ThingModelErrorReason {
  THINGMODEL_OK = 0;

  // ===== 分类默认模型 =====
  TM_CAT_DEFAULT_FEATURE_CATEGORY_NOT_LEAF = 4001;     // category 不是 level=4
  TM_CAT_DEFAULT_FEATURE_DUPLICATE         = 4002;     // (category_id, feature_id) 已存在
  TM_CAT_DEFAULT_FEATURE_OVERRIDE_INVALID  = 4003;     // override_spec 含非白名单字段
  TM_CAT_DEFAULT_FEATURE_FEATURE_DISABLED  = 4004;     // 引用的 thing_features.is_enabled=false

  // ===== 产品 =====
  TM_PRODUCT_CATEGORY_NOT_LEAF             = 4101;
  TM_PRODUCT_CODE_DUPLICATE                = 4102;     // (tenant, code) 冲突
  TM_PRODUCT_NAME_DUPLICATE                = 4103;     // (tenant, category_id, name) 冲突
  TM_PRODUCT_REFERENCED                    = 4104;     // reference_count > 0
  TM_PRODUCT_STATUS_NOT_DRAFT              = 4105;     // 试图修改 PUBLISHED 产品的结构
  TM_PRODUCT_STATUS_NOT_PUBLISHED          = 4106;     // unpublish 时

  // ===== 产品特征 =====
  TM_PF_SOURCE_REF_MISMATCH                = 4201;     // source=LOCAL 但 ref_feature_id 非空，或反之
  TM_PF_DUPLICATE_CODE                     = 4202;     // (product_id, code) 冲突
  TM_PF_DUPLICATE_IDENTIFIER               = 4203;
  TM_PF_SPEC_TYPE_MISMATCH                 = 4204;     // feature_snapshot 与 feature_type 不一致
  TM_PF_OVERRIDE_INVALID                   = 4205;
  TM_PF_PRODUCT_PUBLISHED                  = 4206;     // 产品 PUBLISHED 不允许结构变更
}
```

> 错误用 Kratos `errors.New(reason)` 风格，前端在 hook 层统一 toast；具体 i18n 见 [05 §5](./05-前端实现设计.md)。

---

## 6. ListProducts 查询参数

```protobuf
message ListProductsRequest {
  // 分页（PaginationQuery，项目通用形态）
  uint32 page = 1;
  uint32 page_size = 2 [json_name="pageSize"];
  string order_by  = 3 [json_name="orderBy"];   // "created_at desc" / "sort_order asc" 等

  // 过滤
  optional uint32 category_id  = 10 [json_name="categoryId"];
  optional thingmodel.service.v1.ProductStatus status = 11;
  optional string manufacturer = 12;
  optional string keyword      = 13;  // name / code / model_no 模糊
  optional bool   is_enabled   = 14 [json_name="isEnabled"];

  // 树过滤：传 category_id 时是否包含子树（本期 level=4 已是叶，理论无子；但前端可能传中类，按 code 前缀展开）
  optional bool include_descendants = 20 [json_name="includeDescendants"];
}
```

---

## 7. 与已有 proto 的复用关系

| 复用 | 来源 |
|------|------|
| `FeatureSpec` (oneof: PropertySpec/EventSpec/ServiceSpec/RelationSpec) | `thingmodel/service/v1/feature.proto` |
| `FeatureType` 枚举 | 同上 |
| `DataType / AccessMode / EventLevel / CallMode` 枚举 | 同上 |
| `ValueConstraints` / `UnitRef` / `ParamSpec` | 同上 |
| Pagination / FieldMask 通用形态 | `common/service/v1/` |
| 错误码风格 | `thingmodel/service/v1/errors.proto`（已有，本章追加 reason） |

**绝不重定义**已存在的 message/enum；BFF 层 import 域层 message。
