# 11 · 特征 API 与 Proto 设计

> 严格遵循项目 Proto 两层架构（源领域层 + BFF 层），风格对齐单位管理 `unit.proto` / `i_unit.proto`。
> 错误码追加进已有 `thingmodel_error.proto`（与单位错误码同文件），不新建 error proto。

---

## 1. 文件清单

| 层 | 路径 | 文件 |
|----|------|------|
| 源领域层 | `backend/api/protos/thingmodel/service/v1/` | `feature.proto`（追加错误码到 `thingmodel_error.proto`） |
| BFF 层 | `backend/api/protos/admin/service/v1/` | `i_feature.proto` |

---

## 2. 源领域层

### 2.1 错误码追加（`thingmodel_error.proto`）

在已有 `ThingModelErrorReason` enum 末尾追加特征业务错误码（与单位错误码 40010-40042 段衔接，特征用 50000 段）：

```protobuf
enum ThingModelErrorReason {
    option (errors.default_code) = 500;

    // ... 已有：单位错误码 40010~40042（照搬，省略）...

    // ===== 特征管理业务错误码（追加段）=====
    FEATURE_NOT_FOUND              = 50010 [(errors.code) = 404]; // 特征不存在
    FEATURE_TYPE_INVALID           = 50011 [(errors.code) = 400]; // 特征类型非法
    FEATURE_SPEC_INVALID           = 50020 [(errors.code) = 400]; // spec 结构无效
    FEATURE_SPEC_MISMATCH          = 50021 [(errors.code) = 400]; // 特化列与 spec 不一致
    FEATURE_ENUM_EMPTY             = 50022 [(errors.code) = 400]; // 枚举类型缺 enumItems
    FEATURE_STRUCT_EMPTY           = 50023 [(errors.code) = 400]; // 结构体类型缺 structFields
    FEATURE_CONSTRAINT_INVALID     = 50024 [(errors.code) = 400]; // 约束非法（min>max 等）
    FEATURE_RELATION_TARGET_NOT_FOUND = 50025 [(errors.code) = 404]; // 关系目标特征不存在
    FEATURE_CODE_DUPLICATED        = 50030 [(errors.code) = 409]; // 特征编码重复
    FEATURE_IDENTIFIER_DUPLICATED  = 50031 [(errors.code) = 409]; // 标识符重复
    FEATURE_IN_USE_CANNOT_DELETE   = 50040 [(errors.code) = 409]; // 特征被关系引用，不可删除
    FEATURE_UNIT_REFERENCE_FAIL    = 50050 [(errors.code) = 500]; // 维护单位引用计数失败
}
```

生成后即可使用：`thingmodelV1.ErrorFeatureNotFound("...")`、`thingmodelV1.ErrorFeatureSpecInvalid("...")` 等。

### 2.2 `feature.proto`（特征 + 四类 spec）

