# 10 · 特征参数与 spec 设计 ⭐

> 本章是"特征属性管理"的**类型学核心**：定义属性的数据类型体系、事件/服务参数的结构化表达、struct/array 的递归嵌套规则，以及 relation 的本体语义模型。
> 与 [03 单位换算引擎](./03-单位换算引擎设计.md) 之于单位管理的地位相同——这是把《特征清单》的复杂性收敛为一套可校验 schema 的关键。

---

## 1. 数据类型体系（property 的 dataType）

完全对齐《特征清单》§1.2 与单位管理文档引用的阿里云 TSL + 华为云 IoTDA 规范：

| 数据类型 | 标识 | spec 约束字段 | 对应清单规范 |
|----------|------|--------------|-------------|
| 整数 | `INT` | `constraints.{min,max,step}` + `unit` | 32位整数 |
| 单精度浮点 | `FLOAT` | `constraints.{min,max,step}` + `unit` | 单精度 |
| 双精度浮点 | `DOUBLE` | `constraints.{min,max,step}` + `unit` | 双精度（工程主用） |
| 布尔 | `BOOL` | `boolLabels.{false,true}` | 0/1，各有描述 |
| 枚举 | `ENUM` | `enumItems[]`（值→描述） | 整数值枚举 |
| 字符串 | `TEXT` | `textMaxLength` | 最大长度 |
| 时间戳 | `DATE` | — | UTC 毫秒时间戳字符串 |
| 结构体 | `STRUCT` | `structFields[]`（递归） | 嵌套上述 7 种 |
| 数组 | `ARRAY` | `arraySpec.{size, element}` | 元素可为 int/float/double/text/struct |

### 1.1 为什么用 double 而非 float

工程实测值（温度、压力、功率）精度需求在 double 范围。清单中"双精度浮点数"占绝大多数（温度、压力、流量、功率、能效……），float 仅作兼容保留。proto 层统一 `double`（float64），DB 层 `DOUBLE`。

### 1.2 数据类型与特化列的关系

`data_type` 是从 spec 提升的特化列（[09 章 §1.1](./09-特征数据模型设计.md)），原因：

1. **列表筛选高频**：前端"属性"Tab 下需按 data_type 过滤（"只看 double 类属性"）。
2. **校验前置**：Service 校验器先看 data_type 再决定校验哪组字段（enum→enumItems、struct→structFields）。
3. **冗余一致性**：spec.dataType 与列 data_type 必须一致，由 Service 写入时同步（见 [12 章](./12-特征后端实现设计.md) 校验器）。

---

## 2. 参数结构（Param）

事件输出参数、服务输入/输出参数、struct 子字段，**共用同一个参数结构 ParamSpec**：

```jsonc
{
  "key": "temperature",           // 参数键（程序用，camelCase）
  "name": "温度",                  // 参数中文名（展示用）
  "dataType": "double",           // 数据类型（同 §1 体系）
  "unit": { "unitCode":"celsius", "unitSymbol":"℃" },  // 单位引用（可选）
  "constraints": { "min":-20, "max":60, "step":0.5 },  // 约束（按 dataType）
  "enumItems": [...],             // dataType=enum 时
  "boolLabels": {...},            // dataType=bool 时
  "textMaxLength": 256,           // dataType=text 时
  "structFields": [...],          // dataType=struct 时（递归 ParamSpec）
  "arraySpec": { "size":10, "element":{...} },  // dataType=array 时
  "required": true,               // 是否必填（service 输入参数用）
  "defaultValue": 25              // 默认值（可选）
}
```

**关键设计**：ParamSpec 是**自递归**的——struct 的子字段是 ParamSpec，array 的元素是 ParamSpec，可无限嵌套。这覆盖了清单中最复杂的参数形态（如"冷冻水进出水温度 struct"含 inlet/outlet 两个 double 子字段）。

### 2.1 单位引用的轻量化

ParamSpec 的 `unit` 只存 `{unitCode, unitSymbol}`（冗余符号便于展示），**不存 unitId**（避免参数层维护外键一致性）。仅 property 的 spec.unit 存 `unitId`（顶层属性才维护单位引用计数）。

> 即：属性的顶层单位才"正式引用"单位表（影响 reference_count）；参数级单位仅作展示标注，不引用。

### 2.2 参数结构对照清单

