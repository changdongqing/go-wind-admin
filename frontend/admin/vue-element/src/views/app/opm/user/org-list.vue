<template>
  <div class="dept-container">
    <ElCard class="card-flat">
      <div class="toolbar-container">
        <!-- 租户选择器 -->
        <div v-if="!userViewStore.isTenantUser()" class="input-row">
          <span class="input-label">{{ $t("routes.tenant.member") }}</span>
          <ElSelect
            filterable
            clearable
            class="search-input"
            :placeholder="$t('common.input-search.placeholder')"
            v-model="selectedValue"
            :filter-method="filterOption"
            @change="handleTenantChanged"
          >
            <ElOption
              v-for="option in tenantOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </ElSelect>
        </div>

        <!-- 搜索框和工具栏 -->
        <div class="input-row">
          <span class="input-label">{{ $t("pages.org-unit.moduleName") }}</span>
          <ElInput
            class="search-input"
            clearable
            v-model="searchValue"
            :placeholder="$t('common.input-search.placeholder')"
          />
          <ElDropdown @command="handleToolbarClick">
            <ElButton type="text" :icon="More" />
            <template #dropdown>
              <ElDropdownMenu>
                <ElDropdownItem v-for="item in toolbarList" :key="item.value" :command="item">
                  {{ item.label }}
                </ElDropdownItem>
              </ElDropdownMenu>
            </template>
          </ElDropdown>
        </div>
      </div>
    </ElCard>

    <ElTree
      ref="treeRef"
      :data="treeData"
      :props="{ label: 'label', children: 'children' }"
      :expand-on-click-node="false"
      :default-expanded-keys="expandedKeys"
      :default-checked-keys="selectedKeys"
      node-key="key"
      highlight-current
      class="tree-container"
      @node-expand="handleExpandNode"
      @current-change="handleSelectNode"
    >
      <template #default="{ node }">
        <span class="tree-node-label">
          <span v-if="searchValue && String(node.label).indexOf(searchValue) > -1">
            {{ String(node.label).substring(0, String(node.label).indexOf(searchValue)) }}
            <span class="highlight-text">{{ searchValue }}</span>
            {{
              String(node.label).substring(
                String(node.label).indexOf(searchValue) + searchValue.length
              )
            }}
          </span>
          <span v-else>{{ node.label }}</span>
        </span>
      </template>
    </ElTree>
  </div>
</template>

<script lang="ts" setup>
import type { TreeNode } from "element-plus";

import { computed, onMounted, ref, watch } from "vue";

import {
  ElTree,
  ElCard,
  ElInput,
  ElSelect,
  ElOption,
  ElButton,
  ElDropdown,
  ElDropdownMenu,
  ElDropdownItem,
} from "element-plus";
import { More } from "@element-plus/icons-vue";
import { $t } from "@/i18n";

import { type identityservicev1_OrgUnit as OrgUnit } from "@/api/generated/admin/service/v1";
import { useUserViewStore } from "@/views/app/opm/user/user-view.state";

const userViewStore = useUserViewStore();

const toolbarList = [
  {
    value: "EXPAND_ALL",
    label: $t("common.tree.expand_all"),
    handler: handleMenuExpandAll,
  },
  {
    value: "COLLAPSE_ALL",
    label: $t("common.tree.collapse_all"),
    handler: handleMenuCollapseAll,
  },
  {
    value: "UNSELECT_ALL",
    label: $t("common.tree.unselect_all"),
    handler: handleMenuUnselectedAll,
  },
];

const expandedKeys = ref<(number | string)[]>([]);
const selectedValue = ref<string>("");
const searchValue = ref<string>("");
const autoExpandParent = ref<boolean>(true);

const treeData = ref<any[]>([]);
const selectedKeys = ref<(number | string)[]>([]);

const tenantOptions = computed(() =>
  (userViewStore.tenantList.items ?? []).map((t: any) => ({
    label: t.name ?? t.code ?? String(t.id ?? ""),
    value: String(t.id ?? ""),
    raw: t,
  }))
);

const filterOption = (input: string, option: any) => {
  const text = String(option.value ?? option.label ?? "").toLowerCase();
  return text.includes(input.toLowerCase());
};

