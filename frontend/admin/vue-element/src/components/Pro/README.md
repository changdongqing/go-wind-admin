# Pro 组件库

Pro 是一套基于 Vue 3 + Element Plus + vxe-table 的企业级业务组件库，提供**配置化驱动**的 CRUD 页面开发体验。

核心思想：**一个配置对象描述一个完整的列表页面**（搜索 → 工具栏 → 表格 → 分页 → 弹窗），无需重复编写模板代码。

## 目录结构

```
Pro/
├── ProForm/          # 动态表单
├── ProSearch/        # 搜索栏
├── ProToolbar/       # 工具栏
├── ProTable/         # 双引擎表格（vxe-table / el-table）
├── ProPagination/    # 分页
├── ProModal/         # 弹窗（Dialog / Drawer）
├── ProPage/          # 页面编排（组合以上所有组件）
├── composables/      # 状态管理 hooks
│   ├── useTableState.ts
│   └── useModalState.ts
├── constants/        # 默认值常量
├── index.ts          # 统一导出
└── README.md
```

## 快速开始

### 最小示例

```vue
<template>
  <ProPage :config="pageConfig" />
</template>

<script setup lang="ts">
import { ProPage, type ProPageConfig } from "@/components/Pro";

const pageConfig: ProPageConfig = {
  table: {
    listAction: async (params) => {
      const { page, pageSize, ...query } = params;
      const res = await fetchList({ page, pageSize, ...query });
      return { items: res.items, total: res.total };
    },
    columns: [
      { type: "index", label: "序号", width: 60 },
      { prop: "name", label: "名称", minWidth: 120 },
      { prop: "status", label: "状态", minWidth: 80 },
    ],
  },
};
</script>
```

### 完整示例（租户管理）

```vue
<template>
  <ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit">
    <!-- 自定义列插槽（通过 prop 或 slotName 匹配） -->
    <template #status="scope">
      <ElTag :type="scope.row.status === 1 ? 'success' : 'danger'">
        {{ scope.row.status === 1 ? '启用' : '禁用' }}
      </ElTag>
    </template>
  </ProPage>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { ProPage, type ProPageConfig } from "@/components/Pro";

const pageRef = ref();

const pageConfig: ProPageConfig = {
  permPrefix: "sys:manage_tenants",
  engine: "vxe", // 默认 vxe，可选 "element"

  search: {
    grid: true,
    fields: [
      { type: "input", label: "名称", field: "name", attrs: { clearable: true } },
      { type: "select", label: "状态", field: "status",
        options: [{ label: "启用", value: 1 }, { label: "禁用", value: 0 }],
      },
    ],
  },

  table: {
    listAction: async (params) => {
      const { page, pageSize, ...query } = params;
      const res = await fetchList({ page, pageSize, ...query });
      return { items: res.items, total: res.total };
    },
    deleteAction: async (ids) => await deleteItem({ id: ids }),
    toolbar: ["add"],
    defaultToolbar: ["refresh", "filter", "search"],
    columns: [
      { type: "index", label: "序号", width: 60 },
      { prop: "name", label: "名称", minWidth: 120 },
      { prop: "status", label: "状态", minWidth: 80, slotName: "status" },
      {
        prop: "createdAt", label: "创建时间", minWidth: 160,
        cellType: "date", dateFormat: "YYYY-MM-DD HH:mm:ss",
      },
      {
        prop: "action", label: "操作", fixed: "right", width: 150,
        cellType: "tool",
        buttons: [
          { name: "edit", text: "编辑" },
          { name: "delete", text: "删除", attrs: { type: "danger" } },
        ],
      },
    ],
  },
};

function handleAdd() { /* 打开新增弹窗 */ }
function handleEdit(row: any) { /* 打开编辑弹窗 */ }
</script>
```

## 使用层级

Pro 组件库设计为**渐进式**，从 0 模板的配置驱动到完全自由的原生写法，开发者可按需选择合适的抽象层级。

---

### Level 1：ProPage 开箱即用（标准 CRUD）

