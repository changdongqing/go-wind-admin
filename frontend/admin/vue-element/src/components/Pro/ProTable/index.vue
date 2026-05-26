<template>
  <div class="pro-table">
    <el-table :data="data" v-loading="loading" border style="width: 100%">
      <!-- 自动渲染表格列 -->
      <el-table-column
        v-for="column in columns"
        :key="column.prop"
        :prop="column.prop"
        :label="column.label"
        :width="column.width"
        :min-width="column.minWidth"
        :fixed="column.fixed"
        :sortable="column.sortable"
        :show-overflow-tooltip="column.showOverflowTooltip ?? true"
        v-bind="column.attrs"
      >
        <template #default="scope">
          <!-- 自定义组件优先 -->
          <component
            v-if="column.component"
            :is="column.component"
            v-bind="column.componentProps"
            v-model="scope.row[column.prop]"
          />
          <!-- 自动渲染内置类型 -->
          <component
            v-else
            :is="getTableComponent(column.valueType)"
            v-bind="getColumnProps(column.valueType, scope.row[column.prop])"
          >
            {{ scope.row[column.prop] }}
          </component>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts" name="ProTable">
import type { ProTableProps } from "./types";

// 接收 props（完全使用你的类型）
const props = defineProps<ProTableProps>();

// 映射表格渲染组件
const getTableComponent = (type: any = "text") => {
  const componentMap: Record<string, string> = {
    text: "span",
    index: "span",
    selection: "span",
    expand: "span",
    tag: "el-tag",
    date: "span",
    datetime: "span",
    switch: "el-switch",
    link: "el-link",
    avatar: "el-avatar",
  };
  return componentMap[type] || "span";
};

// 为不同列类型传递原生属性
const getColumnProps = (type: any, value: any) => {
  const propsMap: Record<string, any> = {
    switch: { modelValue: value, disabled: true },
    tag: { type: value ? "success" : "danger" },
    avatar: { src: value, size: 40 },
  };
  return propsMap[type] || {};
};
</script>

<style scoped>
.pro-table {
  margin-top: 16px;
}
</style>
