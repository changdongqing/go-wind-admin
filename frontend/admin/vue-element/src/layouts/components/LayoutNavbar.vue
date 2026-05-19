<template>
  <div class="navbar">
    <div class="flex-y-center">
      <!-- 菜单折叠按钮 -->
      <Hamburger :is-active="isSidebarOpened" @toggle-click="toggleSideBar" />
      <!-- 面包屑导行栏-->
      <Breadcrumb />
    </div>
    <!-- 导航栏操作区域-->
    <div class="navbar__actions">
      <LayoutToolbar />
    </div>
  </div>
</template>

<script setup lang="ts">
import { preferencesManager } from "@/core/preferences";
import { usePreferences } from "@/core/preferences";

import Hamburger from "@/components/Hamburger/index.vue";
import Breadcrumb from "@/components/Breadcrumb/index.vue";

const { sidebarCollapsed } = usePreferences();

const isSidebarOpened = computed(() => !sidebarCollapsed.value);

function toggleSideBar() {
  preferencesManager.updatePreferences({
    sidebar: { collapsed: !sidebarCollapsed.value },
  });
}
</script>

<style lang="scss" scoped>
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: $navbar-height;
  background-color: var(--sidebar-logo-background);  // 与 Logo 区保持一致
  border-bottom: 1px solid var(--sidebar-logo-border-color);  // 底部分割线

  &__actions {
    display: flex;
    align-items: center;
    height: 100%;
  }
}
</style>