**适用场景**：标准增删改查页面，配置即页面，无需写模板。

只需提供一个配置对象，ProPage 自动编排搜索、工具栏、表格、分页、弹窗的完整生命周期。

```vue
<template>
  <ProPage :config="pageConfig" />
</template>

<script setup lang="ts">
import { ProPage, type ProPageConfig } from "@/components/Pro";
import { fetchListTenants, useDeleteTenant } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";

const { mutateAsync: deleteTenant } = useDeleteTenant();

const pageConfig: ProPageConfig = {
  permPrefix: "sys:manage_tenants",

  // 搜索配置
  search: {
    grid: true,
    fields: [
      { type: "input", label: "名称", field: "name", attrs: { clearable: true } },
      { type: "select", label: "状态", field: "status",
        options: [{ label: "启用", value: 1 }, { label: "禁用", value: 0 }],
      },
    ],
  },

  // 表格配置
  table: {
    listAction: async (query) => {
      const { page, pageSize, ...params } = query;
      const result = await fetchListTenants(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 10 },
          formValues: params,
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    deleteAction: async (ids: string) => {
      await deleteTenant({ id: ids as any });
    },
    toolbar: ["add"],
    defaultToolbar: ["refresh", "filter", "search"],
    columns: [
      { type: "index", label: "序号", width: 60 },
      { prop: "name", label: "名称", minWidth: 120 },
      { prop: "status", label: "状态", minWidth: 80, cellType: "tag",
        labelMap: { 1: "启用", 0: "禁用" }, tagType: "success",
      },
      { prop: "createdAt", label: "创建时间", minWidth: 160, cellType: "date" },
      { prop: "action", label: "操作", fixed: "right", width: 150, cellType: "tool",
        buttons: [
          { name: "edit", text: "编辑" },
          { name: "delete", text: "删除", attrs: { type: "danger" } },
        ],
      },
    ],
  },

  // 弹窗配置（新增/编辑共用）
  modal: {
    component: "drawer",
    drawer: { title: "租户管理", size: "500px" },
    fields: [
      { type: "input", label: "名称", field: "name", rules: [{ required: true, message: "请输入名称" }] },
      { type: "input", label: "编码", field: "code", rules: [{ required: true, message: "请输入编码" }] },
      { type: "textarea", label: "备注", field: "remark" },
    ],
    submitAction: async (data) => {
      // 提交逻辑
    },
  },
};
</script>
```

**要点**：
- 一个配置对象 = 一个完整的 CRUD 页面
- 搜索、工具栏、表格、分页、弹窗全自动联动
- 新增/编辑/删除/导出/导入均内置
- 零模板代码，纯声明式

---

### Level 3：插槽定制（复杂交互）

**适用场景**：基础配置不满足时，通过插槽覆盖特定列的渲染，或接管事件处理。

Level 1 的配置 + 插槽 + 事件组合，可以覆盖大多数复杂场景，而无需放弃 ProPage 的自动编排能力。

#### 自定义列渲染

通过 `slotName` 或 `prop` 匹配插槽，覆盖单元格渲染：

