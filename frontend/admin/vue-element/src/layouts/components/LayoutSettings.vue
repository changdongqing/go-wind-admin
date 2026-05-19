<template>
  <el-drawer
    v-model="drawerVisible"
    size="380"
    :title="t('pages.settings.project')"
    :before-close="handleCloseDrawer"
    class="settings-drawer"
  >
    <div class="settings-content">
      <section class="config-section">
        <el-divider>{{ t("pages.settings.theme") }}</el-divider>

        <div class="flex-center">
          <el-switch
            v-model="isDark"
            active-icon="Moon"
            inactive-icon="Sunny"
            class="theme-switch"
            @change="handleThemeChange"
          />
        </div>
      </section>

      <!-- 界面设置 -->
      <section class="config-section">
        <el-divider>{{ t("pages.settings.interface") }}</el-divider>

        <div class="config-item flex-x-between">
          <span class="text-xs">{{ t("pages.settings.themeColor") }}</span>
          <el-color-picker
            v-model="selectedThemeColor"
            :predefine="colorPresets"
            popper-class="theme-picker-dropdown"
          />
        </div>

        <div class="config-item flex-x-between">
          <span class="text-xs">{{ t("pages.settings.showTagsView") }}</span>
          <el-switch
            v-model="preferences.tabbar.enable"
            @change="
              (value) =>
                preferencesManager.updatePreferences({ tabbar: { enable: value as boolean } })
            "
          />
        </div>

        <div class="config-item flex-x-between">
          <span class="text-xs">{{ t("pages.settings.showAppLogo") }}</span>
          <el-switch
            v-model="preferences.logo.enable"
            @change="
              (value) =>
                preferencesManager.updatePreferences({ logo: { enable: value as boolean } })
            "
          />
        </div>

        <div class="config-item flex-x-between">
          <span class="text-xs">{{ t("pages.settings.showWatermark") }}</span>
          <el-switch v-model="preferences.app.watermark" />
        </div>

        <div class="config-item flex-x-between">
          <span class="text-xs">{{ t("pages.settings.pageSwitchingAnimation") }}</span>
          <el-select
            v-model="preferences.transition.name"
            style="width: 150px"
            @change="
              (value) =>
                preferencesManager.updatePreferences({
                  transition: { name: value as string },
                })
            "
          >
            <el-option
              v-for="item in pageSwitchingAnimationOptions"
              :key="item.value"
              :label="t(`pages.settings.${item.value}`)"
              :value="item.value"
            />
          </el-select>
        </div>

        <div class="config-item flex-x-between">
          <span class="text-xs">{{ t("pages.settings.grayMode") }}</span>
          <el-switch v-model="preferences.app.colorGrayMode" />
        </div>

        <div class="config-item flex-x-between">
          <span class="text-xs">{{ t("pages.settings.colorWeakMode") }}</span>
          <el-switch v-model="preferences.app.colorWeakMode" />
        </div>

        <div v-if="!isDark" class="config-item flex-x-between">
          <span class="text-xs">{{ t("pages.settings.sidebarColorScheme") }}</span>
          <el-radio-group v-model="sidebarColor" @change="changeSidebarColor">
            <el-radio :value="SidebarColor.CLASSIC_BLUE">
              {{ t("pages.settings.classicBlue") }}
            </el-radio>
            <el-radio :value="SidebarColor.MINIMAL_WHITE">
              {{ t("pages.settings.minimalWhite") }}
            </el-radio>
          </el-radio-group>
        </div>
      </section>

      <!-- 布局设置 -->
      <section class="config-section">
        <el-divider>{{ t("pages.settings.navigation") }}</el-divider>

        <!-- 整合的布局选择 -->
        <div class="layout-select">
          <div class="layout-grid">
            <el-tooltip
              v-for="item in layoutOptions"
              :key="item.value"
              :content="item.label"
              placement="bottom"
            >
              <div
                role="button"
                tabindex="0"
                :class="[
                  'layout-item',
                  item.className,
                  {
                    'is-active': preferences.app.layout === item.value,
                  },
                ]"
                @click="handleLayoutChange(item.value)"
                @keydown.enter.space="handleLayoutChange(item.value)"
              >
                <!-- 布局预览图标 -->
                <div class="layout-preview">
                  <div v-if="item.value !== 'sidebar-nav'" class="layout-header"></div>
                  <div v-if="item.value !== 'header-nav'" class="layout-sidebar"></div>
                  <div class="layout-main"></div>
                </div>
                <!-- 布局名称 -->
                <div class="layout-name">{{ item.label }}</div>
                <!-- 选中状态指示器 -->
                <div v-if="preferences.app.layout === item.value" class="layout-check">
                  <el-icon><Check /></el-icon>
                </div>
              </div>
            </el-tooltip>
          </div>
        </div>
      </section>
    </div>

    <!-- 操作按钮区域 - 固定到底部 -->
    <template #footer>
      <div class="action-buttons">
        <el-tooltip :content="t('settings.copyTooltip')" placement="top">
          <el-button
            type="primary"
            size="default"
            :icon="copyIcon"
            :loading="copyLoading"
            @click="handleCopySettings"
          >
            {{ copyLoading ? t("pages.settings.copying") : t("pages.settings.copyConfig") }}
          </el-button>
        </el-tooltip>
        <el-tooltip :content="t('settings.resetTooltip')" placement="top">
          <el-button
            type="warning"
            size="default"
            :icon="resetIcon"
            :loading="resetLoading"
            @click="handleResetSettings"
          >
            {{ resetLoading ? t("pages.settings.resetting") : t("pages.settings.resetConfig") }}
          </el-button>
        </el-tooltip>
      </div>
    </template>
  </el-drawer>
