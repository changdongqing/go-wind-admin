# 04 · API 与 Proto 设计

> 严格遵循项目 Proto 两层架构（源领域层 + BFF 层），风格对齐 `dict_type.proto` / `i_dict_type.proto` / `dict_error.proto`。
> 所有 proto 示例可直接作为开发起点，仅需补充 buf 生成。

---

## 1. 文件清单

| 层 | 路径 | 文件 |
|----|------|------|
| 源领域层 | `backend/api/protos/thingmodel/service/v1/` | `unit_category.proto`、`unit.proto`、`thingmodel_error.proto` |
| BFF 层 | `backend/api/protos/admin/service/v1/` | `i_unit_category.proto`、`i_unit.proto` |

---

## 2. 源领域层

### 2.1 `thingmodel_error.proto`（错误定义）

> 全量 HTTP 状态码模板**照搬** `dict_error.proto`（BAD_REQUEST/UNAUTHORIZED/.../NETWORK_CONNECT_TIMEOUT_ERROR），以下仅列出需要在模板基础上**追加的单位业务错误码**。

```protobuf
syntax = "proto3";
package thingmodel.service.v1;

import "errors/errors.proto";

enum ThingModelErrorReason {
    option (errors.default_code) = 500;

    // ===== 照搬 dict_error.proto 的全量 HTTP 状态码枚举 =====
    BAD_REQUEST = 0 [(errors.code) = 400];
    UNAUTHORIZED = 100 [(errors.code) = 401];
    FORBIDDEN = 300 [(errors.code) = 403];
    NOT_FOUND = 400 [(errors.code) = 404];
    CONFLICT = 900 [(errors.code) = 409];
    INTERNAL_SERVER_ERROR = 2000 [(errors.code) = 500];
    // ... 其余照搬 ...

    // ===== 单位管理业务错误码（追加段）=====
    UNIT_NOT_FOUND             = 40010 [(errors.code) = 404];  // 单位不存在
    UNIT_CATEGORY_NOT_FOUND    = 40011 [(errors.code) = 404];  // 物理量分类不存在
    UNIT_DIFFERENT_CATEGORY    = 40020 [(errors.code) = 400];  // 源/目标单位不属于同一物理量分类
    UNIT_NOT_CONVERTIBLE       = 40021 [(errors.code) = 400];  // 单位不可换算（NONE 类型）
    UNIT_LOGARITHMIC_NOT_CONVERTIBLE = 40022 [(errors.code) = 400]; // 对数单位不可线性换算
    UNIT_CONDITIONAL_REQUIRES_PARAMS = 40023 [(errors.code) = 400]; // 条件换算缺失外部参数
    UNIT_OVERFLOW              = 40024 [(errors.code) = 400];  // 换算结果溢出/非数值
    UNIT_INVALID_FACTOR        = 40030 [(errors.code) = 400];  // 系数异常（factor=0 等）
    UNIT_BASE_FACTOR_INVALID   = 40031 [(errors.code) = 400];  // 基准单位系数必须为 1/0
    UNIT_LINEAR_OFFSET_MUST_BE_ZERO = 40032 [(errors.code) = 400]; // 线性单位偏移必须为 0
    UNIT_FACTOR_ZERO           = 40033 [(errors.code) = 400];  // 系数不可为 0
    UNIT_BASE_ALREADY_EXISTS   = 40034 [(errors.code) = 409];  // 该分类已存在基准单位
    UNIT_IN_USE_CANNOT_DELETE  = 40040 [(errors.code) = 409];  // 单位被引用，不可删除
    UNIT_CODE_DUPLICATED       = 40041 [(errors.code) = 409];  // 单位编码重复
    UNIT_CATEGORY_CODE_DUPLICATED = 40042 [(errors.code) = 409]; // 分类编码重复
}
```

生成后即可使用：`thingmodelV1.ErrorUnitNotFound("...")`、`thingmodelV1.ErrorUnitDifferentCategory("...")` 等。

### 2.2 `unit_category.proto`（物理量分类）