```vue
<template>
  <ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit">
    <!-- 类型列：自定义 Tag 配色 -->
    <template #type="scope">
      <ElTag size="small" effect="dark" round :color="typeToColor(scope.row.type)">
        {{ typeToName(scope.row.type) }}
      </ElTag>
    </template>

    <!-- 状态列：带 Switch 切换 -->
    <template #status="scope">
      <ElSwitch
        :model-value="scope.row.status === 1"
        @change="(val) => handleStatusChange(scope.row, val)"
      />
    </template>

    <!-- 展开行：复杂详情 -->
    <template #expand="scope">
      <div class="p-4">
        <ElDescriptions :column="2" border>
          <ElDescriptionsItem label="创建人">{{ scope.row.creator }}</ElDescriptionsItem>
          <ElDescriptionsItem label="创建时间">{{ scope.row.createdAt }}</ElDescriptionsItem>
        </ElDescriptions>
      </div>
    </template>

    <!-- 工具栏扩展 -->
    <template #toolbar-left>
      <ElButton type="warning" @click="handleBatchExport">批量导出</ElButton>
    </template>
  </ProPage>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { ElTag, ElSwitch, ElDescriptions, ElDescriptionsItem, ElButton } from "element-plus";
import { ProPage, type ProPageConfig } from "@/components/Pro";

const pageRef = ref();

const pageConfig: ProPageConfig = {
  table: {
    listAction: async (params) => { /* ... */ return { items: [], total: 0 }; },
    columns: [
      { type: "index", label: "序号", width: 60 },
      { type: "expand", label: "详情", width: 50 },
      { prop: "name", label: "名称", minWidth: 120 },
      { prop: "type", label: "类型", minWidth: 100, slotName: "type" },
      { prop: "status", label: "状态", minWidth: 80, slotName: "status" },
    ],
  },
};

function handleAdd() { /* 打开自定义弹窗 */ }
function handleEdit(row: any) { /* 打开自定义弹窗 */ }
function handleStatusChange(row: any, val: boolean) { /* 调用接口 */ }
function handleBatchExport() { /* 批量导出逻辑 */ }
</script>
```

#### 外部接管弹窗

当内置 ProModal 不满足需求时，可以不配置 `modal`，完全自行控制弹窗：

```vue
<template>
  <ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit">
    <!-- 列插槽 ... -->
  </ProPage>

  <!-- 自定义 Drawer，不受 ProPage 管控 -->
  <TenantDrawer ref="drawerRef" @success="handleSuccess" />
</template>

<script setup lang="ts">
import { ref } from "vue";
import { ProPage } from "@/components/Pro";
import TenantDrawer from "./tenant-drawer.vue";

const pageRef = ref();
const drawerRef = ref();

const pageConfig = {
  table: {
    listAction: async () => { /* ... */ },
    columns: [ /* ... */ ],
  },
  // 不配置 modal → ProPage 不会渲染内置弹窗
};

function handleAdd() {
  drawerRef.value?.open(); // 完全自行控制
}
function handleEdit(row: any) {
  drawerRef.value?.open(row);
}
function handleSuccess() {
  pageRef.value?.refresh(); // 手动刷新表格
}
</script>
```

**要点**：
- 插槽和配置可以**混合使用**，不冲突
- 不需要的功能不配置即可（如不配 `modal` 就不渲染弹窗）
- ProPage 的事件（`@add`、`@edit`）让你完全控制后续流程
- 通过 `pageRef.value?.refresh()` 可随时刷新数据

---

### Level 4：原子组件自由组合（完全控制）

**适用场景**：非标准页面布局、特殊交互流程，需要完全控制页面结构。

跳过 ProPage 编排层，直接使用原子组件（ProSearch、ProToolbar、ProTable、ProModal 等）+ Composables（useTableState、useModalState）自行组装。

