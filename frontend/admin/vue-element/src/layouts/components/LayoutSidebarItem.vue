<template>
  <div v-if="!item.meta || !item.meta.hideInMenu">
    <!--【叶子节点】无可见子节点 -->
    <template v-if="!hasVisibleChildren(item.children)">
      <AppLink
        :to="{
          path: resolvePath(item.path),
          query: item.meta?.params as LocationQueryRaw,
        }"
      >
        <el-menu-item
          :index="resolvePath(item.path)"
          :class="{ 'submenu-title-noDropdown': !isNest }"
        >
          <MenuIcon :icon="getMetaIcon(item.meta)" />
          <span v-if="item.meta?.title" class="menu-title">
            {{ translateRouteTitle(getMetaTitle(item.meta)) }}
          </span>
        </el-menu-item>
      </AppLink>
    </template>

    <!--【非叶子节点】有可见子节点 -->
    <el-sub-menu v-else :index="resolvePath(item.path)" :data-path="item.path" teleported>
      <template #title>
        <template v-if="item.meta">
          <MenuIcon :icon="getMetaIcon(item.meta)" />
          <span v-if="item.meta.title" class="menu-title">
            {{ translateRouteTitle(getMetaTitle(item.meta)) }}
          </span>
        </template>
      </template>

      <LayoutSidebarItem
        v-for="child in item.children"
        :key="child.path"
        :is-nest="true"
        :item="child"
        :base-path="resolvePath(item.path)"
      />
    </el-sub-menu>
  </div>
</template>

<script setup lang="ts">
import path from "path-browserify";
import { RouteRecordRaw, LocationQueryRaw } from "vue-router";
import { isExternal } from "@/utils";
import { translateRouteTitle } from "@/core/i18n";
import { ElIcon } from "element-plus";

defineOptions({
  name: "LayoutSidebarItem",
  inheritAttrs: false,
});

// 菜单图标组件（通用图标解析，支持任意 UnoCSS 图标集）
const MenuIcon = defineComponent({
  props: { icon: String },
  setup(props) {
    return () => {
      if (!props.icon) {
        return h("div", { class: "i-svg:menu" });
      }

      // Element Plus 图标：`el-icon-` 前缀
      if (props.icon.startsWith("el-icon-")) {
        const name = props.icon.replace("el-icon-", "");
        return h(ElIcon, null, () => h(resolveComponent(name)));
      }

      // UnoCSS 图标集：`prefix:name` 格式 → `i-prefix:name`
      // 支持 lucide、fa、mdi 等任意在 uno.config.ts 中注册的集合
      if (props.icon.includes(":")) {
        return h("div", { class: `i-${props.icon}` });
      }

      // 无前缀兜底：本地 SVG 图标
      return h("div", { class: `i-svg:${props.icon}` });
    };
  },
});

const props = defineProps({
  /**
   * 当前路由对象
   */
  item: {
    type: Object as PropType<RouteRecordRaw>,
    required: true,
  },

  /**
   * 父级完整路径
   */
  basePath: {
    type: String,
    required: true,
  },

  /**
   * 是否为嵌套路由
   */
  isNest: {
    type: Boolean,
    default: false,
  },
});

/**
 * 安全获取菜单图标
 */
function getMetaIcon(meta?: RouteRecordRaw["meta"]): string | undefined {
  return (meta as Record<string, unknown>)?.icon as string | undefined;
}

/**
 * 安全获取菜单标题
 */
function getMetaTitle(meta?: RouteRecordRaw["meta"]): string {
  return (meta as Record<string, unknown>)?.title as string;
}

/**
 * 判断是否有可见的子节点
 */
const hasVisibleChildren = (children?: RouteRecordRaw[]) => {
  return children?.some((child) => !child.meta?.hideInMenu);
};

/**
 * 获取完整路径，适配外部链接
 *
 * @param routePath 路由路径
 * @returns 绝对路径
 */
