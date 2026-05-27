<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit">
      <!-- 类型 -->
      <template #type="scope: any">
        <ElTag size="small" effect="dark" round :color="orgUnitTypeToColor(scope.row.type)">
          {{ orgUnitTypeToName(scope.row.type) }}
        </ElTag>
      </template>

      <!-- 状态 -->
      <template #status="scope: any">
        <ElTag size="small" effect="dark" round :color="orgUnitStatusToColor(scope.row.status)">
          {{ orgUnitStatusToName(scope.row.status) }}
        </ElTag>
      </template>
    </ProPage>

    <!-- 新增/编辑抽屉 -->
    <OrgDrawer ref="drawerRef" @success="handleSuccess" />
  </div>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import { ElTag } from "element-plus";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import OrgDrawer from "./org-drawer.vue";

import {
  orgUnitStatusToColor,
  orgUnitStatusToName,
  orgUnitTypeListForQuery,
  orgUnitTypeToColor,
  orgUnitTypeToName,
  fetchListOrgUnits,
  useDeleteOrgUnit,
} from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/i18n";

const { mutateAsync: deleteOrgUnit } = useDeleteOrgUnit();

const pageRef = ref();
const drawerRef = ref();

const pageConfig: ProPageConfig = {
  permPrefix: "sys:platform_admin",

  search: {
    grid: true,
    fields: [
      {
        type: "input",
        label: $t("pages.org_unit.name"),
        field: "name",
        attrs: { placeholder: $t("common.placeholder.input"), clearable: true },
      },
      {
        type: "select",
        label: $t("common.table.status"),
        field: "status",
        attrs: { placeholder: $t("common.placeholder.select"), clearable: true },
        options: [
          { value: "ON", label: $t("enum.status.ON") },
          { value: "OFF", label: $t("enum.status.OFF") },
        ],
      },
      {
        type: "select",
        label: $t("pages.org_unit.type"),
        field: "type",
        attrs: { placeholder: $t("common.placeholder.select"), clearable: true },
        options: orgUnitTypeListForQuery.value,
      },
    ],
  },

  table: {
    listAction: async (query: any) => {
      const { page, pageSize, ...queryParams } = query;
      const result = await fetchListOrgUnits(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 100 },
          formValues: queryParams,
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    deleteAction: async (ids: string) => {
      await deleteOrgUnit({ id: ids as any });
    },
    toolbar: [],
    toolbarRight: ["add"],
    defaultToolbar: ["refresh", "filter"],
    tableAttrs: {
      border: true,
      stripe: false,
      "tree-config": {
        expandAll: true,
        rowField: "id",
        childrenField: "children",
      },
    },
    columns: [
      { type: "index", label: $t("common.table.seq"), width: 60 },
      { prop: "name", label: $t("pages.org_unit.name"), minWidth: 150, treeNode: true },
      { prop: "code", label: $t("pages.org_unit.code"), minWidth: 120 },
      {
        prop: "type",
        label: $t("pages.org_unit.type"),
        minWidth: 100,
        slotName: "type",
      },
      { prop: "leaderName", label: $t("pages.org_unit.leaderName"), minWidth: 100 },
      {
        prop: "status",
        label: $t("common.table.status"),
        minWidth: 100,
        slotName: "status",
      },
      { prop: "sortOrder", label: $t("common.table.sortOrder"), width: 80 },
      {
        prop: "createdAt",
        label: $t("common.table.createdAt"),
        minWidth: 160,
        cellType: "date",
        dateFormat: "YYYY-MM-DD HH:mm:ss",
      },
      { prop: "remark", label: $t("common.table.remark"), minWidth: 150 },
      {
        prop: "action",
        label: $t("common.table.action"),
        fixed: "right",
        width: 150,
        cellType: "tool",
        buttons: [
          { name: "edit", text: $t("common.button.edit") },
          { name: "delete", text: $t("common.button.delete"), attrs: { type: "danger" } },
        ],
      },
    ],
  },
};

function handleAdd() {
  drawerRef.value?.open({ create: true });
}

function handleEdit(row: any) {
  drawerRef.value?.open({ create: false, row });
}

function handleSuccess() {
  pageRef.value?.refresh();
}
</script>

<style lang="scss" scoped>
.app-container {
  padding: 20px;
  width: 100%;
  min-width: 0;
  flex-shrink: 0;
}
</style>