```vue
<template>
  <div class="custom-page">
    <!-- 搜索栏 -->
    <ProSearch
      :fields="searchFields"
      :grid="true"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 中间区域：左侧树 + 右侧表格 -->
    <div class="flex gap-4 flex-1 min-h-0">
      <!-- 左侧部门树 -->
      <ElCard class="w-64 shrink-0" shadow="never">
        <ElTree :data="treeData" @node-click="handleNodeClick" />
      </ElCard>

      <!-- 右侧表格区 -->
      <div class="flex-1 flex flex-col min-w-0">
        <ProToolbar
          :left-buttons="toolbarButtons"
          @button-click="handleToolbarClick"
        />

        <ProTable
          ref="tableRef"
          :columns="columns"
          :data="tableState.data.value"
          :loading="tableState.loading.value"
          :total="tableState.pagination.total"
          :current-page="tableState.pagination.currentPage"
          :page-size="tableState.pagination.pageSize"
          @selection-change="tableState.handleSelectionChange"
          @current-change="(p) => { tableState.pagination.currentPage = p; tableState.fetch(params); }"
          @size-change="(s) => { tableState.pagination.pageSize = s; tableState.fetch(params, true); }"
        >
          <!-- 自定义列插槽 -->
          <template #status="scope">
            <ElSwitch :model-value="scope.row.status === 1" />
          </template>
        </ProTable>
      </div>
    </div>

    <!-- 弹窗 -->
    <ProModal
      v-model:visible="modalState.visible.value"
      :mode="modalState.mode.value"
      :config="modalConfig"
      :form-data="modalState.formData"
      @submit="tableState.fetch(params)"
    />
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from "vue";
import { ElCard, ElTree, ElSwitch } from "element-plus";
import {
  ProSearch,
  ProToolbar,
  ProTable,
  ProModal,
  useTableState,
  useModalState,
  type ProFormField,
  type ProTableColumn,
  type ProModalConfig,
  type ToolbarButton,
} from "@/components/Pro";

// === 搜索 ===
const searchFields: ProFormField[] = [
  { type: "input", label: "名称", field: "name" },
  { type: "select", label: "状态", field: "status",
    options: [{ label: "启用", value: 1 }, { label: "禁用", value: 0 }],
  },
];
const params = reactive<Record<string, any>>({});

function handleSearch(p: any) {
  Object.assign(params, p);
  tableState.fetch(params, true);
}
function handleReset(p: any) {
  Object.keys(params).forEach((k) => delete params[k]);
  tableState.fetch(params, true);
}

// === 表格 ===
const tableState = useTableState({
  indexAction: async (query) => { /* ... */ return { items: [], total: 0 }; },
});
tableState.fetch(params);

const columns: ProTableColumn[] = [
  { type: "selection" },
  { type: "index", label: "序号", width: 60 },
  { prop: "name", label: "名称", minWidth: 120 },
  { prop: "status", label: "状态", minWidth: 80, slotName: "status" },
];

// === 工具栏 ===
const toolbarButtons: ToolbarButton[] = [
  { name: "add", text: "新增", attrs: { type: "success" } },
  { name: "delete", text: "删除", attrs: { type: "danger" } },
];

// === 弹窗 ===
const modalState = useModalState("id");
const modalConfig: ProModalConfig = {
  component: "drawer",
  drawer: { title: "编辑", size: "500px" },
  fields: [
    { type: "input", label: "名称", field: "name", rules: [{ required: true, message: "必填" }] },
  ],
  submitAction: async (data) => { /* ... */ },
};

function handleToolbarClick(name: string) {
  if (name === "add") modalState.open("add");
}

// === 树 ===
const treeData = ref([]);
function handleNodeClick(node: any) {
  params.deptId = node.id;
  tableState.fetch(params, true);
}
</script>
```

**要点**：
- 各原子组件完全独立，可按需组合
- `useTableState` 和 `useModalState` 提供开箱即用的状态管理
- 自由安排页面布局（树 + 表格、多表格、嵌套布局等）
- 所有组件 Props/Events/Slots 完全透明，没有黑盒

---

### 层级对比

| | Level 1 | Level 3 | Level 4 |
|---|---|---|---|
| **核心** | `<ProPage :config />` | ProPage + 插槽 + 事件 | 原子组件 + Composables |
| **模板代码** | 零 | 少量（插槽模板） | 中等（自行组装） |
| **配置化程度** | 完全配置 | 配置为主、插槽补充 | 配置 + 自由编排 |
| **适用场景** | 标准 CRUD | 带特殊列/交互的 CRUD | 非标准布局/复杂交互 |
| **弹窗控制** | 内置 ProModal | 外部接管弹窗 | 完全自行控制 |
| **学习成本** | 最低 | 低 | 中 |

> **建议**：80% 的页面用 Level 1，15% 用 Level 3，5% 用 Level 4。

---

## 组件 API

### ProPage

页面编排组件，通过一个配置对象驱动搜索、工具栏、表格、分页、弹窗的完整生命周期。

#### Props

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `config` | `ProPageConfig` | **必填** | 页面配置对象，详见下方 |

