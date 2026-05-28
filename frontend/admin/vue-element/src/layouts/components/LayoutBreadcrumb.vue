<template>
  <el-breadcrumb v-show="visible" class="breadcrumb" :class="breadcrumbClass">
    <!-- 首页图标 -->
    <el-breadcrumb-item
      v-if="breadcrumbPrefs.showHome"
      :to="{ path: '/' }"
      class="breadcrumb__home"
    >
      <div v-if="breadcrumbPrefs.showIcon" class="i-svg:homepage breadcrumb__icon" />
      {{ $t("common.breadcrumb.home") }}
    </el-breadcrumb-item>
    <el-breadcrumb-item v-for="(item, index) in breadcrumbs" :key="item.path">
      <span
        v-if="item.redirect === 'noredirect' || index === breadcrumbs.length - 1"
        class="color-gray-400"
      >
        <div
          v-if="breadcrumbPrefs.showIcon && item.meta?.icon"
          :class="getIconClass(item.meta.icon as string)"
          class="breadcrumb__item-icon"
        />
        {{ translateRouteTitle((item.meta.title as string) ?? "") }}
      </span>
      <a v-else @click.prevent="handleLink(item)">
        <div
          v-if="breadcrumbPrefs.showIcon && item.meta?.icon"
          :class="getIconClass(item.meta.icon as string)"
          class="breadcrumb__item-icon"
        />
        {{ translateRouteTitle((item.meta.title as string) ?? "") }}
      </a>
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>

<script setup lang="ts">
import { RouteLocationMatched } from "vue-router";
import { compile } from "path-to-regexp";

import { router } from "@/router";
import { translateRouteTitle } from "@/core/i18n";
import { preferences } from "@/core/preferences";

const currentRoute = useRoute();
const pathCompile = (path: string) => {
  const { params } = currentRoute;
  const toPath = compile(path);
  return toPath(params);
};

// 面包屑偏好
const breadcrumbPrefs = computed(() => preferences.breadcrumb);

const breadcrumbs = ref<Array<RouteLocationMatched>>([]);

// 是否可见：启用 + 不只有一个时隐藏检查
const visible = computed(() => {
  if (!breadcrumbPrefs.value.enable) return false;
  return !(breadcrumbPrefs.value.hideOnlyOne && breadcrumbs.value.length <= 1);
});

// 面包屑样式类
const breadcrumbClass = computed(() => {
  return {
    "breadcrumb--background": breadcrumbPrefs.value.styleType === "background",
  };
});

// 图标类名处理（通用：支持任意 UnoCSS 图标集）
// `prefix:name` → `i-prefix:name`（lucide、fa、mdi 等）
// 无前缀 → `i-svg:name`（本地 SVG）
function getIconClass(icon?: string) {
  if (!icon) return "";
  if (icon.includes(":")) return `i-${icon}`;
  return `i-svg:${icon}`;
}

function getBreadcrumb() {
  breadcrumbs.value = currentRoute.matched.filter(
    (item) =>
      item.meta && item.meta.title && item.meta.breadcrumb !== false && !item.meta.hideInBreadcrumb
  );
}

function handleLink(item: any) {
  const { redirect, path } = item;
  if (redirect) {
    router.push(redirect).then(
      () => {},
      (err) => {
        console.warn(err);
      }
    );
    return;
  }
  router.push(pathCompile(path)).then(
    () => {},
    (err) => {
      console.warn(err);
    }
  );
}

watch(
  () => currentRoute.path,
  () => {
    getBreadcrumb();
  }
);

onBeforeMount(() => {
  getBreadcrumb();
});
</script>

<style lang="scss" scoped>
.breadcrumb {
  display: flex;
  align-items: center;

  // 覆盖 element-plus 的样式
  :deep(.el-breadcrumb__inner),
  :deep(.el-breadcrumb__inner a) {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-weight: 400 !important;
    color: var(--el-text-color-regular) !important;
  }

  // background 风格
  &--background {
    :deep(.el-breadcrumb__item) {
      .el-breadcrumb__inner {
        padding: 2px 8px;
        border-radius: 4px;
        transition: background-color 0.2s;
      }

      &:not(:last-child) .el-breadcrumb__inner {
        background-color: var(--el-fill-color-light);
      }

      &:not(:last-child) .el-breadcrumb__inner:hover {
        background-color: var(--el-fill-color);
      }
    }
  }

  &__home {
    :deep(.el-breadcrumb__inner) {
      display: inline-flex;
      align-items: center;
    }
  }

  &__icon,
  &__item-icon {
    width: 16px;
    height: 16px;
    flex-shrink: 0;
    color: currentColor;
  }
}
</style>