```protobuf
syntax = "proto3";
package thingmodel.service.v1;

import "gnostic/openapi/v3/annotations.proto";
import "google/protobuf/empty.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/field_mask.proto";
import "pagination/v1/pagination.proto";

// 物理量分类服务（源领域层，无 HTTP 注解）
service UnitCategoryService {
  rpc List   (pagination.PagingRequest)        returns (ListUnitCategoryResponse)     {}
  rpc Count  (pagination.PagingRequest)        returns (CountUnitCategoryResponse)    {}
  rpc Get    (GetUnitCategoryRequest)          returns (UnitCategory)                 {}
  rpc Create (CreateUnitCategoryRequest)       returns (google.protobuf.Empty)        {}
  rpc Update (UpdateUnitCategoryRequest)       returns (google.protobuf.Empty)        {}
  rpc Delete (DeleteUnitCategoryRequest)       returns (google.protobuf.Empty)        {}
}

message UnitCategory {
  optional uint32 id = 1 [json_name = "id",
    (gnostic.openapi.v3.property) = {description: "物理量分类ID"}];

  optional string code = 2 [json_name = "code",
    (gnostic.openapi.v3.property) = {description: "物理量编码，如 temperature/pressure（不可变）"}];
  optional string name = 3 [json_name = "name",
    (gnostic.openapi.v3.property) = {description: "中文名，如 温度"}];
  optional string name_en = 4 [json_name = "nameEn",
    (gnostic.openapi.v3.property) = {description: "英文名"}];
  optional string quantity = 5 [json_name = "quantity",
    (gnostic.openapi.v3.property) = {description: "量纲/物理量名，如 热力学温度"}];
  optional string base_unit_symbol = 6 [json_name = "baseUnitSymbol",
    (gnostic.openapi.v3.property) = {description: "基准单位符号（冗余展示，如 K）"}];
  optional string icon = 7 [json_name = "icon",
    (gnostic.openapi.v3.property) = {description: "Iconify 图标名"}];
  optional string description = 8 [json_name = "description",
    (gnostic.openapi.v3.property) = {description: "描述"}];

  optional bool is_enabled = 9 [json_name = "isEnabled",
    (gnostic.openapi.v3.property) = {description: "是否启用"}];
  optional uint32 sort_order = 10 [json_name = "sortOrder",
    (gnostic.openapi.v3.property) = {description: "排序号"}];

  optional uint32 tenant_id = 100 [json_name = "tenantId"];   // 租户ID，0=系统全局
  optional string tenant_name = 101 [json_name = "tenantName"];

  optional uint32 created_by = 200 [json_name = "createdBy"];
  optional uint32 updated_by = 201 [json_name = "updatedBy"];
  optional uint32 deleted_by = 202 [json_name = "deletedBy"];
  optional google.protobuf.Timestamp created_at = 300 [json_name = "createdAt"];
  optional google.protobuf.Timestamp updated_at = 301 [json_name = "updatedAt"];
  optional google.protobuf.Timestamp deleted_at = 302 [json_name = "deletedAt"];
}

message ListUnitCategoryResponse  { repeated UnitCategory items = 1; uint64 total = 2; }
message CountUnitCategoryResponse { uint64 count = 1; }

message GetUnitCategoryRequest {
  oneof query_by { uint32 id = 1; string code = 2; }
  optional google.protobuf.FieldMask view_mask = 100 [json_name = "viewMask"];
}
message CreateUnitCategoryRequest { UnitCategory data = 1; }
message UpdateUnitCategoryRequest {
  uint32 id = 1;
  UnitCategory data = 2;
  google.protobuf.FieldMask update_mask = 3 [json_name = "updateMask"];
  optional bool allow_missing = 4 [json_name = "allowMissing"];
}
message DeleteUnitCategoryRequest { repeated uint32 ids = 1; }
```

### 2.3 `unit.proto`（单位 + 换算）

