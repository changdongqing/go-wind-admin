<template>
  <div
    v-if="showPagination"
    class="pro-pagination"
    :class="{ 'pro-pagination--hidden': hideOnSinglePage && total <= pageSize }"
  >
    <ElPagination
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :page-sizes="pageSizes"
      :total="total"
      :background="background"
      :disabled="disabled"
      :layout="computedLayout"
      :pager-count="pagerCount"
      :small="small"
      v-bind="$attrs"
      @current-change="handleCurrentChange"
      @size-change="handleSizeChange"
      @prev-click="handlePrevClick"
      @next-click="handleNextClick"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { ElPagination } from "element-plus";
import type { ProPaginationProps, PaginationEmits } from "./types";

defineOptions({ inheritAttrs: false });

const props = withDefaults(defineProps<ProPaginationProps>(), {
  currentPage: 1,
  pageSize: 20,
  pageSizes: () => [10, 20, 30, 50, 100],
  total: 0,
  background: true,
  disabled: false,
  showTotal: true,
  showSizes: true,
  showJump: true,
  pagerCount: 7,
  small: false,
  hideOnSinglePage: false,
});

const emit = defineEmits<PaginationEmits>();

const currentPage = ref(props.currentPage);
const pageSize = ref(props.pageSize);

// 自动计算 layout
const computedLayout = computed(() => {
  if (props.layout) return props.layout;

  const parts = ["total"];
  if (props.showSizes) parts.push("sizes");
  parts.push("prev", "pager", "next");
  if (props.showJump) parts.push("jump");

  return parts.join(" -> ");
});

const showPagination = computed(() => {
  if (props.hideOnSinglePage) {
    return props.total > props.pageSize;
  }
  return props.total > 0;
});

function handleCurrentChange(val: number) {
  currentPage.value = val;
  emit("update:modelValue", { currentPage: val, pageSize: pageSize.value });
  emit("current-change", val);
}

function handleSizeChange(val: number) {
  pageSize.value = val;
  currentPage.value = 1;
  emit("update:modelValue", { currentPage: 1, pageSize: val });
  emit("size-change", val);
}

function handlePrevClick(val: number) {
  emit("prev-click", val);
}

function handleNextClick(val: number) {
  emit("next-click", val);
}

// 监听外部变化
watch(
  () => props.currentPage,
  (val) => {
    currentPage.value = val;
  }
);
watch(
  () => props.pageSize,
  (val) => {
    pageSize.value = val;
  }
);

// 暴露方法
defineExpose({
  currentPage,
  pageSize,
  reset: () => {
    currentPage.value = 1;
    pageSize.value = props.pageSize;
  },
  setCurrentPage: (page: number) => {
    currentPage.value = page;
  },
  setPageSize: (size: number) => {
    pageSize.value = size;
  },
});
</script>

<style scoped lang="scss">
.pro-pagination {
  padding: 12px 0;
  display: flex;
  justify-content: flex-end;

  &--hidden {
    display: none;
  }

  :deep(.el-pagination) {
    display: flex;
    align-items: center;
    gap: 8px;

    .el-pagination__total {
      margin-right: auto;
    }
  }
}
</style>
