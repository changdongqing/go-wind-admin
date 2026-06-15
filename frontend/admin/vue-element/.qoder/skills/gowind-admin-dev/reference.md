# GoWind Admin 开发参考手册

> 面向二开人员的速查手册。所有类型定义和 API 均来自脚手架真实代码。

---

## 1. API 两层架构

### 层级关系

```
gRPC 生成代码 (generated/)      ← 禁止修改
    ↓ 只导入类型（type import） + apiClient
ApiClient (client.ts)           ← 全局单例，懒加载各服务 Client
    ↓ 调用 apiClient.xxxService.Method()
Composable 层 (composables/)    ← Vue Query hooks + 枚举工具，面向组件
```

### 1.1 ApiClient 单例 (`src/api/client.ts`)

**职责**：全局唯一的 API 入口，通过 `ClientTransport` 适配 axios 请求，懒加载各服务 Client。

```typescript
// src/api/client.ts — 全局单例
import { type ClientTransport, createApiClient } from "@/api/generated/admin/service/v1";
import { requestApi } from "@/core/transport/rest";

const transport: ClientTransport = {
  unary(path, method, body, _meta) {
    return requestApi({ body, method, path });
  },
  serverStream(path, _meta) { throw new Error(...); },
  duplexStream(path, _meta) { throw new Error(...); },
};

export const apiClient = createApiClient(transport);
```

**使用方式**：
```typescript
import { apiClient } from "@/api/client";

// 通过属性访问器获取服务 Client（懒加载单例）
apiClient.userService.List(params);
apiClient.authenticationService.Login(request);
apiClient.positionService.Create({ data: { ... } });
```

**注意事项**：
- `requestApi` 是全局请求适配器，从 `@/core/transport/rest` 导入
- `apiClient` 的服务属性命名：驼峰服务名 + `Service` 后缀（如 `positionService`、`dictTypeService`）
- 新增服务无需手动注册，protobuf 重新生成后 `ApiClient` 类自动包含新属性

### 1.2 Composable 层 (`src/api/composables/`)

**职责**：基于 Vue Query 的 hooks，通过 `apiClient` 调用服务，提供响应式数据管理。

#### use* Hook 系列（组件 setup 中使用）

```typescript
import { apiClient } from "@/api/client";

// 列表查询
export function useListXxx(query: PaginationQuery, options?: UseQueryOptions) {
  return useQuery({ queryKey: ["listXxx", query], queryFn: () => apiClient.xxxService.List(query.toRawParams()), ...options });
}

// 单条查询
export function useGetXxx(req: GetXxxRequest, options?: UseQueryOptions) {
  return useQuery({ queryKey: ["getXxx", req], queryFn: () => apiClient.xxxService.Get(req), ...options });
}
```

#### fetch* 非 Hook 系列（事件处理、手动调用）

```typescript
export async function fetchListXxx(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listXxx", params],
    queryFn: () => apiClient.xxxService.List(params.toRawParams()),
    retry: 0,
  });
}
```

#### Mutation（创建/更新/删除）

```typescript
// 创建 — 参数用 { data: {...} } 包裹
export function useCreateXxx(options?: UseMutationOptions) {
  return useMutation({
    mutationFn: (values) => apiClient.xxxService.Create({ data: { ...values } as XxxType }),
    ...options,
  });
}

// 更新 — 必须使用 makeUpdateMask
export function useUpdateXxx(options?: UseMutationOptions) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      apiClient.xxxService.Update({ id, data: { ...values }, updateMask: makeUpdateMask(Object.keys(values ?? {})) }),
    ...options,
  });
}

// 删除
export function useDeleteXxx(options?: UseMutationOptions) {
  return useMutation({ mutationFn: (req) => apiClient.xxxService.Delete(req), ...options });
}
```

#### 枚举工具函数

```typescript
// 枚举列表 — computed + i18n
export const xxxStatusList = computed(() => [
  { value: "ACTIVE", label: t("enum.xxx.status.ACTIVE") },
  { value: "INACTIVE", label: t("enum.xxx.status.INACTIVE") },
]);

// 枚举值 → 显示名称
export function xxxStatusToName(status: string) {
  return xxxStatusList.value.find((item) => item.value === status)?.label ?? "";
}

// 枚举值 → 颜色
export function xxxStatusToColor(status: string) {
  return STATUS_COLOR_MAP[status] || STATUS_COLOR_MAP.DEFAULT;
}
```

---

## 2. Pro 组件库参考

### 2.1 组件总览