```protobuf
syntax = "proto3";
package thingmodel.service.v1;

import "gnostic/openapi/v3/annotations.proto";
import "google/protobuf/empty.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/field_mask.proto";
import "pagination/v1/pagination.proto";

// 单位服务（源领域层）
service UnitService {
  rpc List   (pagination.PagingRequest)   returns (ListUnitResponse)  {}
  rpc Count  (pagination.PagingRequest)   returns (CountUnitResponse) {}
  rpc Get    (GetUnitRequest)             returns (Unit)              {}
  rpc Create (CreateUnitRequest)          returns (google.protobuf.Empty) {}
  rpc Update (UpdateUnitRequest)          returns (google.protobuf.Empty) {}
  rpc Delete (DeleteUnitRequest)          returns (google.protobuf.Empty) {}

  // 按物理量分类查询单位（属性选单位时高频调用）
  rpc ListByCategory (ListUnitByCategoryRequest) returns (ListUnitResponse) {}

  // 单位换算（核心业务能力，见 03 章）
  rpc Convert (ConvertUnitRequest) returns (ConvertUnitResponse) {}
}

// 换算类型枚举
enum ConversionType {
  CONVERSION_TYPE_UNSPECIFIED = 0;
  LINEAR      = 1;  // 线性 y = x·k
  AFFINE      = 2;  // 仿射 y = x·k + b（温度等）
  LOGARITHMIC = 3;  // 对数（dB/dBm，不可线性换算）
  CONDITIONAL = 4;  // 条件换算（依赖外部参数）
  NONE        = 5;  // 不可换算（无量纲/计数）
}

message Unit {
  optional uint32 id = 1 [json_name = "id",
    (gnostic.openapi.v3.property) = {description: "单位ID"}];

  optional uint32 category_id = 2 [json_name = "categoryId",
    (gnostic.openapi.v3.property) = {description: "所属物理量分类ID"}];
  optional string code = 3 [json_name = "code",
    (gnostic.openapi.v3.property) = {description: "单位编码，如 celsius（不可变）"}];
  optional string symbol = 4 [json_name = "symbol",
    (gnostic.openapi.v3.property) = {description: "单位符号，如 ℃"}];
  optional string name = 5 [json_name = "name",
    (gnostic.openapi.v3.property) = {description: "中文名"}];
  optional string name_en = 6 [json_name = "nameEn",
    (gnostic.openapi.v3.property) = {description: "英文名"}];

  optional bool is_base = 10 [json_name = "isBase",
    (gnostic.openapi.v3.property) = {description: "是否基准单位（每分类唯一）"}];

  optional ConversionType conversion_type = 11 [json_name = "conversionType",
    (gnostic.openapi.v3.property) = {description: "换算类型"}];
  optional double factor = 12 [json_name = "factor",
    (gnostic.openapi.v3.property) = {description: "线性系数 k（基准=原值·k+offset）"}];
  optional double offset = 13 [json_name = "offset",
    (gnostic.openapi.v3.property) = {description: "偏移量 b（仅仿射非0）"}];
  optional string formula_expr = 14 [json_name = "formulaExpr",
    (gnostic.openapi.v3.property) = {description: "公式说明（仅展示）"}];

  optional int32 precision = 15 [json_name = "precision",
    (gnostic.openapi.v3.property) = {description: "建议显示精度（小数位）"}];
  optional bool is_si_unit = 16 [json_name = "isSiUnit",
    (gnostic.openapi.v3.property) = {description: "是否 SI 单位"}];
  optional bool is_legal_unit = 17 [json_name = "isLegalUnit",
    (gnostic.openapi.v3.property) = {description: "是否中国法定计量单位"}];
  optional uint32 reference_count = 18 [json_name = "referenceCount",
    (gnostic.openapi.v3.property) = {description: "被物模型属性引用次数（预留）"}];

  optional bool is_enabled = 19 [json_name = "isEnabled"];
  optional uint32 sort_order = 20 [json_name = "sortOrder"];

  optional uint32 tenant_id = 100 [json_name = "tenantId"];
  optional string tenant_name = 101 [json_name = "tenantName"];
  optional uint32 created_by = 200 [json_name = "createdBy"];
  optional uint32 updated_by = 201 [json_name = "updatedBy"];
  optional uint32 deleted_by = 202 [json_name = "deletedBy"];
  optional google.protobuf.Timestamp created_at = 300 [json_name = "createdAt"];
  optional google.protobuf.Timestamp updated_at = 301 [json_name = "updatedAt"];
  optional google.protobuf.Timestamp deleted_at = 302 [json_name = "deletedAt"];
}

message ListUnitResponse  { repeated Unit items = 1; uint64 total = 2; }
message CountUnitResponse { uint64 count = 1; }

message GetUnitRequest {
  oneof query_by { uint32 id = 1; string code = 2; }
  optional google.protobuf.FieldMask view_mask = 100 [json_name = "viewMask"];
}
message CreateUnitRequest { Unit data = 1; }
message UpdateUnitRequest {
  uint32 id = 1;
  Unit data = 2;
  google.protobuf.FieldMask update_mask = 3 [json_name = "updateMask"];
  optional bool allow_missing = 4 [json_name = "allowMissing"];
}
message DeleteUnitRequest { repeated uint32 ids = 1; }

// 按分类查询
message ListUnitByCategoryRequest {
  oneof category_by { uint32 category_id = 1; string category_code = 2; }
  optional bool only_enabled = 3 [json_name = "onlyEnabled"];
}

// ===== 换算 =====
message ConvertUnitRequest {
  oneof source_by { uint32 source_unit_id = 1 [json_name = "sourceUnitId"]; string source_unit_code = 2 [json_name = "sourceUnitCode"]; }
  oneof target_by { uint32 target_unit_id = 3 [json_name = "targetUnitId"]; string target_unit_code = 4 [json_name = "targetUnitCode"]; }
  double value = 5;
  optional int32 precision = 6 [json_name = "precision"];
}

enum ConvertUnitStatus {
  CONVERT_STATUS_UNSPECIFIED = 0;
  CONVERT_OK                 = 1;  // 换算成功
  CONVERT_NOT_FOUND          = 2;  // 单位不存在
  CONVERT_DIFFERENT_CATEGORY = 3;  // 不同分类
  CONVERT_NOT_CONVERTIBLE    = 4;  // 不可换算
}

message ConvertUnitResponse {
  double result = 1 [json_name = "result"];
  string formula = 2 [json_name = "formula"];          // 人类可读换算链
  ConvertUnitStatus status = 3 [json_name = "status"];
  string message = 4 [json_name = "message"];
  optional double base_value = 5 [json_name = "baseValue"];
}
```