#### Events

| 事件 | 参数 | 说明 |
|------|------|------|
| `search` | `(params)` | 搜索触发 |
| `reset` | `(params)` | 重置触发 |
| `add` | `()` | 新增按钮点击 |
| `edit` | `(row)` | 编辑按钮点击 |
| `view` | `(row)` | 查看按钮点击 |
| `delete` | `(ids)` | 删除完成 |
| `operate` | `({ name, row, $index })` | 操作列自定义按钮点击 |
| `toolbar` | `(name)` | 自定义工具栏按钮点击 |

#### Expose

| 方法 | 说明 |
|------|------|
| `refresh()` | 刷新表格数据 |
| `tableRef` | 表格组件引用 |
| `tableState` | useTableState 实例 |
| `modalState` | useModalState 实例 |
| `searchParams` | 当前搜索参数 |

---

### ProPageConfig

```typescript
interface ProPageConfig<T, Q> {
  permPrefix?: string;           // 权限前缀，如 "sys:manage_tenants"
  engine?: "vxe" | "element";   // 表格引擎，默认 "vxe"
  rowKey?: string;               // 主键字段名，默认 "id"

  search?: {                     // 搜索配置（省略则不显示搜索栏）
    fields?: ProFormField[];
    isExpandable?: boolean;      // 是否支持展开/收起，默认 true
    showNumber?: number;         // 默认显示字段数，默认 3
    colon?: boolean;             // 标签后显示冒号
    grid?: boolean | "left" | "right"; // 自适应网格布局
  };

  table: {                       // 表格配置（必填）
    columns: ProTableColumn[];
    tableAttrs?: Record<string, any>; // 透传表格属性
    pagination?: boolean;        // 是否分页，默认 true
    toolbar?: Array<string | ToolsButton>;    // 左侧工具栏
    toolbarRight?: Array<string | ToolsButton>; // 右侧自定义按钮
    defaultToolbar?: Array<string | ToolsButton>; // 右侧图标按钮
    listAction: ListAction;      // 列表请求函数（必填）
    request?: { pageName: string; limitName: string }; // 分页参数名
    modifyAction?: (data) => Promise<any>; // 行内修改
    deleteAction?: (ids) => Promise<any>;  // 删除
    exportAction?: (params) => Promise<any>; // 导出
    importAction?: (file) => Promise<any>;   // 导入
    importTemplate?: string | (() => Promise<any>); // 导入模板
  };

  modal?: {                      // 弹窗配置（省略则不显示弹窗）
    component?: "dialog" | "drawer"; // 弹窗类型，默认 "dialog"
    dialog?: Partial<DialogProps>;
    drawer?: Partial<DrawerProps>;
    form?: Record<string, any>;
    colon?: boolean;
    fields: ProFormField[];
    beforeSubmit?: (data) => void;
    submitAction?: (data) => Promise<any>;
  };
}
```

---

### ProTable

双引擎表格，支持 vxe-table（默认）和 el-table。

#### Props

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `columns` | `ProTableColumn[]` | **必填** | 列配置 |
| `data` | `T[]` | `[]` | 表格数据 |
| `loading` | `boolean` | `false` | 加载状态 |
| `engine` | `"vxe" \| "element"` | `"vxe"` | 表格引擎 |
| `rowKey` | `string` | `"id"` | 行主键 |
| `table` | `Record<string, any>` | `{}` | 透传表格属性 |
| `pagination` | `boolean` | — | 是否显示分页 |
| `total` | `number` | `0` | 数据总条数 |
| `currentPage` | `number` | `1` | 当前页码 |
| `pageSize` | `number` | `20` | 每页条数 |
| `pageSizes` | `number[]` | `[10,20,30,50,100]` | 可选每页条数 |

#### Events

| 事件 | 参数 | 说明 |
|------|------|------|
| `selection-change` | `(rows)` | 选中行变化 |
| `modify` | `({ row, field, value })` | 行内编辑（switch） |
| `operate` | `({ name, row, $index })` | 操作列按钮点击 |
| `current-change` | `(page)` | 页码变化 |
| `size-change` | `(size)` | 每页条数变化 |