```protobuf
syntax = "proto3";
package thingmodel.service.v1;

import "gnostic/openapi/v3/annotations.proto";
import "google/protobuf/empty.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/field_mask.proto";
import "pagination/v1/pagination.proto";

// 特征类型 / Feature type
enum FeatureType {
  FEATURE_TYPE_UNSPECIFIED = 0;
  PROPERTY = 1;   // 属性 / Property
  EVENT = 2;       // 事件 / Event
  SERVICE = 3;     // 服务 / Service
  RELATION = 4;    // 关系（本体）/ Relation (ontology)
}

// 属性访问模式 / Property access mode
enum AccessMode {
  ACCESS_MODE_UNSPECIFIED = 0;
  R = 1;   // 只读 / Read-only
  RW = 2;  // 读写 / Read-write
}

// 事件级别 / Event level
enum EventLevel {
  EVENT_LEVEL_UNSPECIFIED = 0;
  INFO = 1;    // 信息 / Info
  ALERT = 2;   // 告警 / Alert
  ERROR = 3;   // 故障 / Error
}

// 服务调用模式 / Service call mode
enum CallMode {
  CALL_MODE_UNSPECIFIED = 0;
  ASYNC = 1;  // 异步 / Asynchronous
  SYNC = 2;   // 同步 / Synchronous
}

// 属性数据类型 / Property data type
enum DataType {
  DATA_TYPE_UNSPECIFIED = 0;
  INT = 1;
  FLOAT = 2;
  DOUBLE = 3;
  BOOL = 4;
  ENUM = 5;
  TEXT = 6;
  DATE = 7;
  STRUCT = 8;
  ARRAY = 9;
}

// 特征服务（源领域层，无 HTTP 注解）/ Feature service
service FeatureService {
  rpc List   (pagination.PagingRequest)   returns (ListFeatureResponse)  {}
  rpc Count  (pagination.PagingRequest)   returns (CountFeatureResponse) {}
  rpc Get    (GetFeatureRequest)          returns (Feature)              {}
  rpc Create (CreateFeatureRequest)       returns (google.protobuf.Empty) {}
  rpc Update (UpdateFeatureRequest)       returns (google.protobuf.Empty) {}
  rpc Delete (DeleteFeatureRequest)       returns (google.protobuf.Empty) {}

  // 按特征类型查询（左侧树选类型 → 右侧列表）/ List by feature type
  rpc ListByType (ListFeatureByTypeRequest) returns (ListFeatureResponse) {}

  // 校验 spec（不落库，返回校验结果）/ Validate spec without persisting
  rpc ValidateSpec (ValidateFeatureSpecRequest) returns (ValidateFeatureSpecResponse) {}
}

// ===== 特征主消息 =====
message Feature {
  optional uint32 id = 1 [
    json_name = "id",
    (gnostic.openapi.v3.property) = {description: "特征ID / Feature ID"}
  ];

  optional FeatureType feature_type = 2 [
    json_name = "featureType",
    (gnostic.openapi.v3.property) = {description: "特征类型 PROPERTY/EVENT/SERVICE/RELATION / Feature type"}
  ];

  optional string code = 3 [
    json_name = "code",
    (gnostic.openapi.v3.property) = {description: "清单编码，如 P-RUN-0001（不可变）/ Code, immutable"}
  ];

  optional string identifier = 4 [
    json_name = "identifier",
    (gnostic.openapi.v3.property) = {description: "程序标识符，如 powerSwitch / Program identifier"}
  ];

  optional string name = 5 [
    json_name = "name",
    (gnostic.openapi.v3.property) = {description: "中文名 / Name (zh)"}
  ];
  optional string name_en = 6 [
    json_name = "nameEn",
    (gnostic.openapi.v3.property) = {description: "英文名 / Name (en)"}
  ];
  optional string description = 7 [
    json_name = "description",
    (gnostic.openapi.v3.property) = {description: "说明 / Description"}
  ];
  optional string applicable_scope = 8 [
    json_name = "applicableScope",
    (gnostic.openapi.v3.property) = {description: "适用设备范围，如 冷机/锅炉 / Applicable device scope"}
  ];

  // ===== 特化抽取列（高频筛选）=====
  optional DataType data_type = 10 [
    json_name = "dataType",
    (gnostic.openapi.v3.property) = {description: "property 数据类型 / Property data type"}
  ];
  optional AccessMode access_mode = 11 [
    json_name = "accessMode",
    (gnostic.openapi.v3.property) = {description: "property 访问模式 R/RW / Property access mode"}
  ];
  optional EventLevel event_level = 12 [
    json_name = "eventLevel",
    (gnostic.openapi.v3.property) = {description: "event 级别 / Event level"}
  ];
  optional CallMode call_mode = 13 [
    json_name = "callMode",
    (gnostic.openapi.v3.property) = {description: "service 调用模式 / Service call mode"}
  ];
  optional string relation_type = 14 [
    json_name = "relationType",
    (gnostic.openapi.v3.property) = {description: "relation 关系类型，如 derivedFrom / Relation type"}
  ];

  // ===== 差异容器：spec（四类结构化约束，强类型 oneof）=====
  optional FeatureSpec spec = 20 [
    json_name = "spec",
    (gnostic.openapi.v3.property) = {description: "特征结构化约束（FeatureSpec oneof，按 featureType 取分支）/ Structured spec"}
  ];

  optional bool is_enabled = 30 [
    json_name = "isEnabled",
    (gnostic.openapi.v3.property) = {description: "是否启用 / Is enabled"}
  ];
  optional uint32 sort_order = 31 [
    json_name = "sortOrder",
    (gnostic.openapi.v3.property) = {description: "排序号 / Sort order"}
  ];

  optional uint32 tenant_id = 100 [
    json_name = "tenantId",
    (gnostic.openapi.v3.property) = {description: "租户ID，0=系统全局 / Tenant ID, 0=system global"}
  ];
  optional string tenant_name = 101 [
    json_name = "tenantName",
    (gnostic.openapi.v3.property) = {description: "租户名称 / Tenant name"}
  ];

  optional uint32 created_by = 200 [json_name = "createdBy", (gnostic.openapi.v3.property) = {description: "创建者 / Created by"}];
  optional uint32 updated_by = 201 [json_name = "updatedBy", (gnostic.openapi.v3.property) = {description: "更新者 / Updated by"}];
  optional uint32 deleted_by = 202 [json_name = "deletedBy", (gnostic.openapi.v3.property) = {description: "删除者 / Deleted by"}];

  optional google.protobuf.Timestamp created_at = 300 [json_name = "createdAt", (gnostic.openapi.v3.property) = {description: "创建时间 / Created at"}];
  optional google.protobuf.Timestamp updated_at = 301 [json_name = "updatedAt", (gnostic.openapi.v3.property) = {description: "更新时间 / Updated at"}];
  optional google.protobuf.Timestamp deleted_at = 302 [json_name = "deletedAt", (gnostic.openapi.v3.property) = {description: "删除时间 / Deleted at"}];
}

message ListFeatureResponse  { repeated Feature items = 1; uint64 total = 2; }
message CountFeatureResponse { uint64 count = 1; }

message GetFeatureRequest {
  oneof query_by {
    uint32 id = 1;
    string code = 2;
    string identifier = 3;   // 支持按 identifier 查询
  }
  optional google.protobuf.FieldMask view_mask = 100 [json_name = "viewMask"];
}

message CreateFeatureRequest { Feature data = 1; }

message UpdateFeatureRequest {
  uint32 id = 1;
  Feature data = 2;
  google.protobuf.FieldMask update_mask = 3 [
    (gnostic.openapi.v3.property) = {description: "要更新的字段列表 / Fields to update"},
    json_name = "updateMask"
  ];
  optional bool allow_missing = 4 [
    (gnostic.openapi.v3.property) = {description: "若为 true，资源不存在则新增 / If true, insert when missing"},
    json_name = "allowMissing"
  ];
}

message DeleteFeatureRequest { repeated uint32 ids = 1; }

// 按类型查询 / List by type
message ListFeatureByTypeRequest {
  FeatureType feature_type = 1 [json_name = "featureType"];
  optional bool only_enabled = 2 [json_name = "onlyEnabled"];
  optional string applicable_scope = 3 [json_name = "applicableScope"];
  // 支持分页参数（复用 pagination）
}

// 校验 spec / Validate spec
message ValidateFeatureSpecRequest {
  FeatureType feature_type = 1 [json_name = "featureType"];
  FeatureSpec spec = 2 [json_name = "spec"];
}

message ValidateFeatureSpecResponse {
  bool valid = 1 [json_name = "valid"];
  repeated string errors = 2 [json_name = "errors"];   // 校验失败原因列表
}
```

