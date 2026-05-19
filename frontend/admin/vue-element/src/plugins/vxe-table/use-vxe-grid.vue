<template>
  <div :class="['h-full rounded-md', className]">
    <VxeGrid
      ref="gridRef"
      :class="['p-2', { 'pt-0': showToolbar && !formOptions }, gridClass]"
      v-bind="options"
      v-on="events"
    >
      <!-- 左侧操作区域或者title -->
      <template v-if="showToolbar" #toolbar-actions="slotProps">
        <slot v-if="showTableTitle" name="table-title">
          <div class="mr-1 pl-1 text-base font-medium">
            {{ tableTitle }}
            <HelpTooltip v-if="tableTitleHelp" trigger-class="pb-1">
              {{ tableTitleHelp }}
            </HelpTooltip>
          </div>
        </slot>
        <slot name="toolbar-actions" v-bind="slotProps"></slot>
      </template>

      <!-- 继承默认的slot -->
      <template v-for="slotName in delegatedSlots" :key="slotName" #[slotName]="slotProps">
        <slot :name="slotName" v-bind="slotProps"></slot>
      </template>

      <!-- loading -->
      <template #loading>
        <slot name="loading">
          <div v-loading="true" class="h-full"></div>
        </slot>
      </template>

      <!-- 空状态 -->
      <template #empty>
        <slot name="empty">
          <ElEmpty :description="$t('common.noData')" />
        </slot>
      </template>
    </VxeGrid>
  </div>
</template>

<script lang="ts" setup>
import { computed, nextTick, onMounted, onUnmounted, toRaw, useSlots, useTemplateRef } from "vue";

import type {
  VxeGridDefines,
  VxeGridInstance,
  VxeGridListeners,
  VxeGridPropTypes,
  VxeGridProps as VxeTableGridProps,
} from "vxe-table";

import type { ExtendedVxeGridApi, VxeGridProps } from "./types";

import { usePriorityValues } from "@/composables";
import { $t } from "@/i18n";
import { usePreferences } from "@/core/preferences";
import { cloneDeep, mergeWithArrayOverride } from "@/utils";

import { ElEmpty } from "element-plus";
import { VxeGrid, VxeUI } from "vxe-table";

import "vxe-table/styles/cssvar.scss";
import "vxe-pc-ui/styles/cssvar.scss";
import "./style.css";
import { HelpTooltip } from "@/components/Tooltip";

interface Props extends VxeGridProps {
  api: ExtendedVxeGridApi;
}

const props = withDefaults(defineProps<Props>(), {});

const TOOLBAR_ACTIONS = "toolbar-actions";
const TOOLBAR_TOOLS = "toolbar-tools";

const gridRef = useTemplateRef<VxeGridInstance>("gridRef");

const state = props.api?.useStore?.();

const {
  gridOptions,
  class: className,
  gridClass,
  gridEvents,
  formOptions,
  tableTitle,
  tableTitleHelp,
} = usePriorityValues(props, state);

const { isMobile } = usePreferences();

const slots = useSlots();

const showTableTitle = computed(() => {
  return !!slots["table-title"]?.() || tableTitle.value;
});

const showToolbar = computed(() => {
  return !!slots[TOOLBAR_ACTIONS]?.() || !!slots[TOOLBAR_TOOLS]?.() || showTableTitle.value;
});

const toolbarOptions = computed(() => {
  const slotActions = slots[TOOLBAR_ACTIONS]?.();
  const slotTools = slots[TOOLBAR_TOOLS]?.();

  const toolbarConfig: VxeGridPropTypes.ToolbarConfig = {
    tools: [],
  };

  if (!showToolbar.value) {
    return { toolbarConfig };
  }

  // 强制使用固定的toolbar配置，不允许用户自定义
  // 减少配置的复杂度，以及后续维护的成本
  toolbarConfig.slots = {
    ...(slotActions || showTableTitle.value ? { buttons: TOOLBAR_ACTIONS } : {}),
    ...(slotTools ? { tools: TOOLBAR_TOOLS } : {}),
  };

  return { toolbarConfig };
});

const options = computed(() => {
  const globalGridConfig = VxeUI?.getConfig()?.grid ?? {};

  const mergedOptions: VxeTableGridProps = cloneDeep(
    mergeWithArrayOverride({}, toolbarOptions.value, toRaw(gridOptions.value), globalGridConfig)
  );

  if (mergedOptions.proxyConfig) {
    const { ajax } = mergedOptions.proxyConfig;
    mergedOptions.proxyConfig.enabled = !!ajax;
    // 不自动加载数据, 由组件控制
    mergedOptions.proxyConfig.autoLoad = false;
  }

  if (mergedOptions.pagerConfig) {
    const mobileLayouts = ["PrevJump", "PrevPage", "Number", "NextPage", "NextJump"] as string[];
    const layouts = ["Total", "Sizes", "Home", ...mobileLayouts, "End"] as string[];
    mergedOptions.pagerConfig = mergeWithArrayOverride({}, mergedOptions.pagerConfig, {
      pageSize: 20,
      background: true,
      pageSizes: [10, 20, 30, 50, 100, 200],
      className: "mt-2 w-full",
      layouts: isMobile.value ? mobileLayouts : layouts,
      size: "mini" as const,
    }) as VxeTableGridProps["pagerConfig"];
  }

  return mergedOptions;
});

function onToolbarToolClick(event: VxeGridDefines.ToolbarToolClickEventParams) {
  if (event.code === "search") {
    props.api?.toggleSearchForm?.();
  }
  (gridEvents.value?.toolbarToolClick as VxeGridListeners["toolbarToolClick"])?.(event);
}

const events = computed(() => {
  return {
    ...gridEvents.value,
    toolbarToolClick: onToolbarToolClick,
  };
});

const delegatedSlots = computed(() => {
  const resultSlots: string[] = [];

  for (const key of Object.keys(slots)) {
    if (!["empty", "loading", TOOLBAR_ACTIONS].includes(key)) {
      resultSlots.push(key);
    }
  }
  return resultSlots;
});

async function init() {
  await nextTick();
  const globalGridConfig = VxeUI?.getConfig()?.grid ?? {};
  const defaultGridOptions: VxeTableGridProps = mergeWithArrayOverride(
    {},
    toRaw(gridOptions.value),
    toRaw(globalGridConfig)
  );

  // 内部主动加载数据，防止form的默认值影响
  const autoLoad = defaultGridOptions.proxyConfig?.autoLoad;
  const enableProxyConfig = options.value.proxyConfig?.enabled;
  if (enableProxyConfig && autoLoad) {
    props.api.reload({});
  }

  props.api?.setState?.({ gridOptions: defaultGridOptions as any });
}

onMounted(() => {
  props.api?.mount?.(gridRef.value, null as any);
  init();
});

onUnmounted(() => {
  props.api?.unmount?.();
});
</script>