#### Expose

| 方法 | 说明 |
|------|------|
| `getSelectionRows()` | 获取选中行 |
| `clearSelection()` | 清空选中 |
| `toggleRowSelection(row, selected?)` | 切换行选中状态 |

---

### ProTableColumn

```typescript
interface ProTableColumn<T> {
  type?: "default" | "selection" | "index" | "expand"; // 列类型
  label?: string;                    // 列标题
  prop?: keyof T & string;           // 字段名
  width?: string | number;           // 列宽
  minWidth?: string | number;        // 最小列宽
  fixed?: "left" | "right" | boolean; // 固定列
  align?: "left" | "center" | "right";
  sortable?: boolean | "custom";     // 排序
  show?: boolean;                    // 是否显示，默认 true
  treeNode?: boolean;                // 树节点列

  // === 单元格渲染 ===
  cellType?: CellType;               // 渲染类型（见下表）
  slotName?: string;                 // 自定义插槽名

  // cellType: "image"
  imageWidth?: number;               // 默认 40
  imageHeight?: number;              // 默认 40
  // cellType: "tag"
  labelMap?: Record<string, any>;    // 值-标签映射
  tagType?: string;
  // cellType: "switch"
  activeValue?: any;                 // 默认 1
  inactiveValue?: any;               // 默认 0
  activeText?: string;
  inactiveText?: string;
  // cellType: "date"
  dateFormat?: string;               // 默认 "YYYY-MM-DD HH:mm:ss"
  // cellType: "price"
  pricePrefix?: string;              // 价格前缀，如 "¥"
  // cellType: "tool"
  buttons?: Array<{ name: string; text?: string; attrs?: any; visible?: (row) => boolean }>;

  attrs?: Record<string, any>;       // 透传原生属性
  initFn?: (item) => void;           // 初始化函数
  reserveSelection?: boolean;        // 跨页保留选中
}
```

#### cellType 类型

| 值 | 说明 | 附加属性 |
|----|------|----------|
| `text` | 默认文本（可省略） | — |
| `image` | 图片预览 | `imageWidth`, `imageHeight` |
| `tag` | 标签 | `labelMap`, `tagType` |
| `switch` | 开关（触发 modify 事件） | `activeValue`, `inactiveValue`, `activeText`, `inactiveText` |
| `date` | 日期格式化 | `dateFormat` |
| `link` | 超链接 | — |
| `price` | 价格 | `pricePrefix` |
| `percent` | 百分比 | — |
| `icon` | 图标 | — |
| `tool` | 操作列按钮 | `buttons` |
| `custom` | 自定义插槽 | `slotName` 或 `prop` 作为插槽名 |

---

### ProSearch

搜索栏组件，支持 inline/flex/grid 三种布局模式。

#### Props

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `fields` | `ProFormField[]` | **必填** | 搜索字段配置 |
| `colon` | `boolean` | `false` | 标签冒号 |
| `grid` | `boolean \| "left" \| "right"` | `false` | 自适应网格布局 |
| `inline` | `boolean` | `true` | 行内表单模式 |
| `isExpandable` | `boolean` | `true` | 支持展开/收起 |
| `showNumber` | `number` | `3` | 默认显示字段数 |

#### Events

| 事件 | 参数 | 说明 |
|------|------|------|
| `search` | `(params)` | 搜索 |
| `reset` | `(params)` | 重置 |
| `expand` | `(expanded)` | 展开/收起 |

---

### ProToolbar

工具栏组件，管理左侧操作按钮和右侧快捷工具。

#### Props

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `leftButtons` | `ToolbarButton[]` | `[]` | 左侧按钮 |
| `rightButtons` | `ToolbarButton[]` | `[]` | 右侧自定义按钮 |
| `defaultToolbar` | `("refresh" \| "filter" \| "search" \| "exports" \| "imports")[]` | `["refresh","filter","search"]` | 右侧图标按钮 |
| `permPrefix` | `string` | — | 权限前缀 |
| `visible` | `boolean` | `true` | 是否显示 |