</template>

<script setup lang="ts">
import { DocumentCopy, RefreshLeft, Check } from "@element-plus/icons-vue";

import { SidebarColor } from "@/constants";
import { usePreferences, preferencesManager, preferences } from "@/core/preferences";

const { t } = useI18n();

// 页面切换动画选项
const pageSwitchingAnimationOptions: Record<string, any> = {
  none: { value: "none", label: "none" },
  fade: { value: "fade", label: "fade" },
  "fade-slide": { value: "fade-slide", label: "ade-slide" },
  "fade-scale": { value: "fade-scale", label: "fade-scale" },
};

// 按钮图标
const copyIcon = markRaw(DocumentCopy);
const resetIcon = markRaw(RefreshLeft);

// 加载状态
const copyLoading = ref(false);
const resetLoading = ref(false);

// 布局选项配置
interface LayoutOption {
  value: LayoutType;
  label: string;
  className: string;
}

const layoutOptions: LayoutOption[] = [
  { value: "sidebar-nav", label: t("pages.settings.leftLayout"), className: "left" },
  { value: "header-nav", label: t("pages.settings.topLayout"), className: "top" },
  { value: "mixed-nav", label: t("pages.settings.mixLayout"), className: "mix" },
];

// 颜色预设（用于颜色选择器）
const colorPresets = [
  "#4080FF",
  "#1890FF",
  "#409EFF",
  "#FA8C16",
  "#722ED1",
  "#13C2C2",
  "#52C41A",
  "#F5222D",
  "#2F54EB",
  "#EB2F96",
];

const { isDark } = usePreferences();

// 注入设置面板可见性状态
const settingsVisible = inject<Ref<boolean>>("settingsVisible", ref(false));

// TODO: 以下字段需要在 preferences 中添加对应字段后迁移
// - sidebarColorScheme
// - themeColor
// 暂时使用固定值
const SIDEBAR_COLOR_SCHEME = "minimal-white"; // 固定为极简白色
const THEME_COLOR = "#409eff"; // 固定为主题蓝色

const sidebarColor = ref(SIDEBAR_COLOR_SCHEME);

const selectedThemeColor = computed({
  get: () => THEME_COLOR,
  set: () => {
    // TODO: 后续迁移到 preferences 后实现
    console.warn(t("pages.settings.themeColorNotImplemented"));
  },
});