| 组件 | 用途 | 导入方式 |
|------|------|----------|
| `ProPage` | 页面编排（组合搜索+工具栏+表格+分页） | `import ProPage from "@/components/Pro/ProPage/index.vue"` |
| `ProSearch` | 搜索栏 | 由 ProPage 内部编排，通过 config.search 配置 |
| `ProToolbar` | 工具栏 | 由 ProPage 内部编排，通过 config.table 配置 |
| `ProTable` | 双引擎表格（vxe/el） | 由 ProPage 内部编排 |
| `ProPagination` | 分页 | 由 ProPage 内部编排 |
| `ProModal` | 弹窗 | 由 ProPage 内部编排或独立使用 |
| `ProForm` | 动态表单 | 独立使用 |
| `ProFileSelect` | 文件选择器 | 独立使用 |

**统一导出**：`import { ProPage, useProModal, injectProModalApi } from "@/components/Pro"`

### 2.2 ProPageConfig 配置

```typescript
// src/components/Pro/ProPage/types.ts
interface ProPageConfig<T = any, Q = any> {
  exportFilename?: string;       // 导出文件名
  engine?: "vxe" | "element";   // 表格引擎，默认 "vxe"
  rowKey?: string;               // 行唯一标识
  tableId?: string;              // 表格 ID（用于列配置持久化）
  pageKey?: string;              // 页面缓存 key

  search?: {
    fields?: ProFormField[];      // 搜索字段
    isExpandable?: boolean;       // 可展开/折叠
    showNumber?: number;          // 默认显示字段数
    colon?: boolean;              // 标签后冒号
    grid?: boolean | "left" | "right";  // 网格布局
  };

  table: {
    columns: ProTableColumn<T>[];   // 列配置（必填）
    tableAttrs?: Record<string, any>; // 透传给底层表格的属性
    pagination?: boolean;            // 是否分页，默认 true
    toolbar?: Array<ToolbarLeft | ToolsButton>;      // 左侧工具栏
    toolbarRight?: Array<ToolbarLeft | ToolsButton>;  // 右侧工具栏（如 ["add"]）
    defaultToolbar?: Array<ToolbarRight | ToolsButton>; // 默认工具栏
    listAction: ListAction<T, Q>;    // 列表数据请求（必填）
    request?: { pageName: string; limitName: string }; // 分页参数名
    modifyAction?: (data) => Promise<any>;  // 行内编辑
    deleteAction?: (ids: string) => Promise<any>;  // 删除
    exportsAction?: (queryParams: Q) => Promise<any[]>; // 导出
    importsAction?: (data: any[]) => Promise<any>; // 批量导入
    importAction?: (file: File) => Promise<any>;   // 单文件导入
    importTemplate?: string | (() => Promise<any>); // 导入模板
  };

  modal?: {
    component?: "dialog" | "drawer";
    dialog?: Partial<Omit<DialogProps, "modelValue">>;
    drawer?: Partial<Omit<DrawerProps, "modelValue">>;
    form?: Record<string, any>;
    colon?: boolean;
    fields: ProFormField<T>[];
    beforeSubmit?: (data: T) => Promise<T> | T;
    submitAction?: (data: T) => Promise<any>;
    afterSubmit?: () => void;
  };
}
```

### 2.3 ProTableColumn 列配置

```typescript
// src/components/Pro/ProTable/types.ts
interface ProTableColumn<T = any> {
  type?: "default" | "selection" | "index" | "expand";
  label?: string;
  prop?: keyof T & string;
  width?: string | number;
  minWidth?: string | number;
  fixed?: "left" | "right" | boolean;
  align?: "left" | "center" | "right";
  sortable?: boolean | "custom";
  resizable?: boolean;
  show?: boolean;
  treeNode?: boolean;

  // 渲染类型
  cellType?: "text" | "image" | "tag" | "switch" | "input" | "date" | "link" | "price" | "percent" | "icon" | "tool" | "custom";

  // cellType 扩展属性
  dateFormat?: string;          // cellType: "date" 时生效
  labelMap?: Record<string, any>; // cellType: "tag" 时的值映射
  tagType?: string;             // tag 类型
  tagEffect?: "light" | "dark" | "plain";
  imageWidth?: number;
  imageHeight?: number;
  pricePrefix?: string;

  // switch 属性
  activeValue?: any;
  inactiveValue?: any;
  activeText?: string;
  inactiveText?: string;

  // 操作列按钮
  buttons?: Array<{
    name: string;
    label?: string;
    icon?: string;
    auth?: string | string[];
    attrs?: Partial<ButtonProps>;
    visible?: (row: T) => boolean;
  }>;

  // 自定义插槽
  slotName?: string;

  // 搜索字段远程初始化
  initFn?: (item: any) => void;
  filterJoin?: string;
}
```

