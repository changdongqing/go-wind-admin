import type {PluginOption} from 'vite';
import path from 'path';
import AutoImport from 'unplugin-auto-import/vite';

/**
 * 自动导入处理
 *
 * 配置要点（避免 Vite "Duplicated imports" 警告）：
 * 1. imports 只保留 react-router-dom（项目实际依赖），不再同时列 react-router——
 *    两者都导出同名 hooks(useNavigate/useLocation/...)，同时注册会产生重复声明。
 * 2. dirs 用精确的 barrel 通配（指向各 index.ts），而不是 src/components/** 递归全量——
 *    后者会把 Editor/index.ts(barrel) 和 Editor/src/CodeEditor.tsx(被 re-export 的源文件)
 *    都当独立模块根，导致同一符号(CodeEditor)被注册两次 → 重复警告。
 */
export const autoImportPlugin = (): PluginOption => {
    return AutoImport({
        dirs: [
            // hooks 目录扁平，递归扫描无害（无 barrel/源文件重复）。
            // types:false —— 这些模块里的 type/interface（如 ECOption）都被业务代码显式
            // import，若同时全局自动注册会触发 Vite "Duplicated imports" 警告。
            {glob: 'src/hooks', types: false},
            // 只扫各层 index barrel，避免 src 子文件与 barrel 重复。
            // 同样 types:false —— barrel 里的 *Props/*State 等类型均被显式导入。
            {glob: 'src/components/**/index.{ts,tsx}', types: false},
            {glob: 'src/stores', types: false},
            {glob: 'types', types: false},
        ],
        imports: [
            'react',
            'react-router-dom',
            'react-i18next',
            {from: 'react', imports: ['FC'], type: true},
        ],
        dts: 'types/autoImports.d.ts',
        include: [/\.[tj]sx?$/],
        resolvers: [
            (name) => {
                // 处理 @/ 开头的路径别名
                if (name.startsWith('@/')) {
                    return {
                        from: name.replace('@/', path.resolve(__dirname, 'src/') + '/'),
                    };
                }
            },
        ],
    });
};
