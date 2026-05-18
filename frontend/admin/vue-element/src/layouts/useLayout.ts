/**
 * 布局 Composable
 *
 * 整合布局状态、设备检测、菜单数据
 */
import { computed, watchEffect } from "vue";
import { useRoute } from "vue-router";

import { useAccessStore } from "@/stores";
import { preferences, preferencesManager, usePreferences } from "@/core/preferences";

export function useLayout() {
  const route = useRoute();
  const permissionStore = useAccessStore();
  const { isMobile, sidebarCollapsed, appPreferences, tabbarPreferences, logoPreferences } =
    usePreferences();

  // ============================================
  // 设备检测
  // ============================================

  const isDesktop = computed(() => !isMobile.value);

  // ============================================
  // 设备检测
  // ============================================

  // 监听窗口变化，自动调整设备类型和侧边栏
  watchEffect(() => {
    // 只在首次或值真正变化时才更新
    const currentMobile = preferences.app.isMobile;
    if (isMobile.value !== currentMobile) {
      preferencesManager.updatePreferences({
        app: { isMobile: isMobile.value },
      });
    }

    // 根据设备类型自动展开/收起侧边栏
    const currentCollapsed = preferences.sidebar.collapsed;
    if (isDesktop.value && currentCollapsed) {
      // 桌面端且当前是收起状态，则展开
      preferencesManager.updatePreferences({
        sidebar: { collapsed: false },
      });
    } else if (isMobile.value && !currentCollapsed) {
      // 移动端且当前是展开状态，则收起
      preferencesManager.updatePreferences({
        sidebar: { collapsed: true },
      });
    }
  });

  // ============================================
  // 布局状态
  // ============================================

  const currentLayout = computed(() => appPreferences.value.layout);
  const isSidebarOpen = computed(() => !sidebarCollapsed.value);
  const showTagsView = computed(() => tabbarPreferences.value.enable);
  const showSettings = computed(() => appPreferences.value.enablePreferences);
  const showLogo = computed(() => logoPreferences.value.enable);

  const layoutClass = computed(() => ({
    hideSidebar: sidebarCollapsed.value,
    openSidebar: !sidebarCollapsed.value,
    mobile: isMobile.value,
    [`layout-${appPreferences.value.layout}`]: true,
  }));

  // ============================================
  // 菜单数据
  // ============================================

  /** 路由列表（左侧/顶部菜单） */
  const routes = computed(() => permissionStore.accessRoutes);

  /** 混合布局侧边菜单（根据顶级菜单路径动态计算） */
  const sideMenuRoutes = computed(() => {
    const topMenuPath = activeTopMenuPath.value;
    // 从所有路由中找到匹配的顶级菜单
    const topMenu = permissionStore.accessRoutes.find((route) => route.path === topMenuPath);

    if (!topMenu?.children) {
      return [];
    }

    // 过滤掉隐藏的菜单
    return topMenu.children.filter((child) => !child.meta?.hidden);
  });

  /** 顶部菜单激活路径（仅混合布局使用） */
  const activeTopMenuPath = computed(() => {
    const path = route.path;
    // 提取第一段路径作为顶级菜单
    // /system/user → /system
    // /dashboard → /dashboard
    const segments = path.split("/").filter(Boolean);
    return segments.length > 0 ? `/${segments[0]}` : "/";
  });

  /** 当前激活菜单 */
  const activeMenu = computed(() => {
    const { meta, path } = route;
    return meta?.activeMenu || path;
  });

  // ============================================
  // 操作方法
  // ============================================

  function toggleSidebar() {
    preferencesManager.updatePreferences({
      sidebar: { collapsed: !sidebarCollapsed.value },
    });
  }

  function closeSidebar() {
    preferencesManager.updatePreferences({
      sidebar: { collapsed: true },
    });
  }

  return {
    // 设备
    isDesktop,
    isMobile,
    // 布局
    currentLayout,
    layoutClass,
    isSidebarOpen,
    showTagsView,
    showSettings,
    showLogo,
    // 菜单
    routes,
    sideMenuRoutes,
    activeMenu,
    activeTopMenuPath,
    // 方法
    toggleSidebar,
    closeSidebar,
  };
}
