# 03 · API 与 Proto 设计（v2.0）

> 严格遵循项目 Proto 两层架构（源领域层 + BFF 层），风格对齐 `unit_category.proto` / `i_unit_category.proto`。

---

## 1. 文件清单

| 层 | 路径 | 文件 |
|----|------|------|
| 源领域层 | `backend/api/protos/thingmodel/service/v1/` | `category.proto`（错误码追加进既有 `thingmodel_error.proto`） |
| BFF 层 | `backend/api/protos/admin/service/v1/` | `i_category.proto` |

---

## 2. 错误码追加（`thingmodel_error.proto`）

在 `thingmodel_error.proto` 既有枚举中追加（与单位错误码段位错开，使用 `41xxx` 段）：

```protobuf
// ===== 分类管理业务错误码（追加段）=====
CATEGORY_NOT_FOUND              = 41010 [(errors.code) = 404];  // 分类不存在
CATEGORY_PARENT_NOT_FOUND       = 41011 [(errors.code) = 404];  // 父分类不存在

CATEGORY_CODE_FORMAT_INVALID    = 41020 [(errors.code) = 400];  // 编码格式不合法（非纯数字）
CATEGORY_CODE_LENGTH_MISMATCH   = 41021 [(errors.code) = 400];  // 编码长度与 level 不匹配（必须 = level*2）
CATEGORY_CODE_PREFIX_MISMATCH   = 41022 [(errors.code) = 400];  // 编码必须以父编码为严格前缀且长度 + 2

CATEGORY_LEVEL_INVALID          = 41023 [(errors.code) = 400];  // 层级不在 1..4 范围
CATEGORY_LEVEL_PARENT_MISMATCH  = 41024 [(errors.code) = 400];  // 父分类层级与本层级不匹配
CATEGORY_KIND_PARENT_MISMATCH   = 41025 [(errors.code) = 400];  // 父子分类 kind 不一致
CATEGORY_PARENT_REQUIRED        = 41026 [(errors.code) = 400];  // level>1 必须指定父分类
CATEGORY_PARENT_FORBIDDEN       = 41027 [(errors.code) = 400];  // level=1 不能指定父分类

CATEGORY_HAS_CHILDREN           = 41030 [(errors.code) = 409];  // 存在子分类，不可删除
CATEGORY_IN_USE_CANNOT_DELETE   = 41031 [(errors.code) = 409];  // 被引用，不可删除
CATEGORY_CODE_DUPLICATED        = 41040 [(errors.code) = 409];  // 同 (kind) 下编码重复

CATEGORY_IMMUTABLE_FIELD        = 41050 [(errors.code) = 400];  // 试图修改不可变字段（kind/code/parent_id/level）
```

> 与 v1.0 的差异：去掉 `CATEGORY_PATH_*` 相关错误（不再有 path 字段）；`CATEGORY_CODE_LENGTH_MISMATCH` 是新增的 v2 专属错误。

---

## 3. 源领域层：`category.proto`

