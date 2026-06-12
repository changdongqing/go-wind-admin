<script setup lang="ts">
import {
  computed,
  getCurrentInstance,
  nextTick,
  onMounted,
  onUnmounted,
  ref,
  watch,
} from 'vue';

import { preferences } from '@vben/preferences';

import VueJsonEditor from 'json-editor-vue';

// vanilla-jsoneditor 暗色主题通过内联 <style> 引入（间接依赖无法直接 import）

// 类型定义
interface Props {
  modelValue: string;
  height?: number | string;
  disabled?: boolean;
  placeholder?: string;
  options?: {
    mode?: any;
    modes?: any[];
    search?: boolean;
  };
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  height: 500,
  placeholder: '{}',
  options: () => ({
    mode: 'text',
    modes: ['tree', 'code', 'form', 'text', 'view'],
    search: true,
  }),
});

const emit = defineEmits<{
  (e: 'change', value: string): void;
  (e: 'error', error: Error): void;
  (e: 'ready'): void;
  (e: 'update:modelValue', value: string): void;
}>();

// 响应式数据
const localValue = ref(props.modelValue);
const jsonData = ref<any[] | null | Record<string, any>>(null);
const parseError = ref<string>('');
const isValidJson = ref(true);
const instance = getCurrentInstance();
let observer: MutationObserver | null = null;
let themeObserver: MutationObserver | null = null;

const isDark = ref(false);

const updateIsDark = () => {
  const prefersDark = preferences.theme.mode === 'dark';
  if (typeof document === 'undefined') {
    isDark.value = prefersDark;
    return;
  }
  const root = document.documentElement;
  isDark.value =
    prefersDark ||
    root.classList.contains('dark') ||
    root.classList.contains('theme-dark') ||
    root.classList.contains('json-editor-dark');
};

// computed
// 验证并格式化 JSON
const validateAndFormat = (value: string) => {
  try {
    if (!value?.trim()) {
      parseError.value = '';
      isValidJson.value = true;
      return { parsed: null, formatted: '' };
    }
    const parsed = JSON.parse(String(value));
    const formatted = JSON.stringify(parsed, null, 2);
    parseError.value = '';
    isValidJson.value = true;
    return { parsed, formatted };
  } catch (error) {
    const err = error as Error;
    parseError.value = `JSON解析错误: ${err.message || '未知错误'}`;
    isValidJson.value = false;
    emit('error', err);
    return { parsed: null, formatted: value };
  }
};

// 初始化数据
const initData = () => {
  const { parsed, formatted } = validateAndFormat(props.modelValue);
  localValue.value = formatted || props.placeholder;

  // 🛡️ 确保 jsonData 是对象类型
  if (parsed !== null && typeof parsed === 'object') {
    jsonData.value = parsed;
  } else if (parsed === null) {
    jsonData.value = {};
  } else {
    // 兜底：非对象值包装处理
    jsonData.value = { value: parsed };
  }
};

// 监听外部值变化
watch(
  () => props.modelValue,
  (newVal) => {
    if (newVal !== localValue.value) {
      const { parsed, formatted } = validateAndFormat(newVal);
      localValue.value = formatted || newVal || props.placeholder;
      console.log('props.modelValue');
      try {
        jsonData.value = parsed || JSON.parse(props.placeholder);
      } catch {
        jsonData.value = {};
      }
    }
  },
  { immediate: true, deep: false },
);

// 监听编辑器内部数据变化
watch(
  () => jsonData.value,
  (newVal) => {
    if (newVal === null) return;

    if (typeof newVal === 'string') {
      if (newVal !== localValue.value) {
        localValue.value = newVal;
        emit('update:modelValue', newVal);
        emit('change', newVal);
      }
      return;
    }

    // 正常对象/数组：序列化为字符串
    try {
      const newValue = JSON.stringify(newVal, null, 2);
      if (newValue !== localValue.value) {
        localValue.value = newValue;
        emit('update:modelValue', newValue);
        emit('change', newValue);
      }
      parseError.value = '';
      isValidJson.value = true;
    } catch (error) {
      const err = error as Error;
      parseError.value = `JSON序列化错误: ${err.message || '未知错误'}`;
      isValidJson.value = false;
      emit('error', err);
    }
  },
  { deep: true },
);

