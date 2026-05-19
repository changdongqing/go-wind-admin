<template>
  <ElTooltip
    :delay="delayDuration"
    :effect="effect"
    :placement="placement"
    :popper-class="contentClass"
    :popper-style="contentStyle"
    :show-after="0"
  >
    <slot name="trigger"></slot>
    <template #content>
      <slot></slot>
    </template>
  </ElTooltip>
</template>

<script setup lang="ts">
import type { StyleValue } from "vue";

import { ElTooltip } from "element-plus";

interface Props {
  contentClass?: string;
  contentStyle?: StyleValue;
  delayDuration?: number;
  effect?: "dark" | "light";
  placement?:
    | "top"
    | "top-start"
    | "top-end"
    | "bottom"
    | "bottom-start"
    | "bottom-end"
    | "left"
    | "left-start"
    | "left-end"
    | "right"
    | "right-start"
    | "right-end";
}

withDefaults(defineProps<Props>(), {
  delayDuration: 0,
  effect: "dark",
  placement: "right",
});
</script>