```protobuf
syntax = "proto3";
package thingmodel.service.v1;

import "gnostic/openapi/v3/annotations.proto";
import "google/protobuf/empty.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/field_mask.proto";
import "pagination/v1/pagination.proto";

// 分类种类（未来扩展只需在此追加枚举值）
enum CategoryKind {
  CATEGORY_KIND_UNSPECIFIED = 0;

  // ===== 本期：物模型三 kind =====
  SYSTEM   = 1;  // 智能系统（30-36 段大类）
  SPACE    = 2;  // 空间（10 段大类）
  FACILITY = 3;  // 设备设施（20-26 段大类）

  // ===== 未来扩展示例（暂注释，业务方接入时启用）=====
  // DOCUMENT    = 10;
  // CERTIFICATE = 11;
  // FORM        = 12;
}

// 分类服务（源领域层，无 HTTP 注解）
service CategoryService {
  rpc List     (pagination.PagingRequest)   returns (ListCategoryResponse)   {}
  rpc Count    (pagination.PagingRequest)   returns (CountCategoryResponse)  {}
  rpc Get      (GetCategoryRequest)         returns (Category)               {}
  rpc Create   (CreateCategoryRequest)      returns (google.protobuf.Empty)  {}
  rpc Update   (UpdateCategoryRequest)      returns (google.protobuf.Empty)  {}
  rpc Delete   (DeleteCategoryRequest)      returns (google.protobuf.Empty)  {}
}

// ⚠️ 移除 v1.0 的 ListChildren / GetTree RPC——v2.0 用 List 即可：
//    List(kind=SYSTEM, level=4, code_prefix='3001')  → 取某子树的细类
//    List(kind=SYSTEM, level=1)                     → 取所有大类
//    List(kind=SYSTEM, code_prefix='3001')          → 取某子树全部 4 层
// 前端按 code 升序排展示即可，无需后端聚合成树。

message Category {
  optional uint32 id = 1 [json_name = "id",
    (gnostic.openapi.v3.property) = {description: "分类ID"}];

  optional CategoryKind kind = 2 [json_name = "kind",
    (gnostic.openapi.v3.property) = {description: "分类种类（不可变）"}];
  optional string code = 3 [json_name = "code",
    (gnostic.openapi.v3.property) = {description: "分类编码（变长 2/4/6/8 位纯数字，不可变）"}];
  optional uint32 level = 4 [json_name = "level",
    (gnostic.openapi.v3.property) = {description: "层级：1=大类 2=中类 3=小类 4=细类"}];
  optional uint32 parent_id = 5 [json_name = "parentId",
    (gnostic.openapi.v3.property) = {description: "父分类ID（level=1 时为空）"}];

  optional string name = 6 [json_name = "name",
    (gnostic.openapi.v3.property) = {description: "中文名"}];
  optional string name_en = 7 [json_name = "nameEn",
    (gnostic.openapi.v3.property) = {description: "英文名"}];
  optional string icon = 8 [json_name = "icon",
    (gnostic.openapi.v3.property) = {description: "Iconify 图标名"}];
  optional string description = 9 [json_name = "description",
    (gnostic.openapi.v3.property) = {description: "描述"}];

  optional uint32 reference_count = 10 [json_name = "referenceCount",
    (gnostic.openapi.v3.property) = {description: "被物模型/实例引用次数（预留）"}];

  optional bool   is_enabled = 11 [json_name = "isEnabled"];
  optional uint32 sort_order = 12 [json_name = "sortOrder"];

  optional uint32 tenant_id   = 100 [json_name = "tenantId"];
  optional string tenant_name = 101 [json_name = "tenantName"];

  optional uint32 created_by = 200 [json_name = "createdBy"];
  optional uint32 updated_by = 201 [json_name = "updatedBy"];
  optional uint32 deleted_by = 202 [json_name = "deletedBy"];
  optional google.protobuf.Timestamp created_at = 300 [json_name = "createdAt"];
  optional google.protobuf.Timestamp updated_at = 301 [json_name = "updatedAt"];
  optional google.protobuf.Timestamp deleted_at = 302 [json_name = "deletedAt"];
}

message ListCategoryResponse  { repeated Category items = 1; uint64 total = 2; }
message CountCategoryResponse { uint64 count = 1; }

message GetCategoryRequest {
  oneof query_by {
    uint32 id = 1;
    string code = 2;     // 与 kind 配合使用
  }
  optional CategoryKind kind = 3 [json_name = "kind"];  // 当用 code 查时必填
  optional google.protobuf.FieldMask view_mask = 100 [json_name = "viewMask"];
}

message CreateCategoryRequest { Category data = 1; }

message UpdateCategoryRequest {
  uint32 id = 1;
  Category data = 2;
  google.protobuf.FieldMask update_mask = 3 [json_name = "updateMask"];
  optional bool allow_missing = 4 [json_name = "allowMissing"];
}

message DeleteCategoryRequest {
  repeated uint32 ids = 1;
}
```

> 与 v1.0 差异：
> - **去掉 `ListChildren` / `GetTree` RPC**（前端用 List + 过滤即可）
> - **去掉 `is_leaf` / `path` 字段**
> - `kind` 枚举注释中给出未来扩展的占位

---

## 4. BFF 层：`i_category.proto`