/**
 * 递归转换树形数据（替代 mapTree）
 */
function mapTreeData(nodes: OrgUnit[], _parentId: number | null = null): any[] {
  return nodes.map((node) => ({
    ...node,
    key: `${node.parentId}-${node.id}`,
    label: node.name,
    id: node.id,
    children: node.children && node.children.length > 0 ? mapTreeData(node.children, node.id) : [],
  }));
}

/**
 * 获取组织单元列表
 */
async function fetchOrgUnits() {
  try {
    const response = await userViewStore.fetchOrgUnitList();
    treeData.value = mapTreeData(response.items ?? []);
  } catch (error) {
    console.error(error);
  }
}

async function fetch() {
  await fetchOrgUnits();
}

/**
 * 展开所有节点
 */
function handleMenuExpandAll() {
  const keys: (number | string)[] = [];

  const traverse = (nodes: any[] | undefined) => {
    if (!nodes || nodes.length === 0) return;
    for (const node of nodes) {
      if (node.children && node.children.length > 0) {
        if (node.key !== undefined && node.key !== null) {
          keys.push(node.key);
        } else if (node.id !== undefined || node.parentId !== undefined) {
          keys.push(`${node.parentId ?? ""}-${node.id ?? ""}`);
        }
        traverse(node.children);
      }
    }
  };

  traverse(treeData.value);
  expandedKeys.value = keys;
  autoExpandParent.value = true;
}

/**
 * 折叠所有节点
 */
function handleMenuCollapseAll() {
  expandedKeys.value = [];
  autoExpandParent.value = false;
}

/**
 * 取消选中所有节点
 */
function handleMenuUnselectedAll() {
  clearSelection();
}

/**
 * 展开单个节点
 */
const handleExpandNode = (keys: string[]) => {
  expandedKeys.value = keys;
  autoExpandParent.value = false;
};

/**
 * 选中组织单元
 */
function handleSelectOrgUnit(node: any) {
  userViewStore.setCurrentOrgUnitId(node ? node.id || null : null);
}

/**
 * 选中单个节点
 */
function handleSelectNode(node: TreeNode) {
  selectedKeys.value = [node.key];
  handleSelectOrgUnit(node);
}

function handleTenantChanged(value: any) {
  userViewStore.setCurrentTenantId(value);
  fetchOrgUnits();
}

function clearSelection() {
  selectedKeys.value = [];
  handleSelectOrgUnit(null);
}

function handleToolbarClick(action: any) {
  action.handler();
}

watch(searchValue, (val) => {
  const q = String(val ?? "").trim();
  if (!q) {
    expandedKeys.value = [];
    autoExpandParent.value = false;
    return;
  }

  const parentKeys = new Set<number | string>();
  const collect = (nodes: any[] | undefined, parents: (number | string)[] = []) => {
    if (!nodes || nodes.length === 0) return;
    for (const node of nodes) {
      const title = String(node.label ?? "");
      const key = node.key;
      if (title.toLowerCase().includes(q.toLowerCase())) {
        parents.forEach((p) => {
          if (p !== undefined && p !== null) parentKeys.add(p);
        });
      }
      collect(node.children, [...parents, key as number | string]);
    }
  };

  collect(treeData.value);
  expandedKeys.value = [...parentKeys];
  autoExpandParent.value = true;
});

onMounted(async () => {
  if (!userViewStore.isTenantUser()) {
    await userViewStore.fetchTenantList({ status: "ON" });
  }

  await fetch();
});
</script>

<style lang="scss" scoped>
.dept-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  padding: 8px;

  .card-flat {
    flex: 0 0 auto;
    margin-bottom: 8px;

    :deep(.el-card__body) {
      padding: 8px;
    }
  }

  .toolbar-container {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .input-row {
    display: flex;
    align-items: center;
    gap: 8px;

    .input-label {
      flex-shrink: 0;
      font-weight: 500;
      white-space: nowrap;
    }

    .search-input {
      flex: 1;
    }
  }

  .tree-container {
    flex: 1;
    overflow-y: auto;
    min-height: 0;
  }

  .tree-node-label {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .highlight-text {
    color: var(--el-color-danger);
    font-weight: 500;
  }
}
</style>