// 高度计算（优化类型安全）
const editorHeight = computed(() => {
  let baseHeight = 500;

  if (typeof props.height === 'number') {
    baseHeight = props.height;
  } else if (typeof props.height === 'string') {
    const numericHeight = Number(props.height);
    if (!Number.isNaN(numericHeight)) {
      baseHeight = numericHeight;
    } else if (props.height.endsWith('px')) {
      const pxValue = Number(props.height.replace('px', ''));
      if (!Number.isNaN(pxValue)) {
        baseHeight = pxValue;
      }
    } else {
      // 百分比等非数值高度直接返回原字符串
      return props.height;
    }
  }

  const finalHeight = Math.max(baseHeight - 40, 200);
  return `${finalHeight}px`;
});

// 刷新编辑器样式
const refreshEditor = () => {
  nextTick(() => {
    const container = instance?.proxy?.$el as HTMLElement | undefined;
    if (!container) return;
    container.dataset.theme = isDark.value ? 'dark' : 'light';
  });
};

// 监听主题变化
watch(
  () => preferences.theme.mode,
  () => {
    updateIsDark();
    refreshEditor();
  },
  { immediate: true },
);

// 监听编辑器模式变化
watch(
  () => props.options?.mode,
  () => {
    refreshEditor();
  },
);

// 编辑器change事件处理
const handleEditorChange = (value: any) => {
  if (typeof value === 'string') {
    const rawValue = value;
    localValue.value = rawValue;
    emit('update:modelValue', rawValue);
    emit('change', rawValue);

    const { parsed } = validateAndFormat(rawValue);
    if (parsed !== null && typeof parsed === 'object') {
      jsonData.value = parsed;
    }
    return;
  }

  if (Array.isArray(value) || (value !== null && typeof value === 'object')) {
    return;
  }

  jsonData.value = { value };
  refreshEditor();
};

// 初始化和销毁逻辑
onMounted(() => {
  updateIsDark();
  initData();
  nextTick(() => {
    emit('ready');
    refreshEditor();

    if (typeof document !== 'undefined') {
      const root = document.documentElement;
      themeObserver = new MutationObserver(() => {
        updateIsDark();
        refreshEditor();
      });
      themeObserver.observe(root, {
        attributes: true,
        attributeFilter: ['class'],
      });
    }

    const container = instance?.proxy?.$el as HTMLElement | undefined;
    if (!container) return;

    const editorEl = container.querySelector('.json-editor-core');

    if (editorEl) {
      observer = new MutationObserver((mutations) => {
        const hasStyleChange = mutations.some(
          (m) =>
            m.type === 'attributes' &&
            ['class', 'style'].includes(m.attributeName || ''),
        );
        if (isDark.value && hasStyleChange) {
          refreshEditor();
        }
      });

      observer.observe(editorEl, {
        childList: true,
        subtree: true,
        attributes: true,
        attributeFilter: ['class', 'style'],
      });
    }
  });
});

onUnmounted(() => {
  if (themeObserver) {
    themeObserver.disconnect();
    themeObserver = null;
  }
  if (observer) {
    observer.disconnect();
    observer = null;
  }
});
</script>

<template>
  <div class="json-editor-container" :class="{ 'jse-theme-dark': isDark }">
    <!-- 错误提示 -->
    <div v-if="parseError" class="error-message">
      {{ parseError }}
    </div>

    <VueJsonEditor
      v-model="jsonData"
      :mode="options?.mode"
      :disabled="disabled"
      :search="options?.search"
      :placeholder="placeholder"
      :style="{ height: editorHeight, width: '100%' }"
      class="json-editor-core"
      @change="handleEditorChange"
    />
  </div>
</template>

