<template>
  <ElDialog
    v-model="state.visible"
    :title="t('pages.curd.export.title')"
    width="600px"
    align-center
    @close="handleClose"
  >
    <ElScrollbar max-height="60vh">
      <ElForm
        ref="formRef"
        :model="state.form"
        :rules="formRules"
        style="padding-right: var(--el-dialog-padding-primary)"
      >
        <ElFormItem :label="t('pages.curd.export.filename')" prop="filename">
          <ElInput v-model="state.form.filename" clearable />
        </ElFormItem>
        <ElFormItem :label="t('pages.curd.export.sheetname')" prop="sheetname">
          <ElInput v-model="state.form.sheetname" clearable />
        </ElFormItem>
        <ElFormItem :label="t('pages.curd.export.origin')" prop="origin">
          <ElSelect v-model="state.form.origin">
            <ElOption :label="t('pages.curd.export.originOptions.current')" value="current" />
            <ElOption
              :label="t('pages.curd.export.originOptions.selected')"
              value="selected"
              :disabled="selectionData.length === 0"
            />
            <ElOption
              :label="t('pages.curd.export.originOptions.remote')"
              value="remote"
              :disabled="!exportsAction"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="t('pages.curd.export.fields')" prop="fields">
          <ElCheckboxGroup v-model="state.form.fields">
            <ElCheckbox
              v-for="col in exportableColumns"
              :key="col.prop"
              :value="col.prop"
              :label="col.label"
            />
          </ElCheckboxGroup>
        </ElFormItem>
      </ElForm>
    </ElScrollbar>
    <template #footer>
      <ElButton type="primary" @click="handleSubmit">
        {{ t("common.button.confirm") }}
      </ElButton>
      <ElButton @click="handleClose">{{ t("common.button.cancel") }}</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
import { reactive, ref, computed, nextTick } from "vue";
import {
  ElMessage,
  ElDialog,
  ElForm,
  ElFormItem,
  ElInput,
  ElSelect,
  ElOption,
  ElScrollbar,
  ElButton,
  ElCheckbox,
  ElCheckboxGroup,
} from "element-plus";
import type { FormInstance, FormRules } from "element-plus";
import { useThrottleFn } from "@vueuse/core";
import ExcelJS from "exceljs";
import { useI18n } from "@/i18n";
import type { ProTableColumn } from "../ProTable/types";

const props = defineProps<{
  columns: ProTableColumn[];
  selectionData: Record<string, any>[];
  tableData: Record<string, any>[];
  exportsAction?: (queryParams: any) => Promise<any[]>;
  searchParams?: Record<string, any>;
  defaultFilename?: string;
}>();

const { t } = useI18n();

const formRef = ref<FormInstance>();

// 可导出的列（有 prop 和 label 的普通列）
const exportableColumns = computed(() =>
  props.columns.filter((col: any) => col.prop && col.label && !col.type)
);

const defaultFields = computed(() =>
  exportableColumns.value.map((col: any) => col.prop).filter(Boolean)
);

const state = reactive<{
  visible: boolean;
  form: {
    filename: string;
    sheetname: string;
    fields: string[];
    origin: "current" | "selected" | "remote";
  };
}>({
  visible: false,
  form: {
    filename: "",
    sheetname: "",
    fields: [],
    origin: "current",
  },
});

const formRules: FormRules = {
  fields: [{ required: true, message: t("pages.curd.message.selectFields") }],
  origin: [{ required: true, message: t("pages.curd.message.selectOrigin") }],
};

function open() {
  state.form.fields = [...defaultFields.value];
  state.visible = true;
}

function handleClose() {
  state.visible = false;
  nextTick(() => formRef.value?.clearValidate());
}

const handleSubmit = useThrottleFn(() => {
  formRef.value?.validate((valid: boolean) => {
    if (!valid) return;
    doExport();
    handleClose();
  });
}, 3000);

function doExport() {
  const filename = state.form.filename || props.defaultFilename || "export";
  const sheetname = state.form.sheetname || "sheet";
  const workbook = new ExcelJS.Workbook();
  const worksheet = workbook.addWorksheet(sheetname);
  const excelCols: Partial<ExcelJS.Column>[] = [];

  exportableColumns.value.forEach((col: any) => {
    if (col.prop && col.label && state.form.fields.includes(col.prop)) {
      excelCols.push({ header: col.label, key: col.prop });
    }
  });
  worksheet.columns = excelCols;

  if (state.form.origin === "remote") {
    if (!props.exportsAction) {
      ElMessage.error(t("pages.curd.message.noExportsAction"));
      return;
    }
    props.exportsAction({ ...props.searchParams }).then((data) => {
      worksheet.addRows(data);
      workbook.xlsx.writeBuffer().then((buffer) => saveXlsx(buffer, filename));
    });
  } else {
    const rows = state.form.origin === "selected" ? props.selectionData : (props.tableData ?? []);
    worksheet.addRows(rows);
    workbook.xlsx.writeBuffer().then((buffer) => saveXlsx(buffer, filename));
  }
}

function saveXlsx(fileData: any, fileName: string) {
  const fileType =
    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet;charset=utf-8";
  const blob = new Blob([fileData], { type: fileType });
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = fileName;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
}

defineExpose({ open });
</script>
