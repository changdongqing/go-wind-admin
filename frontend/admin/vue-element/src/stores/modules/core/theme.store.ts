import { defineStore } from "pinia";
import { ref, computed } from "vue";

import { ThemeMode, APP_PREFIX, STORAGE_KEYS } from "@/constants";
import { StorageManager } from "@/utils/storage";
import { toggleDarkMode } from "@/utils/theme";

// 创建主题存储管理器实例
const themeStorage = new StorageManager({
  prefix: APP_PREFIX,
  storageType: "localStorage",
});

/**
 * 主题 Store
 *
 * @description
 * 管理应用的主题模式设置（浅色/深色/自动）
 */
export const useThemeStore = defineStore("theme", () => {
  // 状态 - 只管理主题模式
  const themeMode = ref<ThemeMode>(ThemeMode.LIGHT);

  // 计算属性
  const isDark = computed(() => themeMode.value === ThemeMode.DARK);
  const isLight = computed(() => themeMode.value === ThemeMode.LIGHT);
  const isAuto = computed(() => themeMode.value === ThemeMode.AUTO);

  // 初始化时从本地存储加载主题设置
  const initTheme = () => {
    try {
      const savedTheme = themeStorage.getItem<ThemeMode>(STORAGE_KEYS.THEME_MODE);
      if (savedTheme && isValidThemeMode(savedTheme)) {
        themeMode.value = savedTheme;
      }
    } catch (error) {
      console.error("Failed to load theme preference:", error);
    }

    // 应用初始主题
    applyThemeMode(themeMode.value);
  };

  /**
   * 验证主题模式是否有效
   * @param mode 主题模式
   * @returns 是否为有效主题模式
   */
  const isValidThemeMode = (mode: string): mode is ThemeMode => {
    return [ThemeMode.LIGHT, ThemeMode.DARK, ThemeMode.AUTO].includes(mode as ThemeMode);
  };

  /**
   * 应用主题模式
   * @param mode 主题模式
   */
  const applyThemeMode = (mode: ThemeMode) => {
    try {
      // 处理自动模式
      const actualMode =
        mode === ThemeMode.AUTO
          ? window.matchMedia("(prefers-color-scheme: dark)").matches
            ? ThemeMode.DARK
            : ThemeMode.LIGHT
          : mode;

      // 更新 DOM 类名
      toggleDarkMode(actualMode === ThemeMode.DARK);

      // 保存到本地存储
      saveTheme(mode);
    } catch (error) {
      console.error("Failed to apply theme mode:", error);
    }
  };

  /**
   * 设置主题模式
   * @param mode 主题模式
   */
  const setTheme = async (mode: ThemeMode) => {
    if (!isValidThemeMode(mode)) {
      console.warn(`Invalid theme mode: ${mode}`);
      return;
    }

    if (themeMode.value === mode) {
      return; // 如果已经是当前主题，则无需更改
    }

    themeMode.value = mode;
    await applyThemeMode(mode);
  };

  /**
   * 切换主题模式（在浅色和深色之间切换）
   */
  const toggleTheme = async () => {
    const newMode = themeMode.value === ThemeMode.DARK ? ThemeMode.LIGHT : ThemeMode.DARK;
    await setTheme(newMode);
  };

  /**
   * 保存主题设置到本地存储
   * @param mode 主题模式
   */
  const saveTheme = (mode: ThemeMode) => {
    try {
      themeStorage.setItem(STORAGE_KEYS.THEME_MODE, mode);
    } catch (error) {
      console.error("Failed to save theme preference:", error);
    }
  };

  /**
   * 重置为默认主题
   */
  const resetTheme = async () => {
    themeMode.value = ThemeMode.LIGHT;
    await applyThemeMode(ThemeMode.LIGHT);

    // 清除本地存储
    themeStorage.removeItem(STORAGE_KEYS.THEME_MODE);
  };

  // 监听系统主题变化（仅在自动模式下生效）
  const setupSystemThemeListener = () => {
    if (typeof window === "undefined") return;

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const handleChange = () => {
      if (themeMode.value === ThemeMode.AUTO) {
        applyThemeMode(ThemeMode.AUTO);
      }
    };

    mediaQuery.addEventListener("change", handleChange);

    // 返回清理函数
    return () => mediaQuery.removeEventListener("change", handleChange);
  };

  // 初始化
  initTheme();
  setupSystemThemeListener();

  return {
    // 状态
    themeMode,

    // 计算属性
    isDark,
    isLight,
    isAuto,

    // 方法
    setTheme,
    toggleTheme,
    resetTheme,
  };
});