const drawerVisible = computed({
  get: () => settingsVisible.value,
  set: (value) => (settingsVisible.value = value),
});

/**
 * 处理主题切换
 *
 * @param value 是否启用暗黑模式
 */
const handleThemeChange = (value: string | number | boolean) => {
  preferencesManager.updatePreferences({
    theme: { mode: value ? "dark" : "light" },
  });
};

/**
 * 更改侧边栏颜色
 *
 * @param _val 颜色方案名称
 */
const changeSidebarColor = () => {
  // TODO: 后续迁移到 preferences 后实现
  console.warn(t("pages.settings.sidebarColorNotImplemented"));
};

/**
 * 切换布局
 *
 * @param layout - 布局模式
 */
const handleLayoutChange = (layout: LayoutType) => {
  if (preferences.app.layout === layout) return;

  preferencesManager.updatePreferences({
    app: { layout },
  });
};

/**
 * 复制当前配置
 */
const handleCopySettings = async () => {
  try {
    copyLoading.value = true;

    // 生成配置代码
    const configCode = generateSettingsCode();

    // 复制到剪贴板
    await navigator.clipboard.writeText(configCode);

    // 显示成功消息
    ElMessage.success({
      message: t("pages.settings.copySuccess"),
      duration: 3000,
    });
  } catch {
    ElMessage.error(t("pages.settings.copyFailed"));
  } finally {
    copyLoading.value = false;
  }
};

/**
 * 重置为默认配置
 */
const handleResetSettings = async () => {
  resetLoading.value = true;

  try {
    // 重置 preferences
    preferencesManager.resetPreferences();

    // 同步更新本地状态
    sidebarColor.value = SIDEBAR_COLOR_SCHEME;

    ElMessage.success(t("pages.settings.resetSuccess"));
  } catch {
    ElMessage.error(t("pages.settings.resetFailed"));
  } finally {
    resetLoading.value = false;
  }
};

/**
 * 生成配置代码字符串
 */
const generateSettingsCode = (): string => {
  const settings = {
    title: "pkg.name",
    version: "pkg.version",
    showSettings: true,
    showTagsView: preferences.tabbar.enable,
    showAppLogo: preferences.logo.enable,
    layout: `"${preferences.app.layout}"`,
    theme: `ThemeMode.${isDark.value ? "DARK" : "LIGHT"}`,
    size: "ComponentSize.DEFAULT",
    language: "LanguageEnum.ZH_CN",
    themeColor: `"${THEME_COLOR}"`,
    showWatermark: preferences.app.watermark,
    watermarkContent: "pkg.name",
    sidebarColorScheme: `SidebarColor.${SIDEBAR_COLOR_SCHEME.toUpperCase().replace("-", "_")}`,
  };

  return `const defaultSettings: AppSettings = {
  title: ${settings.title},
  version: ${settings.version},
  showSettings: ${settings.showSettings},
  showTagsView: ${settings.showTagsView},
  showAppLogo: ${settings.showAppLogo},
  layout: ${settings.layout},
  theme: ${settings.theme},
  size: ${settings.size},
  language: ${settings.language},
  themeColor: ${settings.themeColor},
  showWatermark: ${settings.showWatermark},
  watermarkContent: ${settings.watermarkContent},
  sidebarColorScheme: ${settings.sidebarColorScheme},
};`;
};

/**
 * 关闭抽屉前的回调
 */
const handleCloseDrawer = () => {
  settingsVisible.value = false;
};
</script>

<style lang="scss" scoped>
/* 设置抽屉样式 */
.settings-drawer {
  :deep(.el-drawer__body) {
    position: relative;
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 0;
    overflow: hidden;
  }
}

/* 设置内容区域 */
.settings-content {
  /* let drawer body control height with flex and make this area scrollable */
  flex: 1 1 auto;
  padding: 20px;
  overflow-y: auto;
}

