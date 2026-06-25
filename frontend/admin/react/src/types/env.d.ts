/// <reference types="vite/client" />

interface ImportMetaEnv {
    readonly VITE_API_URL: string
    readonly VITE_SSE_URL: string

    readonly VITE_APP_TITLE: string
    readonly VITE_APP_NAMESPACE: string
    readonly VITE_APP_VERSION: string

    readonly VITE_AES_KEY: string

    readonly VITE_ENV: string

    readonly VITE_MOCK: boolean
    readonly VITE_ANALYZE: boolean
}

interface ImportMeta {
    readonly env: ImportMetaEnv
}

// React Router handle 类型增强
declare module 'react-router' {
    interface RouteMatch {
        handle: RouteHandle;
    }
}

// 兼容 process.env
declare namespace NodeJS {
    interface ProcessEnv extends ImportMetaEnv {
    }
}

/**
 * 运行时配置（由 public/app-config.js 在应用启动前注入到 window.__APP_CONFIG__）
 * 部署后实施人员可直接编辑 public/app-config.js 修改这些值，无需重新编译
 */
interface AppRuntimeConfig {
    /** 站点 / 应用名称：左上角标题 + 浏览器页签标题 */
    appName?: string;
    /** 登录页左侧品牌大标题（不填则用 i18n 默认文案） */
    systemTitle?: string;
    /** 登录页左侧品牌描述（不填则用 i18n 默认文案） */
    systemDescription?: string;
    /** 版权公司名称（登录页底部及后台页脚） */
    copyrightCompanyName?: string;
    /** 版权年份，如 "2026" 或 "2021-2026" */
    copyrightYear?: string;
    /** 公司网址 */
    copyrightSiteLink?: string;
    /** ICP 备案号（中国大陆部署需填写，否则留空） */
    copyrightIcp?: string;
    /** API 基础地址（如 https://api.customer-a.com）。留空则用构建期 VITE_API_URL；桌面端常配此项指向远程后端 */
    apiBaseUrl?: string;
    /** SSE 推送基础地址（如 https://sse.customer-a.com/events）。留空则用构建期 VITE_SSE_URL */
    sseUrl?: string;
    /** 桌面端自动更新源地址（如 https://updates.customer-a.com/desktop/）。留空则不检查更新；仅桌面端生效 */
    updateServerUrl?: string;
}

interface Window {
    __APP_CONFIG__?: AppRuntimeConfig;
}