| 清单表述 | ParamSpec 表达 |
|----------|----------------|
| `temperature(double, ℃, -20~60)` | `{key:temperature, dataType:double, unit:{unitCode:celsius,unitSymbol:℃}, constraints:{min:-20,max:60}}` |
| `runMode(enum: 0=停机,1=手动,2=自动,3=远程)` | `{key:runMode, dataType:enum, enumItems:[{value:0,label:停机},{value:1,label:手动},{value:2,label:自动},{value:3,label:远程}]}` |
| `historyData(array<struct>)` | `{key:historyData, dataType:array, arraySpec:{element:{dataType:struct, structFields:[...]}}}` |
| `ratedChilledWaterTemp(struct: inlet(double,℃,0~40), outlet(double,℃,0~40))` | `{key:ratedChilledWaterTemp, dataType:struct, structFields:[{key:inlet,dataType:double,unit:{℃},constraints:{min:0,max:40}},{key:outlet,...}]}` |
| `errorType(enum: 0=超时,1=校验错误,2=连接断开)` | `{key:errorType, dataType:enum, enumItems:[...]}` |

---

## 3. 四类 spec 的 proto 结构（强类型契约）

proto 层定义强类型 `FeatureSpec` oneof（与 [11 章](./11-特征API与Proto设计.md) 的枚举共用），保证编译期结构检查。这是"DB 存 JSON + proto/API/校验全程强类型"的结合：

```protobuf
// 特征 spec（按 feature_type 取 oneof 分支）
message FeatureSpec {
  oneof spec {
    PropertySpec  property  = 1;
    EventSpec     event     = 2;
    ServiceSpec   service   = 3;
    RelationSpec  relation  = 4;
  }
}

// ===== 属性 =====
message PropertySpec {
  DataType data_type = 1 [json_name="dataType"];        // 复用 DataType 枚举
  AccessMode access_mode = 2 [json_name="accessMode"];  // 复用 AccessMode 枚举
  string category = 3;                                  // runtime/measurement/setting/rated/statistic/environment
  UnitRef unit = 4;                                     // 单位引用（顶层属性才正式引用）
  ValueConstraints constraints = 5;
  repeated EnumItem enum_items = 6 [json_name="enumItems"];
  BoolLabels bool_labels = 7 [json_name="boolLabels"];
  int32 text_max_length = 8 [json_name="textMaxLength"];
  repeated ParamSpec struct_fields = 9 [json_name="structFields"];
  ArraySpec array_spec = 10 [json_name="arraySpec"];
  bool is_rated = 11 [json_name="isRated"];
}

// ===== 事件 =====
message EventSpec {
  EventLevel level = 1;                                 // 复用 EventLevel 枚举
  repeated ParamSpec output_params = 2 [json_name="outputParams"];
  string trigger_condition = 3 [json_name="triggerCondition"];
  int32 severity = 4;
}

// ===== 服务 =====
message ServiceSpec {
  CallMode call_mode = 1 [json_name="callMode"];        // 复用 CallMode 枚举
  repeated ParamSpec input_params = 2 [json_name="inputParams"];
  repeated ParamSpec output_params = 3 [json_name="outputParams"];
  int32 timeout = 4;
}

// ===== 关系（本体）=====
message RelationSpec {
  string relation_type = 1 [json_name="relationType"]; // derivedFrom/partOf/feeds/...
  EntityRef source = 2;
  EntityRef target = 3;
  string cardinality = 4;                              // oneToOne/oneToMany/manyToOne/manyToMany
  bool directional = 5;
  map<string, string> properties = 6;                  // 关系自身属性
}

// ===== 公共子结构 =====
message ParamSpec {
  string key = 1;
  string name = 2;
  DataType data_type = 3 [json_name="dataType"];       // 复用 DataType 枚举
  UnitRef unit = 4;
  ValueConstraints constraints = 5;
  repeated EnumItem enum_items = 6 [json_name="enumItems"];
  BoolLabels bool_labels = 7 [json_name="boolLabels"];
  int32 text_max_length = 8 [json_name="textMaxLength"];
  repeated ParamSpec struct_fields = 9 [json_name="structFields"];  // 递归
  ArraySpec array_spec = 10 [json_name="arraySpec"];
  bool required = 11;
  string default_value = 12 [json_name="defaultValue"];
}

message UnitRef {
  uint32 unit_id = 1 [json_name="unitId"];      // 仅 property 顶层用
  string unit_code = 2 [json_name="unitCode"];
  string unit_symbol = 3 [json_name="unitSymbol"];
}

message EntityRef {
  string kind = 1;           // feature / external
  uint32 id = 2;             // kind=feature 时
  string code = 3;
  string identifier = 4;
  string type = 5;           // kind=external 时：device/space/product
}

message ValueConstraints {
  double min = 1;
  double max = 2;
  double step = 3;
  string default_value = 4 [json_name="defaultValue"];
}

message EnumItem { int32 value = 1; string label = 2; }
message BoolLabels { string false = 1; string true = 2; }
message ArraySpec { int32 size = 1; ParamSpec element = 2; }
```

