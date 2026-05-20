<template>
  <el-dropdown trigger="click" @command="handleDarkChange">
    <el-icon :size="20">
      <component :is="isDark ? Sunny : Moon" />
    </el-icon>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          v-for="item in themeList"
          :key="item.value"
          :command="item.value"
          :disabled="preferences.theme.mode === item.value"
        >
          <el-icon>
            <component :is="item.component" />
          </el-icon>
          {{ item.label }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { Moon, Sunny, Monitor } from "@element-plus/icons-vue";
import { preferences, usePreferences } from "@/core/preferences";

const { t } = useI18n();
const { isDark, setTheme } = usePreferences();

// 使用计算属性，确保语言切换时自动更新
const themeList = computed(() => [
  { label: t("core.theme.light"), value: "light", component: Sunny },
  { label: t("core.theme.dark"), value: "dark", component: Moon },
  { label: t("core.theme.auto"), value: "auto", component: Monitor },
]);

const handleDarkChange = (theme: ThemeModeType) => {
  setTheme(theme);
};
</script>