**常用列配置示例**：

```typescript
// 序号列
{ type: "index", label: $t("common.table.seq"), width: 60 }

// 普通文本列
{ prop: "name", label: $t("pages.xxx.name"), minWidth: 120 }

// 数字列（右对齐）
{ prop: "price", label: $t("pages.xxx.price"), width: 100, align: "right" }

// 自定义插槽列（用于 Tag、自定义渲染）
{ prop: "status", label: $t("common.table.status"), minWidth: 100, slotName: "status" }

// 日期列
{ prop: "createdAt", label: $t("common.table.createdAt"), minWidth: 160, cellType: "date", dateFormat: "YYYY-MM-DD HH:mm:ss" }

// 开关列
{ prop: "enabled", label: $t("common.table.enabled"), width: 100, cellType: "switch", activeValue: true, inactiveValue: false }

// 操作列
{
  prop: "action", label: $t("common.table.action"), fixed: "right", width: 150, cellType: "tool",
  buttons: [
    { name: "edit", label: $t("common.button.edit"), icon: "lucide:pen-line" },
    { name: "delete", label: $t("common.button.delete"), icon: "lucide:trash-2", attrs: { type: "danger" } },
  ],
}
```

### 2.4 ProFormField 搜索/表单字段配置

```typescript
// src/components/Pro/ProForm/types.ts
type FormValueType =
  | "input" | "textarea" | "select" | "radio" | "checkbox" | "switch"
  | "date-picker" | "time-picker" | "time-select" | "input-number"
  | "cascader" | "tree-select" | "api-tree-select" | "input-tag"
  | "custom-tag" | "icon-select" | "number" | "date" | "custom";

interface ProFormField<T = Record<string, any>> {
  type?: FormValueType;         // 组件类型，默认 input
  label: string;                 // 标签文本
  field: FieldKey<T>;            // 字段名
  tips?: string;                 // 标签提示
  attrs?: Record<string, any>;   // 组件属性
  options?: { label: string; value: any; disabled?: boolean }[]; // 选择项
  rules?: FormItemRule[];        // 校验规则
  initialValue?: any;            // 初始值
  slotName?: string;             // 自定义插槽名
  hidden?: boolean;              // 是否隐藏
  displayIf?: (model: T) => boolean; // 显隐联动
  span?: number;                 // 栅格占位
  col?: Partial<ColProps>;       // ElCol 配置
  events?: Record<string, (...args: any[]) => void>; // 组件事件
  initFn?: (field: ProFormField<T>) => void; // 初始化函数（异步加载数据）
  api?: () => Promise<any[]>;    // 异步数据源
}
```

**常用搜索字段示例**：

```typescript
// 输入框
{ type: "input", label: $t("pages.xxx.name"), field: "name", attrs: { placeholder: $t("common.placeholder.input"), clearable: true } }

// 下拉选择
{ type: "select", label: $t("common.table.status"), field: "status", attrs: { clearable: true }, options: statusList.value }

// 日期选择
{ type: "date-picker", label: $t("pages.xxx.date"), field: "date" }

// 树形选择（异步加载）
{
  type: "tree-select", label: $t("pages.xxx.org"), field: "orgId",
  attrs: {
    clearable: true, filterable: true, "default-expand-all": true,
    nodeKey: "id",
    props: { label: "name", value: "id", children: "children" },
  },
  initFn: async (item) => {
    const result = await fetchListXxx(new PaginationQuery({ formValues: { status: "ON" } }));
    item.attrs.data = result.items || [];
  },
}
```

### 2.5 ProPage 事件

```vue
<ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit" @delete="handleDelete">
  <!-- 自定义列插槽 -->
  <template #status="scope: any">
    <ElTag :color="statusToColor(scope.row.status)">{{ statusToName(scope.row.status) }}</ElTag>
  </template>
</ProPage>
```

- `add` — 点击新增按钮
- `edit(row)` — 点击编辑按钮，参数为行数据
- `delete(row)` — 点击删除按钮

**ProPage ref 方法**：
- `pageRef.value?.refresh()` — 刷新列表

---

## 3. PaginationQuery 工具类