> **存取约定**：Service 层收到 `Feature` 消息后，将 `spec` 字段（proto `FeatureSpec`）`Marshal` 为 JSON 存 DB；读取时从 DB JSON `Unmarshal` 回 `FeatureSpec`。DB 列用 `map[string]any`，API 用强类型 proto。

---

## 4. 属性类别体系（category）

对齐《特征清单》§1.6：

| category | 含义 | 默认 access_mode | 示例 |
|----------|------|------------------|------|
| `runtime` | 运行状态 | R | 开关状态、运行模式、故障代码 |
| `measurement` | 测量参数 | R | 温度、压力、流量、湿度 |
| `setting` | 设定参数 | RW | 设定温度、设定压力、时间计划 |
| `rated` | 额定参数（铭牌） | R | 额定功率、额定电压（isRated=true） |
| `statistic` | 统计参数 | R | 累计运行时间、累计能耗 |
| `environment` | 环境参数 | R | 环境温度、环境湿度 |

**category 与 isRated 的关系**：`category=rated` 时 `isRated` 应为 true（冗余标记，便于快速筛选铭牌参数）。两者一致性由校验器保证。

---

## 5. 关系语义模型（本体 Ontology）

关系是本设计区别于主流物模型三维标准的核心扩展。其语义模型如下：

### 5.1 关系三元组

每条 relation 特征本质是一个**本体三元组**：

```
( source ) ──[ relationType ]──▶ ( target )
```

- **source/target**：`EntityRef`，可为同表特征（`kind=feature`）或外部实体（`kind=external`）。
- **relationType**：谓词，定义关系语义。
- **cardinality**：基数约束。
- **directional**：是否有向。

### 5.2 EntityRef 的两种 kind

| kind | 含义 | 定位方式 | 本期校验 |
|------|------|----------|----------|
| `feature` | 指向同表 thing_feature | `id`（或 code/identifier） | **强校验**：目标须存在 |
| `external` | 指向外部实体（未来 device/space/product） | `type` + `code` | **弱校验**：本期不校验存在性（表未建） |

> 这种"表内自引用 + 外部弱引用"的设计，让关系能表达"属性A derivedFrom 属性B"（表内），也能表达"设备 locatedIn 楼层"（跨表，未来），而无需现在建 device/space 表。

### 5.3 关系类型规划（relation_type 词表）

源自建筑设备 ontology 与 GB/T 51269 BIM 关系实践：

| relation_type | 语义 | 方向性 | 典型 cardinality | 示例 |
|---------------|------|:---:|------------------|------|
| `partOf` | 组成（整体-部分） | 有向 | manyToOne | 压缩机 partOf 制冷机组 |
| `feeds` | 供给（能量/物质流） | 有向 | manyToOne | 冷冻水泵 feeds 空调末端 |
| `suppliedBy` | 被供给（feeds 反向） | 有向 | oneToMany | 冷机 suppliedBy 冷却塔 |
| `controls` | 控制 | 有向 | oneToMany | DDC controls 冷机 |
| `controlledBy` | 被控制（controls 反向） | 有向 | manyToOne | 冷机 controlledBy DDC |
| `derivedFrom` | 派生（属性推导） | 有向 | oneToOne | COP derivedFrom 制冷量÷功率 |
| `dependsOn` | 依赖（服务前置） | 有向 | manyToOne | 启动 dependsOn 自检通过 |
| `locatedIn` | 空间包含 | 有向 | manyToOne | 设备 locatedIn 楼层 |
| `monitors` | 监测 | 有向 | oneToMany | 传感器 monitors 设备 |
| `relatedTo` | 泛化关联 | 无向 | — | 备用泵 relatedTo 主泵 |