#### Events

| 事件 | 参数 | 说明 |
|------|------|------|
| `button-click` | `(name, data?)` | 按钮点击 |
| `refresh` | `()` | 刷新 |
| `search` | `()` | 搜索切换 |
| `filter` | `()` | 筛选 |
| `export` | `()` | 导出 |
| `import` | `()` | 导入 |

---

### ProModal

弹窗组件，支持 Dialog 和 Drawer 两种模式，内置表单。

#### Props

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `visible` | `boolean` | **必填** | 是否显示 |
| `mode` | `"add" \| "edit" \| "view"` | `"add"` | 弹窗模式 |
| `config` | `ProModalConfig` | **必填** | 弹窗配置 |
| `formData` | `T` | **必填** | 表单数据（reactive 对象） |

#### Events

| 事件 | 参数 | 说明 |
|------|------|------|
| `update:visible` | `(boolean)` | 显示状态变化 |
| `submit` | `()` | 提交成功 |

---

### ProForm

动态表单，通过字段配置自动渲染表单组件。

#### Props

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `modelValue` | `T` | **必填** | 表单数据对象 |
| `fields` | `ProFormField[]` | **必填** | 字段配置 |
| `colon` | `boolean` | `false` | 标签冒号 |
| `disabled` | `boolean` | `false` | 禁用所有字段 |

#### Expose

| 方法 | 说明 |
|------|------|
| `validate()` | 表单校验 |
| `resetFields()` | 重置表单 |

---

### ProFormField

表单字段配置，被 ProForm、ProSearch、ProModal 共用。

```typescript
interface ProFormField<T> {
  type?: FormValueType;    // 组件类型（见下表）
  label: string;           // 标签文本
  field: keyof T & string; // 字段名
  tips?: string;           // 标签提示
  attrs?: Record<string, any>; // 组件属性（placeholder、clearable 等）
  options?: { label: string; value: any }[]; // 选项列表
  rules?: FormItemRule[];  // 校验规则
  initialValue?: any;      // 初始值
  slotName?: string;       // 自定义插槽名
  hidden?: boolean;        // 是否隐藏
  span?: number;           // 栅格占位（默认 24）
  events?: Record<string, Function>; // 组件事件
  initFn?: (item) => void; // 初始化回调
  api?: () => Promise<any[]>; // 异步数据源（api-tree-select）
}
```

#### FormValueType 类型

| 值 | 说明 |
|----|------|
| `input` | 输入框 |
| `textarea` | 文本域 |
| `select` | 下拉选择 |
| `radio` | 单选 |
| `checkbox` | 多选 |
| `switch` | 开关 |
| `date-picker` / `date` | 日期选择 |
| `time-picker` | 时间选择 |
| `time-select` | 时间选择（固定选项） |
| `input-number` / `number` | 数字输入 |
| `cascader` | 级联选择 |
| `tree-select` | 树选择 |
| `api-tree-select` | 异步树选择 |
| `input-tag` | 标签输入 |
| `icon-select` | 图标选择 |
| `custom` | 自定义插槽 |

---

### ProPagination

分页组件，自动计算 layout。

#### Props

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `total` | `number` | `0` | 总条数 |
| `showTotal` | `boolean` | `true` | 显示总条数 |
| `showSizes` | `boolean` | `true` | 显示页大小选择 |
| `showJump` | `boolean` | `true` | 显示跳转 |
| `background` | `boolean` | `true` | 按钮背景色 |
| `hideOnSinglePage` | `boolean` | `false` | 单页隐藏 |

---

## Composables

### useTableState

表格数据状态管理，处理分页、加载、选中。

