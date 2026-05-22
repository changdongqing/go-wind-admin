import { initI18n } from '@/core/i18n';
import { usePreferencesStore } from '@/core/preferences';

/**
 * 应用启动初始化
 */
export async function bootstrap() {
  await _initI18n();

  // 可放全局初始化逻辑
  console.log('✅ 应用启动初始化完成');
}

async function _initI18n() {
  // 从 preferences 获取初始语言
  const initialLocale = usePreferencesStore.getState().preferences.app.locale;

  // 初始化 i18n（传入初始语言）
  const i18nInstance = await initI18n(initialLocale);

  // 调试：打印 i18n 资源
  console.log('✅ i18n 初始化完成');
  console.log('📚 当前语言:', i18nInstance.language);
  console.log('📖 可用的命名空间:', i18nInstance.options.ns);

  // 预加载 preferences 模块
  try {
    console.log('[bootstrap] 开始加载 preferences 命名空间...');

    // 直接加载模块并添加到 i18n
    let prefsData: any = null;

    if (initialLocale === 'zh-CN') {
      const { loadModule } = await import('@/locales/zh-CN/index.ts');
      prefsData = await loadModule('preferences');
    } else if (initialLocale === 'en-US') {
      const { loadModule } = await import('@/locales/en-US/index.ts');
      prefsData = await loadModule('preferences');
    }

    console.log('[bootstrap] 模块加载结果:', prefsData ? '成功' : '失败');

    if (prefsData) {
      // 手动添加资源到 i18n
      i18nInstance.addResourceBundle(initialLocale, 'preferences', prefsData, true, true);
      console.log('[bootstrap] addResourceBundle 完成');

      const prefsResource = i18nInstance.getResource(initialLocale, 'preferences');
      console.log('[bootstrap] getResource 结果:', prefsResource);

      if (prefsResource && Object.keys(prefsResource).length > 0) {
        console.log('✅ preferences 命名空间已加载');
        console.log('📦 preferences 资源键数量:', Object.keys(prefsResource).length);
        console.log('📝 示例翻译 - title:', prefsResource.title);
      } else {
        console.warn('⚠️ preferences 资源为空或不存在');
      }
    } else {
      console.error('❌ 无法加载 preferences 模块数据');
    }
  } catch (error) {
    console.error('❌ preferences 命名空间加载失败:', error);
  }
}