function resolvePath(routePath: string) {
  if (isExternal(routePath)) return routePath;
  if (isExternal(props.basePath)) return props.basePath;

  // 拼接父路径和当前路径
  return path.resolve(props.basePath, routePath);
}
</script>

<style lang="scss">
/* stylelint-disable no-descending-specificity */
/* 菜单图标统一样式 */
.el-menu-item,
.el-sub-menu__title {
  .el-icon {
    width: 1em !important;
    margin-right: 0 !important;
    font-size: 18px;
    color: currentcolor;
  }

  [class^="i-"] {
    width: 18px;
    height: 18px;
    font-size: 18px;
  }

  // 图标与文字间距（8-10px）
  .menu-title {
    margin-left: 8px !important;
  }
}

/* 折叠状态下的样式 - el-menu :collapse=true 时自动添加 el-menu--collapse */
.el-menu--collapse {
  .el-menu-item,
  .el-sub-menu > .el-sub-menu__title {
    [class^="i-"] {
      width: 18px !important;
      min-width: 18px !important;
      height: 18px !important;
      font-size: 18px !important;
    }

    // 隐藏菜单文字（不依赖 .hideSidebar 类）
    .menu-title {
      display: none !important;
    }
  }

  /* tooltip 弹出层中的图标 */
  .el-tooltip__trigger {
    [class^="i-"] {
      width: 18px !important;
      min-width: 18px !important;
      height: 18px !important;
      font-size: 18px !important;
    }
  }

  // 折叠时隐藏子菜单箭头
  .el-sub-menu__icon-arrow {
    display: none;
  }
}

/* hideSidebar 状态下的图标 */
.hideSidebar {
  [class^="i-"] {
    width: 18px !important;
    min-width: 18px !important;
    height: 18px !important;
    font-size: 18px !important;
  }

  .submenu-title-noDropdown {
    position: relative;

    & > span {
      display: inline-block;
      visibility: hidden;
      width: 0;
      height: 0;
      overflow: hidden;
    }
  }

  .el-sub-menu {
    overflow: hidden;

    & > .el-sub-menu__title {
      .sub-el-icon {
        margin-left: 19px;
      }

      .el-sub-menu__icon-arrow {
        display: none;
      }
    }
  }

  .el-menu--collapse {
    width: $sidebar-width-collapsed;

    .el-sub-menu {
      & > .el-sub-menu__title > span {
        display: inline-block;
        visibility: hidden;
        width: 0;
        height: 0;
        overflow: hidden;
      }
    }
  }
}

html.dark {
  .el-menu-item:hover {
    background-color: $menu-hover;
  }
}

html.sidebar-color-blue {
  .el-menu-item:hover {
    background-color: $menu-hover;
  }
}

// 父菜单激活状态样式 - 当子菜单激活时，父菜单显示激活状态
.el-sub-menu {
  // 当父菜单包含激活子菜单时的样式
  &.has-active-child > .el-sub-menu__title {
    color: var(--el-color-primary) !important;
    background-color: var(--el-color-primary-light-9) !important;

    .menu-icon {
      color: var(--el-color-primary) !important;
    }
  }

  // 深色主题下的父菜单激活状态"
  html.dark & {
    &.has-active-child > .el-sub-menu__title {
      color: var(--el-color-primary-light-3) !important;
      background-color: rgba(64, 128, 255, 0.15) !important;

      .menu-icon {
        color: var(--el-color-primary-light-3) !important;
      }
    }
  }

  // 深蓝色侧边栏配色下的父菜单激活状态"
  html.sidebar-color-blue & {
    &.has-active-child > .el-sub-menu__title {
      color: var(--el-color-primary-light-3) !important;
      background-color: rgba(64, 128, 255, 0.2) !important;

      .menu-icon {
        color: var(--el-color-primary-light-3) !important;
      }
    }
  }
}
/* stylelint-enable no-descending-specificity */
</style>
