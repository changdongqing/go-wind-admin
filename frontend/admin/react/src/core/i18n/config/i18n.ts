import i18n from 'i18next';
import {initReactI18next} from 'react-i18next';
import {resources, type SupportedLocale, type ModuleLoader} from '@/locales';


// 扩展模块加载器（按需加载）
const loadModule: ModuleLoader = async (lang: string, namespace: string) => {
    // 动态导入对应语言的模块加载器
    const {loadModule} = await import(`@/locales/${lang}/index.ts`);
    return loadModule(namespace);
};

export const initI18n = async (initialLang: SupportedLocale) => {
    await i18n
        .use(initReactI18next)
        .init({
            lng: initialLang, // 设置初始语言
            resources, // 核心命名空间预加载
            fallbackLng: 'zh-CN',
            supportedLngs: ['zh-CN', 'en-US'],

            // 命名空间配置
            defaultNS: 'common',
            ns: ['common', 'auth', 'menu'], // 只预加载核心命名空间

            // 后端动态词条（可选）
            backend: {
                loadPath: '/api/i18n/{{lng}}/{{ns}}',
            },

            // 缺失 key 处理（开发环境）
            missingKeyHandler: import.meta.env.DEV
                ? (lngs, ns, key) => {
                    console.warn(`[i18n] Missing: "${key}" in "${ns}" for "${lngs[0]}"`);
                }
                : undefined,
        });

    // 注册扩展模块加载器（按需）
    i18n.services.backendConnector.loading = async (
        lng: string,
        ns: string
    ) => {
        // 核心命名空间已由 resources 预加载，跳过
        if (['common', 'auth', 'menu'].includes(ns)) {
            return {};
        }

        // 扩展命名空间：按需加载
        return await loadModule(lng, ns);
    };

    return i18n;
};