/* 底部操作区域样式 */
.action-buttons {
  display: flex;

  & > .el-button {
    flex: 1;
    font-size: 14px;
    border-radius: 8px;
    transition: all 0.3s ease;

    &:hover {
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
      transform: translateY(-2px);
    }
  }
}
/* 主题切换器优化 */
.theme-switch {
  transform: scale(1.2);
  transition: all 0.3s ease;

  &:hover {
    transform: scale(1.25);
  }
}

.config-section {
  margin-bottom: 24px;

  .config-item {
    padding: 12px 0;
    border-bottom: 1px solid var(--el-border-color-light);
    transition: all 0.3s ease;

    &:last-child {
      border-bottom: none;
    }

    &:hover {
      padding-right: 8px;
      padding-left: 8px;
      margin: 0 -8px;
      background-color: var(--el-fill-color-light);
      border-radius: 6px;
    }
  }
}

/* 布局选择器样式优化 */
.layout-select {
  padding: 16px 8px;

  .layout-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
    justify-items: center;
  }
}

.layout-item {
  position: relative;
  width: 70px;
  height: 80px;
  overflow: hidden;
  cursor: pointer;
  background: linear-gradient(145deg, var(--el-bg-color) 0%, var(--el-bg-color-page) 100%);
  border: 2px solid var(--el-border-color);
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);

  &:hover {
    background: linear-gradient(
      145deg,
      var(--el-bg-color) 0%,
      var(--el-color-primary-light-9) 100%
    );
    border-color: var(--el-color-primary-light-3);
    transform: translateY(-4px) scale(1.05);
  }

  &:active {
    transform: translateY(-2px) scale(1.02);
  }

  .layout-preview {
    position: relative;
    width: 100%;
    height: 50px;
    margin: 8px 0 4px 0;
  }

  .layout-header {
    position: absolute;
    top: 0;
    right: 4px;
    left: 4px;
    height: 8px;
    background: linear-gradient(
      90deg,
      var(--el-color-primary) 0%,
      var(--el-color-primary-light-3) 100%
    );
    border-radius: 2px;
  }

  .layout-sidebar {
    position: absolute;
    left: 4px;
    width: 12px;
    background: linear-gradient(
      180deg,
      var(--el-color-primary-dark-2) 0%,
      var(--el-color-primary) 100%
    );
    border-radius: 2px;
  }

  .layout-main {
    position: absolute;
    background: linear-gradient(135deg, var(--el-fill-color-light) 0%, var(--el-fill-color) 100%);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 2px;
  }

  .layout-name {
    position: absolute;
    right: 0;
    bottom: 6px;
    left: 0;
    font-size: 10px;
    font-weight: 500;
    color: var(--el-text-color-regular);
    text-align: center;
    transition: color 0.3s ease;
  }

  .layout-check {
    position: absolute;
    top: 4px;
    right: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    font-size: 10px;
    color: white;
    background: var(--el-color-success);
    border-radius: 50%;
  }

  // 左侧布局
  &.left {
    .layout-sidebar {
      top: 4px;
      bottom: 4px;
    }
    .layout-main {
      top: 4px;
      right: 4px;
      bottom: 4px;
      left: 20px;
    }
  }

  // 顶部布局
  &.top {
    .layout-header {
      height: 12px;
    }
    .layout-main {
      top: 16px;
      right: 4px;
      bottom: 4px;
      left: 4px;
    }
  }

  // 混合布局
  &.mix {
    .layout-header {
      height: 10px;
    }
    .layout-sidebar {
      top: 14px;
      bottom: 4px;
    }
    .layout-main {
      top: 14px;
      right: 4px;
      bottom: 4px;
      left: 20px;
    }
  }

  &.is-active {
    background: linear-gradient(
      145deg,
      var(--el-color-primary-light-9) 0%,
      var(--el-color-primary-light-8) 100%
    );
    border-color: var(--el-color-primary);
    transform: translateY(-2px) scale(1.08);

    .layout-name {
      font-weight: 600;
      color: var(--el-color-primary);
    }
  }
}

:deep(.copy-config-dialog) {
  .el-message-box__content {
    max-height: 400px;
    overflow-y: auto;
  }
}
</style>
