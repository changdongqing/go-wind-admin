<template>
  <template v-if="col.formatter">
    {{ col.formatter(row, col) }}
  </template>
  <component
    :is="renderer"
    v-else-if="renderer"
    :col="col"
    :row="row"
    :field="field"
    :row-index="rowIndex"
    @modify="(d: any) => emit('modify', d)"
    @operate="(d: any) => emit('operate', d)"
  />
  <span v-else>{{ field ? row[field] : "" }}</span>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { Component } from "vue";
import { getCellRenderer } from "./cellRendererRegistry";
import type { ProTableColumn } from "./types";

const props = defineProps<{
  col: ProTableColumn;
  row: any;
  rowIndex: number;
}>();

const emit = defineEmits<{
  modify: [data: { row: any; field: string; value: any }];
  operate: [data: { name: string; row: any; $index: number }];
}>();

const field = computed(() => props.col.prop ?? "");
const renderer = computed<Component | undefined>(() => {
  const type = props.col.cellType;
  if (!type || type === "custom") return undefined;
  return getCellRenderer(type);
});
</script>
