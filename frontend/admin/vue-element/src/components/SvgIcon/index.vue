<!--
  统一图标组件

  将项目中所有图标渲染统一到 Iconify 体系：
  - `lucide:users`       → @iconify/vue 渲染 Lucide 图标
  - `ep:setting`         → @iconify/vue 渲染 Element Plus 图标
  - `fa:user`            → @iconify/vue 渲染 Font Awesome 图标
  - `svg:menu`           → UnoCSS presetIcons 渲染本地 SVG
  - `el-icon-Setting`    → 兼容旧格式，自动转为 `ep:setting`
  - `menu`               → 无前缀兜底，当作本地 SVG 处理

  扩展新图标集只需安装对应的 @iconify-json/xxx 包即可。
-->
<template>
  <!-- 本地 SVG 图标：用 UnoCSS CSS 类渲染 -->
  <div
    v-if="isSvgIcon"
    :class="[cssClass, props.class]"
    :style="{ width: `${size || 16}px`, height: `${size || 16}px` }"
  />
  <!-- Iconify 图标：用 @iconify/vue 组件渲染 -->
  <Icon v-else :icon="resolvedIcon" :width="size || 16" :height="size || 16" :class="props.class" />
</template>

<script setup lang="ts">
import { Icon } from "@iconify/vue";
import { computed } from "vue";

const props = defineProps<{
  /** 图标名称，支持 Iconify 格式（lucide:users）和旧格式（el-icon-Setting） */
  icon?: string;
  /** 图标尺寸 */
  size?: number | string;
  /** 自定义类名 */
  class?: string;
}>();

/**
 * 将旧格式图标名转为 Iconify 标准格式
 */
const resolvedIcon = computed(() => {
  if (!props.icon) return "";

  // 旧格式兼容：el-icon-Setting → ep:setting
  if (props.icon.startsWith("el-icon-")) {
    const name = props.icon.replace("el-icon-", "");
    return `ep:${name.charAt(0).toLowerCase() + name.slice(1)}`;
  }

  // 已经是 Iconify 格式（lucide:users, fa:user, ep:setting 等）
  if (props.icon.includes(":")) {
    return props.icon;
  }

  // 无前缀：本地 SVG 图标，通过 UnoCSS 渲染
  return `svg:${props.icon}`;
});

/**
 * 判断是否使用 UnoCSS CSS 类方式渲染（本地 SVG）
 * 本地 SVG 图标通过 UnoCSS presetIcons 的 FileSystemIconLoader 加载，
 * 不走 @iconify/vue 组件
 */
const isSvgIcon = computed(() => {
  return resolvedIcon.value.startsWith("svg:");
});

/**
 * UnoCSS 图标类名
 */
const cssClass = computed(() => {
  if (!isSvgIcon.value) return "";
  // svg:menu → i-svg:menu
  return `i-${resolvedIcon.value}`;
});
</script>