> **反向关系**：`feeds/suppliedBy`、`controls/controlledBy` 是成对的有向关系。本期不自动维护反向关系（避免冗余与不一致），由用户显式创建或前端提供"反向创建"快捷操作。

### 5.4 关系与特征类型的关联

relation 的 source/target 可指向任意类型的 feature，但实践中有典型模式：

| 关系 | 典型 source 类型 | 典型 target 类型 |
|------|------------------|------------------|
| derivedFrom | PROPERTY | PROPERTY |
| dependsOn | SERVICE | SERVICE/EVENT |
| partOf | PROPERTY（设备级）/ external | PROPERTY（设备级）/ external |
| feeds | external(device) | external(device) |

> 本期 relation 主要表达**属性间派生**与**服务间依赖**（source/target 为同表 feature）；设备级拓扑（feeds/partOf）涉及外部实体，本期可建定义但弱校验。

---

## 6. spec 校验规则总表

汇总四类 spec 的校验规则（Service 层 `feature_validator.go` 实现，见 [12 章](./12-特征后端实现设计.md)）：

| 规则 | 适用类型 | 条件 | 错误 |
|------|----------|------|------|
| V1 | PROPERTY | dataType 必填 | `FEATURE_SPEC_INVALID` |
| V2 | PROPERTY | accessMode 必填（R/RW） | `FEATURE_SPEC_INVALID` |
| V3 | PROPERTY | dataType=ENUM 时 enumItems 非空 | `FEATURE_ENUM_EMPTY` |
| V4 | PROPERTY | dataType=STRUCT 时 structFields 非空 | `FEATURE_STRUCT_EMPTY` |
| V5 | PROPERTY | dataType=ARRAY 时 arraySpec.element 非空 | `FEATURE_SPEC_INVALID` |
| V6 | PROPERTY | dataType=BOOL 时 boolLabels 建议填写 | 警告（不阻断） |
| V7 | PROPERTY | constraints.min ≤ constraints.max | `FEATURE_CONSTRAINT_INVALID` |
| V8 | PROPERTY | unit.unitId 存在且 unit 启用（查 unit_repo） | `UNIT_NOT_FOUND`（复用单位错误码） |
| V9 | PROPERTY | category=rated 时 isRated=true | `FEATURE_SPEC_INVALID` |
| V10 | EVENT | level 必填（INFO/ALERT/ERROR） | `FEATURE_SPEC_INVALID` |
| V11 | SERVICE | callMode 必填（ASYNC/SYNC） | `FEATURE_SPEC_INVALID` |
| V12 | SERVICE | callMode=SYNC 时建议填 timeout | 警告 |
| V13 | RELATION | relationType 必填 | `FEATURE_SPEC_INVALID` |
| V14 | RELATION | source/target 必填 | `FEATURE_SPEC_INVALID` |
| V15 | RELATION | source.kind=feature 时 target.id 须存在 | `FEATURE_RELATION_TARGET_NOT_FOUND` |
| V16 | RELATION | directional=false 时 source/target 顺序无关 | —（信息） |
| V17 | 特化列 | data_type/access_mode/event_level/call_mode/relation_type 与 spec 一致 | `FEATURE_SPEC_MISMATCH` |

> 这些规则集中在 `feature_validator.go`，按 feature_type 分派，单元测试必须覆盖（见 [14 章](./14-特征种子数据与实施计划.md) 验收标准）。

---

## 7. 为什么 spec 用 proto oneof + JSON 存储（而非纯 JSON 或纯结构化表）

| 方案 | 优点 | 缺点 | 本设计取舍 |
|------|------|------|-----------|
| 纯 JSON（无 proto 结构） | DB 最灵活 | 无编译期检查；前后端靠文档约定，易漂移 | **否决**：项目是 contract-first，必须有 proto 契约 |
| 纯结构化表（每参数一行子表） | DB 可索引/约束 | 参数/子字段/嵌套层级爆炸；struct 递归无法表达 | **否决**：复杂度不可控 |
| **proto oneof + JSON 存储** | API 强类型（proto）+ DB 灵活（JSON） | DB 层无法约束 JSON 内部（靠应用层校验器） | **采用**：兼顾契约与灵活，校验器兜底 |

这与单位管理"`factor+offset` 结构化数值 + `formula_expr` 文本"的取舍哲学一致：**能结构化标量的就抽列，复杂容器收敛到 JSON，强类型靠 proto + 校验器保证**。