```typescript
import { useTableState } from "@/components/Pro";

const tableState = useTableState({
  indexAction: (params) => fetchList(params), // 必填
  rowKey: "id",          // 默认 "id"
  pagination: true,      // 默认 true
  request: { pageName: "page", limitName: "pageSize" }, // 自定义参数名
});

// 返回值
tableState.data           // ShallowRef<T[]> 表格数据
tableState.loading        // Ref<boolean>    加载状态
tableState.pagination     // Reactive        分页状态 { currentPage, pageSize, total, ... }
tableState.selection      // ShallowRef<T[]> 选中行
tableState.showPagination // boolean         是否显示分页
tableState.fetch(params?, resetPage?) // 请求数据
tableState.handleSelectionChange(rows) // 处理选中
tableState.getSelectionIds() // 获取选中行 ID
```

### useModalState

弹窗状态管理。

```typescript
import { useModalState } from "@/components/Pro";

const modalState = useModalState("id"); // 传入 rowKey

// 返回值
modalState.visible  // Ref<boolean>        是否显示
modalState.mode     // Ref<"add"|"edit"|"view"> 弹窗模式
modalState.formData // Reactive             表单数据
modalState.open(mode, row?) // 打开弹窗
```

---

## 表格引擎

ProTable 内置两种渲染引擎，通过 `engine` prop 切换：

| 引擎 | 值 | 说明 |
|------|-----|------|
| vxe-table | `"vxe"` | **默认**，功能丰富（虚拟滚动、树形等） |
| el-table | `"element"` | Element Plus 原生表格，轻量 |

引擎差异自动映射：

| 功能 | vxe-table | el-table |
|------|-----------|----------|
| 复选框列 | `type: "checkbox"` | `type: "selection"` |
| 序号列 | `type: "seq"` | `type: "index"` |
| 列字段 | `field` | `prop` |
| 列标题 | `title` | `label` |
| 行索引 | `scope.rowIndex` | `scope.$index` |
| 获取选中 | `getCheckboxRecords()` | `getSelectionRows()` |
| 清空选中 | `clearCheckboxRow()` | `clearSelection()` |

---

## 工具栏按钮

### 内置按钮

ProPage 的 `toolbar` 和 `defaultToolbar` 支持字符串快捷方式：

```typescript
// 左侧工具栏
toolbar: ["add", "delete", "import", "export"]

// 右侧图标按钮
defaultToolbar: ["refresh", "filter", "search", "exports", "imports"]
```

### 自定义按钮

```typescript
toolbar: [
  {
    name: "custom-action",
    text: "自定义操作",
    perm: "custom",           // 权限标识，会拼接 permPrefix
    attrs: { type: "warning" }, // ElButton 属性
    visible: (row) => true,   // 动态显隐
  },
]
```

---

## 插槽

### ProPage 插槽

表格自定义列插槽通过 `prop` 或 `slotName` 匹配：

```vue
<ProPage :config="pageConfig">
  <!-- 通过 prop 匹配 -->
  <template #status="scope">
    <ElTag>{{ scope.row.status }}</ElTag>
  </template>

  <!-- 通过 slotName 匹配 -->
  <template #customSlot="scope">
    <span>{{ scope.row }} {{ scope.column }} {{ scope.index }}</span>
  </template>
</ProPage>
```

在 columns 中配置：

```typescript
columns: [
  { prop: "status", label: "状态", slotName: "status" },   // 或 cellType: "custom"
  { prop: "custom", label: "自定义", slotName: "customSlot" },
]
```

### ProToolbar 插槽

| 插槽名 | 说明 |
|--------|------|
| `left` | 左侧按钮区末尾 |
| `right` | 右侧按钮区末尾 |
| `filter` | 筛选弹出框内容 |

---

## 设计原则

1. **配置化驱动**：通过 TypeScript 类型安全的配置对象描述页面，减少模板代码
2. **按需使用**：ProPage 是编排层，各子组件可独立使用
3. **双引擎支持**：vxe-table / el-table 一键切换，API 统一
4. **插槽开放**：所有渲染点都支持插槽覆盖，配置化与自定义并存
5. **CURD 兼容**：API 设计参考现有 CURD 组件，迁移成本低