---

## 3. BFF 层（REST）

### 3.1 `i_unit_category.proto`

```protobuf
syntax = "proto3";
package admin.service.v1;

import "google/api/annotations.proto";
import "google/protobuf/empty.proto";
import "pagination/v1/pagination.proto";
import "thingmodel/service/v1/unit_category.proto";

service UnitCategoryService {
  rpc List   (pagination.PagingRequest)                        returns (thingmodel.service.v1.ListUnitCategoryResponse) {
    option (google.api.http) = { get: "/admin/v1/thingmodel/unit-categories" };
  }
  rpc Get    (thingmodel.service.v1.GetUnitCategoryRequest)    returns (thingmodel.service.v1.UnitCategory) {
    option (google.api.http) = {
      get: "/admin/v1/thingmodel/unit-categories/{id}"
      additional_bindings { get: "/admin/v1/thingmodel/unit-categories/code/{code}" }
    };
  }
  rpc Create (thingmodel.service.v1.CreateUnitCategoryRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = { post: "/admin/v1/thingmodel/unit-categories" body: "*" };
  }
  rpc Update (thingmodel.service.v1.UpdateUnitCategoryRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = { put: "/admin/v1/thingmodel/unit-categories/{id}" body: "*" };
  }
  rpc Delete (thingmodel.service.v1.DeleteUnitCategoryRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = { delete: "/admin/v1/thingmodel/unit-categories" };
  }
}
```

### 3.2 `i_unit.proto`