> **关于 spec 用强类型 `FeatureSpec` oneof**：proto 用 `FeatureSpec` oneof（[10 章 §3](./10-特征参数与spec设计.md)）直接作为 Feature.spec 的类型，编译期类型安全。oneof 在 JSON 序列化后是带判别字段的普通 JSON 对象，前端按 `featureType` 取对应分支渲染表单。Ent schema 也用 `field.JSON("spec", &thingmodelV1.FeatureSpec{})` 强类型目标（对齐项目 `menu.meta` 实践）。API 强类型 + DB 强类型，校验器直接操作 oneof 分支，无需额外转换层。

### 2.3 spec 消息定义（`feature_spec.proto` 或并入 `feature.proto`）

[10 章 §3](./10-特征参数与spec设计.md) 定义的 `FeatureSpec`（oneof）及其四个分支 `PropertySpec`/`EventSpec`/`ServiceSpec`/`RelationSpec`，加上公共子结构 `ParamSpec`/`UnitRef`/`EntityRef`/`ValueConstraints`/`EnumItem`/`BoolLabels`/`ArraySpec`，**必须与 Feature 消息同文件或被 import**（因为 Feature.spec 引用 FeatureSpec）。建议并入 `feature.proto` 同文件，避免 import 循环。

> 这样设计的好处：proto 全程强类型（API + DB + 校验共享同一类型定义），前后端编译期检查 spec 结构，校验器直接 `switch spec.GetSpec().(type)` 分派。

---

## 3. BFF 层（REST）

### 3.1 `i_feature.proto`