<style scoped>
/* ============ 容器 ============ */
.json-editor-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  overflow: hidden;
  border: 1px solid var(--jse-panel-border, #d7d7d7);
  border-radius: 8px;
  background-color: var(--jse-background-color, #fff);
  /* vanilla-jsoneditor 内置 --jse-* 变量体系 */
  --jse-theme-color: #3b82f6;
  --jse-theme-color-highlight: #60a5fa;
}

/* ============ 错误提示 ============ */
.error-message {
  padding: 8px 12px;
  margin: 0;
  font-size: 12px;
  line-height: 1.4;
  color: var(--jse-error-color, #ee5341);
  background-color: #fef2f2;
  border-bottom: 1px solid #fecaca;
}

/* ============ 编辑器核心 ============ */
.json-editor-container :deep(.json-editor-core) {
  flex: 1;
  width: 100%;
  overflow: hidden;
}
</style>

<!-- vanilla-jsoneditor 内置暗色主题（从 jse-theme-dark.css 内联，因 pnpm 间接依赖无法直接 import） -->
<style>
.jse-theme-dark {
  --jse-theme: dark !important;
  --jse-theme-color: #2a3548 !important;
  --jse-theme-color-highlight: #344054 !important;
  --jse-background-color: #1e1e1e;
  --jse-text-color: #d4d4d4;
  --jse-text-color-inverse: #4d4d4d;
  --jse-main-border: 1px solid #4f4f4f;
  --jse-menu-color: #fff;
  --jse-modal-background: #2f2f2f;
  --jse-modal-overlay-background: rgb(0 0 0 / 50%);
  --jse-modal-code-background: #2f2f2f;
  --jse-tooltip-color: var(--jse-text-color);
  --jse-tooltip-background: #4b4b4b;
  --jse-tooltip-border: 1px solid #737373;
  --jse-tooltip-action-button-color: inherit;
  --jse-tooltip-action-button-background: #737373;
  --jse-panel-background: #333;
  --jse-panel-background-border: 1px solid #464646;
  --jse-panel-color: var(--jse-text-color);
  --jse-panel-color-readonly: #737373;
  --jse-panel-border: 1px solid #3c3c3c;
  --jse-panel-button-color-highlight: #e5e5e5;
  --jse-panel-button-background-highlight: #464646;
  --jse-navigation-bar-background: #656565;
  --jse-navigation-bar-background-highlight: #7e7e7e;
  --jse-navigation-bar-dropdown-color: var(--jse-text-color);
  --jse-context-menu-background: #4b4b4b;
  --jse-context-menu-background-highlight: #595959;
  --jse-context-menu-separator-color: #595959;
  --jse-context-menu-color: var(--jse-text-color);
  --jse-context-menu-pointer-background: #737373;
  --jse-context-menu-pointer-background-highlight: #818181;
  --jse-context-menu-pointer-color: var(--jse-context-menu-color);
  --jse-key-color: #9cdcfe;
  --jse-value-color: var(--jse-text-color);
  --jse-value-color-number: #b5cea8;
  --jse-value-color-boolean: #569cd6;
  --jse-value-color-null: #569cd6;
  --jse-value-color-string: #ce9178;
  --jse-value-color-url: #ce9178;
  --jse-delimiter-color: #949494;
  --jse-edit-outline: 2px solid var(--jse-text-color);
  --jse-selection-background-color: #464646;
  --jse-selection-background-inactive-color: #333;
  --jse-hover-background-color: #343434;
  --jse-active-line-background-color: rgb(255 255 255 / 6%);
  --jse-search-match-background-color: #343434;
  --jse-collapsed-items-background-color: #333;
  --jse-collapsed-items-selected-background-color: #565656;
  --jse-collapsed-items-link-color: #b2b2b2;
  --jse-collapsed-items-link-color-highlight: #ec8477;
  --jse-search-match-color: #724c27;
  --jse-search-match-outline: 1px solid #966535;
  --jse-search-match-active-color: #9f6c39;
  --jse-search-match-active-outline: 1px solid #bb7f43;
  --jse-tag-background: #444;
  --jse-tag-color: #bdbdbd;
  --jse-table-header-background: #333;
  --jse-table-header-background-highlight: #424242;
  --jse-table-row-odd-background: rgb(255 255 255 / 10%);
  --jse-input-background: #3d3d3d;
  --jse-input-border: var(--jse-main-border);
  --jse-button-background: #808080;
  --jse-button-background-highlight: #7a7a7a;
  --jse-button-color: #e0e0e0;
  --jse-button-secondary-background: #494949;
  --jse-button-secondary-background-highlight: #5d5d5d;
  --jse-button-secondary-background-disabled: #9d9d9d;
  --jse-button-secondary-color: var(--jse-text-color);
  --jse-a-color: #55abff;
  --jse-a-color-highlight: #4387c9;
  --jse-svelte-select-background: #3d3d3d;
  --jse-svelte-select-border: 1px solid #4f4f4f;
  --list-background: #3d3d3d;
  --item-hover-bg: #505050;
  --multi-item-bg: #5b5b5b;
  --input-color: #d4d4d4;
  --multi-clear-bg: #8a8a8a;
  --multi-item-clear-icon-color: #d4d4d4;
  --multi-item-outline: 1px solid #696969;
  --list-shadow: 0 2px 8px 0 rgb(0 0 0 / 40%);
  --jse-color-picker-background: #656565;
  --jse-color-picker-border-box-shadow: #8c8c8c 0 0 0 1px;
}
</style>