```typescript
import { PaginationQuery, makeUpdateMask } from "@/core/transport/rest";

// 创建分页查询
const pq = new PaginationQuery({
  paging: { page: 1, pageSize: 10 },    // 分页参数
  formValues: { name: "test", status: "ON" }, // 搜索条件
  orderBy: ["-created_at"],              // 排序（负号=降序）
});

pq.toRawParams();  // 转换为 gRPC API 参数格式
// 返回: { page, pageSize, noPaging, fieldMask, orderBy: "[\"-created_at\"]", query: "{\"name\":\"test\",\"status\":\"ON\"}", sorting: undefined, offset: undefined, limit: undefined, token: undefined, filter: undefined, filterExpr: undefined }
```

**makeUpdateMask** — 生成字段更新掩码（自动追加 id）：
```typescript
import { makeUpdateMask } from "@/core/transport/rest";
makeUpdateMask(["name", "status"]);  // "name,status,id"
```

---

## 4. 国际化规范

### 4.1 翻译文件结构

```
locales/zh-CN/
├── common.json       # 通用文本（按钮、表单、表格、通知等）
├── enum.json         # 枚举值翻译（追加式）
├── routes.json       # 路由标题（追加式）
├── preferences.json  # 偏好设置
├── core.json         # 核心功能文本
├── validation.json   # 验证消息
└── pages/            # 页面级翻译（每模块一个文件）
    ├── position.json
    ├── user.json
    └── xxx.json
```

### 4.2 翻译 key 命名规范

| 分类 | 格式 | 示例 |
|------|------|------|
| 页面文本 | `pages.<module>.<field>` | `pages.position.name` |
| 枚举翻译 | `enum.<module>.<field>.<VALUE>` | `enum.position.type.REGULAR` |
| 路由标题 | `routes.<module>.<page>` | `routes.opm.position` |
| 通用文本 | `common.<category>.<key>` | `common.button.edit` |
| 表单占位 | `common.placeholder.input` / `common.placeholder.select` | — |
| 验证消息 | `common.validation.required` / `common.validation.selectRequired` | — |
| 通知消息 | `common.notification.createSuccess` / `common.notification.updateSuccess` | — |
| 弹窗标题 | `common.modal.create` / `common.modal.update` | 需传 `{ moduleName }` 参数 |

### 4.3 在代码中使用

```typescript
// 组件模板中（响应式）
$<t("pages.position.name")

// script setup 中（响应式场景）
import { $t } from "@/core/i18n";
const label = $t("pages.position.name");

// composable 顶层（非响应式场景）
import { i18n } from "@/core/i18n";
const t = i18n.global.t;
const label = t("enum.position.type.REGULAR");
```

---

## 5. 路由配置

### 5.1 路由文件位置

动态路由放在 `src/router/routes/modules/app/` 目录下，文件导出 `default` 路由数组，自动被 `import.meta.glob` 扫描加载。

### 5.2 meta 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `title` | `string` | i18n key 或直接文本 |
| `icon` | `string` | `lucide:` 前缀的 UnoCSS 图标名 |
| `order` | `number` | 菜单排序，数字越小越靠前 |
| `authority` | `string[]` | 权限标识数组 |
| `hideInMenu` | `boolean` | 是否隐藏菜单项 |
| `hideInTab` | `boolean` | 是否隐藏标签页 |
| `hideInBreadcrumb` | `boolean` | 是否隐藏面包屑 |
| `ignoreAccess` | `boolean` | 是否忽略权限检查 |
| `keepAlive` | `boolean` | 是否缓存组件 |

---

## 6. 常用导入速查

```typescript
// API 层
import { apiClient } from "@/api/client";  // Composable 内部使用
import { fetchListXxx, useCreateXxx, useUpdateXxx, useDeleteXxx, xxxStatusList } from "@/api/composables";  // 组件中使用

// Pro 组件
import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import { useProModal, injectProModalApi } from "@/components/Pro";

// 工具
import { PaginationQuery, makeUpdateMask } from "@/core/transport/rest";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";

// 通用枚举（已内置）
import { statusList, statusToName, statusToColor, enableList, enableBoolToColor } from "@/api/composables";
```

---

## 7. 构建与命令

```bash
pnpm dev              # 启动开发服务器
pnpm build            # 类型检查 + 生产构建
pnpm build-only       # 仅构建（不检查类型）
pnpm type-check       # TypeScript 类型检查
pnpm lint             # ESLint + Prettier + Stylelint
pnpm commit           # Git 提交（cz-git 交互式）
```

- Node 版本要求：`^20.19.0 || >=22.12.0`
- 包管理器：仅允许 pnpm（preinstall 检查）
- Git 提交信息：遵循 Conventional Commits 规范