```protobuf
syntax = "proto3";
package admin.service.v1;

import "google/api/annotations.proto";
import "google/protobuf/empty.proto";
import "pagination/v1/pagination.proto";
import "thingmodel/service/v1/unit.proto";

service UnitService {
  rpc List            (pagination.PagingRequest)                  returns (thingmodel.service.v1.ListUnitResponse) {
    option (google.api.http) = { get: "/admin/v1/thingmodel/units" };
  }
  rpc Get             (thingmodel.service.v1.GetUnitRequest)      returns (thingmodel.service.v1.Unit) {
    option (google.api.http) = {
      get: "/admin/v1/thingmodel/units/{id}"
      additional_bindings { get: "/admin/v1/thingmodel/units/code/{code}" }
    };
  }
  rpc Create          (thingmodel.service.v1.CreateUnitRequest)   returns (google.protobuf.Empty) {
    option (google.api.http) = { post: "/admin/v1/thingmodel/units" body: "*" };
  }
  rpc Update          (thingmodel.service.v1.UpdateUnitRequest)   returns (google.protobuf.Empty) {
    option (google.api.http) = { put: "/admin/v1/thingmodel/units/{id}" body: "*" };
  }
  rpc Delete          (thingmodel.service.v1.DeleteUnitRequest)   returns (google.protobuf.Empty) {
    option (google.api.http) = { delete: "/admin/v1/thingmodel/units" };
  }
  rpc ListByCategory  (thingmodel.service.v1.ListUnitByCategoryRequest) returns (thingmodel.service.v1.ListUnitResponse) {
    option (google.api.http) = { get: "/admin/v1/thingmodel/unit-categories/{category_id}/units" };
  }
  rpc Convert         (thingmodel.service.v1.ConvertUnitRequest)  returns (thingmodel.service.v1.ConvertUnitResponse) {
    option (google.api.http) = { post: "/admin/v1/thingmodel/units/convert" body: "*" };
  }
}
```

---

## 4. REST 路由总表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/v1/thingmodel/unit-categories` | 分类分页列表 |
| GET | `/admin/v1/thingmodel/unit-categories/{id}` | 分类详情（按 ID） |
| GET | `/admin/v1/thingmodel/unit-categories/code/{code}` | 分类详情（按 code） |
| POST | `/admin/v1/thingmodel/unit-categories` | 新建分类 |
| PUT | `/admin/v1/thingmodel/unit-categories/{id}` | 更新分类 |
| DELETE | `/admin/v1/thingmodel/unit-categories?ids=1,2,3` | 批量删除分类 |
| GET | `/admin/v1/thingmodel/units` | 单位分页列表（支持 `categoryId` 过滤） |
| GET | `/admin/v1/thingmodel/units/{id}` | 单位详情（按 ID） |
| GET | `/admin/v1/thingmodel/units/code/{code}` | 单位详情（按 code） |
| POST | `/admin/v1/thingmodel/units` | 新建单位 |
| PUT | `/admin/v1/thingmodel/units/{id}` | 更新单位 |
| DELETE | `/admin/v1/thingmodel/units?ids=1,2,3` | 批量删除单位 |
| GET | `/admin/v1/thingmodel/unit-categories/{category_id}/units` | 某分类下全部单位（下拉选择器用） |
| POST | `/admin/v1/thingmodel/units/convert` | 单位换算 |

---

## 5. 命名与字段编号约定（与 dict 一致）

| 段位 | 用途 | 说明 |
|------|------|------|
| `1~9` | 核心业务字段 | id、外键、code、symbol、name 等 |
| `10~19` | 扩展业务字段 | 本设计用于换算字段（is_base/conversion_type/factor/offset...） |
| `100~102` | tenant_id/tenant_name / created_by / updated_by / deleted_by | 审计与租户 |
| `200~202` | created_at / updated_at / deleted_at | 时间戳 |

> **注意**：本设计的业务字段稍多（单位表含换算字段），故审计/时间字段段位顺延至 `100+/200+`，与 dict 的 `10+/100+` 错开但同构。生成代码后 json_name 不受影响，前端无感知。

---

## 6. 生成命令

```bash
# 后端 Go 代码 + OpenAPI
cd backend && make api && make openapi

# 三端 TS 客户端（BFF 层 protos/admin/service/v1 自动被三个 buf 模板消费）
make ts
```

生成产物（禁止手改）：
- `backend/api/gen/go/thingmodel/service/v1/*.go`
- `backend/api/gen/go/admin/service/v1/i_unit*_http.pb.go`
- `frontend/admin/react/src/api/generated/admin/service/v1/index.ts`（新增 `unitservicev1_*`、`unitcategoryservicev1_*` 类型与 client）
