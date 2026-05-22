// 使用 Vite glob 异步加载扩展模块

export const enUSModules = import.meta.glob<Record<string, any>>(
    './**/*.json',
    {
        eager: false,      // 异步加载，减小首屏体积
        import: 'default', // 直接获取 JSON 内容
    }
);

// 路径解析工具：'./system/user.json' → 'system.user'
export const parseModulePath = (path: string): string | null => {
    const match = path.match(/\.\/(.+)\.json$/);
    if (!match) return null;
    return match[1].replace(/\//g, '.');
};

// 按需加载单个模块
export const loadModule = async (namespace: string) => {
    const path = `./${namespace.replace(/\./g, '/')}.json`;
    const loader = enUSModules[path];

    if (!loader) {
        console.warn(`[locales] Module not found: ${path}`);
        return null;
    }

    try {
        const module = await loader();
        // 返回 default 导出（JSON 内容）
        return module.default || module;
    } catch (error) {
        console.error(`[locales] Failed to load ${path}:`, error);
        return null;
    }
};