```protobuf
syntax = "proto3";
package admin.service.v1;

import "google/api/annotations.proto";
import "google/protobuf/empty.proto";
import "pagination/v1/pagination.proto";
import "thingmodel/service/v1/feature.proto";

// 特征管理服务（BFF 层 REST）/ Feature admin service
service FeatureService {
  // 分页查询特征列表 / List features
  rpc List (pagination.PagingRequest) returns (thingmodel.service.v1.ListFeatureResponse) {
    option (google.api.http) = {
      get: "/admin/v1/thingmodel/features"
    };
  }

  // 查询特征详情 / Get feature
  rpc Get (thingmodel.service.v1.GetFeatureRequest) returns (thingmodel.service.v1.Feature) {
    option (google.api.http) = {
      get: "/admin/v1/thingmodel/features/{id}"
      additional_bindings { get: "/admin/v1/thingmodel/features/code/{code}" }
      additional_bindings { get: "/admin/v1/thingmodel/features/identifier/{identifier}" }
    };
  }

  // 创建特征 / Create feature
  rpc Create (thingmodel.service.v1.CreateFeatureRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      post: "/admin/v1/thingmodel/features"
      body: "*"
    };
  }

  // 更新特征 / Update feature
  rpc Update (thingmodel.service.v1.UpdateFeatureRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      put: "/admin/v1/thingmodel/features/{id}"
      body: "*"
    };
  }

  // 删除特征 / Delete feature
  rpc Delete (thingmodel.service.v1.DeleteFeatureRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      delete: "/admin/v1/thingmodel/features"
    };
  }

  // 按特征类型查询（左侧树选类型）/ List features by type
  rpc ListByType (thingmodel.service.v1.ListFeatureByTypeRequest) returns (thingmodel.service.v1.ListFeatureResponse) {
    option (google.api.http) = {
      get: "/admin/v1/thingmodel/features/types/{feature_type}"
    };
  }

  // 校验 spec（不落库）/ Validate spec
  rpc ValidateSpec (thingmodel.service.v1.ValidateFeatureSpecRequest) returns (thingmodel.service.v1.ValidateFeatureSpecResponse) {
    option (google.api.http) = {
      post: "/admin/v1/thingmodel/features:validateSpec"
      body: "*"
    };
  }
}
```

---

## 4. REST 路由总表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/v1/thingmodel/features` | 特征分页列表（支持 `featureType`/`dataType`/`eventLevel` 过滤） |
| GET | `/admin/v1/thingmodel/features/{id}` | 特征详情（按 ID） |
| GET | `/admin/v1/thingmodel/features/code/{code}` | 特征详情（按 code） |
| GET | `/admin/v1/thingmodel/features/identifier/{identifier}` | 特征详情（按 identifier） |
| POST | `/admin/v1/thingmodel/features` | 新建特征 |
| PUT | `/admin/v1/thingmodel/features/{id}` | 更新特征 |
| DELETE | `/admin/v1/thingmodel/features?ids=1,2,3` | 批量删除特征 |
| GET | `/admin/v1/thingmodel/features/types/{feature_type}` | 某类型下全部特征（左侧树联动） |
| POST | `/admin/v1/thingmodel/features:validateSpec` | 校验 spec（前端表单实时校验用） |

> 路由风格与单位管理一致：`/admin/v1/thingmodel/features` 为主资源，`:validateSpec` 用冒号自定义动词（与单位 `/units:convert` 同风格）。

---

## 5. 命名与字段编号约定（与 unit 对齐）

| 段位 | 用途 | 说明 |
|------|------|------|
| `1~8` | 核心公共字段 | id、feature_type、code、identifier、name、name_en、description、applicable_scope |
| `10~14` | 特化抽取列 | data_type、access_mode、event_level、call_mode、relation_type |
| `20` | spec 容器 | JSON 结构化约束 |
| `30~31` | is_enabled / sort_order | 治理字段 |
| `100~101` | tenant_id / tenant_name | 租户 |
| `200~202` | created_by / updated_by / deleted_by | 审计 |
| `300~302` | created_at / updated_at / deleted_at | 时间戳 |

> 与单位 proto（核心 1~6、换算 10~18、审计 100+/200+）同构，段位含义一致。

---

## 6. 生成命令

```bash
# 后端 Go 代码 + OpenAPI
cd backend && make api && make openapi

# 三端 TS 客户端
make ts
```

生成产物（禁止手改）：
- `backend/api/gen/go/thingmodel/service/v1/feature.pb.go`、`feature_grpc.pb.go`
- `backend/api/gen/go/admin/service/v1/i_feature_http.pb.go`
- `frontend/admin/react/src/api/generated/admin/service/v1/index.ts`（新增 `featureservicev1_*` 类型与 client）

---

## 7. 查询能力（分页参数支持）

`List` 接口通过 `pagination.PagingRequest` 的搜索参数支持：

| 参数 | 说明 | 实例 |
|------|------|------|
| `featureType` | 按类型过滤 | `featureType=PROPERTY` |
| `dataType` | property 下按数据类型 | `dataType=DOUBLE` |
| `eventLevel` | event 下按级别 | `eventLevel=ALERT` |
| `applicableScope` | 按适用设备 | `applicableScope=制冷机组` |
| `keyword` | code/name/identifier 模糊 | `keyword=温度` |

> 这些作为 `PagingRequest.SearchParams`（项目通用分页机制，参照 dict/unit 的 List 实现）传入，Repo 层 `ListWithPaging` 自动构建 where。