```protobuf
syntax = "proto3";
package admin.service.v1;

import "google/api/annotations.proto";
import "google/protobuf/empty.proto";
import "pagination/v1/pagination.proto";
import "thingmodel/service/v1/category.proto";

service CategoryService {
  rpc List   (pagination.PagingRequest)                    returns (thingmodel.service.v1.ListCategoryResponse) {
    option (google.api.http) = { get: "/admin/v1/thingmodel/categories" };
  }
  rpc Get    (thingmodel.service.v1.GetCategoryRequest)    returns (thingmodel.service.v1.Category) {
    option (google.api.http) = {
      get: "/admin/v1/thingmodel/categories/{id}"
      additional_bindings { get: "/admin/v1/thingmodel/categories/code/{code}" }
    };
  }
  rpc Create (thingmodel.service.v1.CreateCategoryRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = { post: "/admin/v1/thingmodel/categories" body: "*" };
  }
  rpc Update (thingmodel.service.v1.UpdateCategoryRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = { put: "/admin/v1/thingmodel/categories/{id}" body: "*" };
  }
  rpc Delete (thingmodel.service.v1.DeleteCategoryRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = { delete: "/admin/v1/thingmodel/categories" };
  }
}
```

→ **共 5 个 REST 路由**（v1.0 是 9 个，砍掉 4 个）。

---

## 5. REST 路由总表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/v1/thingmodel/categories` | 分类分页列表（支持 `kind` / `level` / `parentId` / `codePrefix` / `keyword` 过滤） |
| GET | `/admin/v1/thingmodel/categories/{id}` | 分类详情（按 ID） |
| GET | `/admin/v1/thingmodel/categories/code/{code}?kind=SYSTEM` | 分类详情（按 code，需带 kind） |
| POST | `/admin/v1/thingmodel/categories` | 新建分类 |
| PUT | `/admin/v1/thingmodel/categories/{id}` | 更新分类（不可改 kind/code/parent_id/level） |
| DELETE | `/admin/v1/thingmodel/categories?ids=1,2,3` | 批量删除（有子节点的拒绝） |

---

## 6. 列表查询 - 通用过滤器约定

`List` RPC 走标准 `pagination.PagingRequest`，前端在 `query` 中传过滤器（与项目其他列表一致）：

| 字段 | 类型 | 说明 |
|------|------|------|
| `kind` | string | `SYSTEM` / `SPACE` / `FACILITY`，**UI 默认根据当前 Tab 强制带上** |
| `level` | uint32 | 1..4，限定当前层级 |
| `parent_id` | uint32 | 仅查指定父节点直接子节点 |
| `code_prefix` | string | 子树查询（如 `3001` 取「冷热源系统」整子树，含 4 层节点） |
| `keyword` | string | 模糊匹配 `name` / `code` |
| `is_enabled` | bool | 启停过滤 |

### 6.1 排序约定

- **默认按 `code` 字符串升序**（天然 = 深度优先遍历层级顺序）。
- 用户可改按 `sort_order` 升序 + `code` 升序作为次序。
- 前端默认不暴露排序切换——code 升序就是最自然的展示。

---

## 7. 与 v1.0 的 API 兼容性

v2.0 是**首次落地**版本，v1.0 仅为设计稿，因此**无需考虑 API 兼容性**。本设计直接以 v2.0 为最终实现版。

---

## 8. 命名与字段编号约定（与 unit 一致）

| 段位 | 用途 |
|------|------|
| `1~12` | 核心业务字段（id / kind / code / level / parent_id / name / ... / sort_order） |
| `100~101` | tenant_id / tenant_name |
| `200~202` | created_by / updated_by / deleted_by |
| `300~302` | created_at / updated_at / deleted_at |

---

## 9. 生成命令

```bash
# 后端 Go 代码 + OpenAPI
cd backend && make api && make openapi

# 三端 TS 客户端
make ts
```

生成产物（禁止手改）：
- `backend/api/gen/go/thingmodel/service/v1/category.pb.go` 等
- `backend/api/gen/go/admin/service/v1/i_category_http.pb.go`
- `frontend/admin/react/src/api/generated/admin/service/v1/index.ts` 新增 `categoryservicev1_*` 类型与 client
