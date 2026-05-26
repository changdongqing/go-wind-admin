<template>
  <div class="pro-page">
    <!-- 搜索区 -->
    <ProSearch
      v-if="config.search.fields.length"
      :config="config.search"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 表格区 -->
    <el-card shadow="never" class="pro-table-card">
      <div class="toolbar flex justify-between mb-3">
        <div class="flex gap-2">
          <el-button
            v-if="config.table.toolbar?.includes('add')"
            type="success"
            icon="plus"
            @click="modal.open('add')"
          >
            新增
          </el-button>
          <el-button
            v-if="config.table.toolbar?.includes('delete')"
            type="danger"
            icon="delete"
            :disabled="!selection.length"
            @click="handleBatchDelete"
          >
            删除
          </el-button>
        </div>
        <div class="flex gap-2">
          <el-button icon="download" @click="handleExport">导出</el-button>
          <el-button icon="upload" @click="$refs.fileInput.click()">导入</el-button>
          <input
            type="file"
            ref="fileInput"
            class="hidden"
            accept=".xlsx,.xls"
            @change="handleImport"
          />
        </div>
      </div>

      <ProTable
        :columns="config.table.columns"
        :data="table.data.value"
        :loading="table.loading.value"
        :pagination="config.table.pagination"
        v-bind="config.table.table"
        @selection-change="table.handleSelectionChange"
        @page-change="
          (p) => {
            table.pagination.currentPage = p;
            table.fetch(searchParams);
          }
        "
        @size-change="
          (s) => {
            table.pagination.pageSize = s;
            table.fetch(searchParams, true);
          }
        "
      >
        <!-- 透传所有自定义列插槽 -->
        <template v-for="(_, name) in $slots" #[name]="slotProps">
          <slot :name="name" v-bind="slotProps" />
        </template>
      </ProTable>
    </el-card>

    <!-- 弹窗/抽屉 -->
    <ProModal
      v-model:visible="modal.visible.value"
      :mode="modal.mode.value"
      :config="config.modal"
      :form-data="modal.formData as any"
      @submit="table.fetch(searchParams, true)"
    >
      <template v-for="(_, name) in $slots" #[name]="slotProps">
        <slot :name="name" v-bind="slotProps" />
      </template>
    </ProModal>
  </div>
</template>

<script setup lang="ts" generic="T extends Record<string, any>, Q extends Record<string, any>">
import { ref, reactive } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import ProSearch from "./ProSearch.vue";
import ProTable from "./ProTable.vue";
import ProModal from "./ProModal.vue";
import { useTableState } from "./composables/useTableState";
import { useModalState } from "./composables/useModalState";
import type { ProPageConfig } from "./types";

const props = defineProps<{ config: ProPageConfig<T, Q> }>();
const emit = defineEmits<{ search: [Q]; reset: [Q] }>();

const searchParams = reactive<Q>({} as Q);
const table = useTableState(props.config.table);
const modal = useModalState(props.config.modal);

function handleSearch(params: Q) {
  Object.assign(searchParams, params);
  table.fetch(searchParams, true);
  emit("search", searchParams);
}
function handleReset(params: Q) {
  Object.keys(searchParams).forEach((k) => delete (searchParams as any)[k]);
  Object.assign(searchParams, params);
  table.fetch(searchParams, true);
  emit("reset", searchParams);
}
async function handleBatchDelete() {
  const ids = table.getSelectionIds().join(",");
  if (!ids) return ElMessage.warning("请选择数据");
  await ElMessageBox.confirm("确认删除？");
  await props.config.table.deleteAction?.(ids);
  table.fetch(searchParams, true);
}
async function handleExport() {
  if (!props.config.table.exportAction) return;
  const blob = await props.config.table.exportAction(searchParams);
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "export.xlsx";
  a.click();
}
async function handleImport(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0];
  if (!file || !props.config.table.importAction) return;
  await props.config.table.importAction(file);
  table.fetch(searchParams, true);
}
</script>

<style scoped>
.pro-page {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
}
.pro-table-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
</style>
