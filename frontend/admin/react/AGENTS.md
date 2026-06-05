# AGENTS.md - GoWind React Admin 脚手架开发指南

> 本文件面向二开人员，帮助 AI 助手理解项目架构并遵循开发规范。

## 项目概述

GoWind React Admin 是基于 React 19 的企业级后台管理脚手架，采用 Ant Design v6 + ProComponents 作为 UI 框架。

## 技术栈

- React 19 + TypeScript 6
- Ant Design v6 + ProComponents 2
- Vite 8 (SWC 编译)
- Zustand 5 (状态管理) + TanStack React Query 5 (数据请求)
- React Router v6
- i18next + react-i18next (国际化)
- Less + UnoCSS (样式)
- Iconify (图标，lucide 图标集)
- pnpm (包管理)

## 目录结构

```
src/
├── api/                    # API 层（三层架构）
│   ├── generated/          # 自动生成代码（禁止手动修改）
│   ├── service/            # Service 层 - API 调用纯函数封装
│   └── hooks/              # Hooks 层 - React Query 集成
├── core/                   # 核心模块（access/i18n/preferences/router/storage/transport）
├── hooks/                  # 业务 Hooks
├── layouts/                # 布局组件（MainLayout/BlankLayout/UserLayout/IFrameLayout）
├── locales/                # 翻译资源（zh-CN/en-US，_core/ + _modules/）
├── pages/                  # 页面组件（app/ 业务页面，core/ 系统页面）
├── router/                 # 路由配置（config/guards/modules）
├── stores/                 # Zustand Stores（auth/user/tabs/pageRefresh）
├── styles/                 # 全局样式
└── utils/                  # 工具函数
```

## API 三层架构

```
Generated (自动生成类型和 Client) → Service (纯函数封装) → Hooks (React Query 集成)
```

**使用规则**：
- React 组件中 → `useXxx()` Hook（来自 `api/hooks/`）
- Zustand Store / 路由守卫 / 工具函数 → `fetchXxx()` 方法（来自 `api/hooks/`）
- Service 层 → 纯 async 函数，不使用 React Hook

**命名规范**：
- Service 层：`listXxx()`, `getXxx()`, `createXxx()`, `updateXxx()`, `deleteXxx()`
- Hooks 层：`useListXxx()`, `useGetXxx()` + `fetchListXxx()`, `fetchXxx()`

**Service 层模板**：

```typescript
import { createXxxServiceClient } from '@/api/generated/admin/service/v1';
import { type PaginationQuery, requestApi } from '@/core';

let _instance: ReturnType<typeof createXxxServiceClient> | null = null;
export function getXxxService() {
  if (!_instance) _instance = createXxxServiceClient(requestApi);
  return _instance;
}

export async function listXxx(query: PaginationQuery) {
  return getXxxService().List(query.toRawParams());
}
```

**Hooks 层模板**：

```typescript
import { useMutation, type UseMutationOptions } from '@tanstack/react-query';
import { listXxx } from '@/api/service/xxx';
import { type PaginationQuery, queryClient } from '@/core';

export function useListXxx(options?: UseMutationOptions<...>) {
  return useMutation({ mutationFn: (q) => listXxx(q), ...options });
}

export async function fetchListXxx(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ['listXxx', params], queryFn: () => listXxx(params), retry: 0,
  });
}
```

## 路由系统

### 路由模块（自动导入）

`src/router/modules/*.tsx` 中的路由通过 `import.meta.glob` 自动导入，合并到主布局的 children 中。

### 路由配置模板

```tsx
import type { AppRouteObject } from '@/core/router/types';
import { createLazyRoute } from '@/core/router';

export const myModuleRoutes: AppRouteObject[] = [
  {
    name: 'my-module',
    path: 'my-module',
    meta: {
      title: 'routes:myModule',       // i18n 翻译键（routes 命名空间）
      icon: 'lucide:some-icon',       // Iconify 图标名
      order: 10,                      // 菜单排序
      authority: ['sys:my_module:view'], // 权限码
    },
    children: [
      {
        name: 'my-module-list',
        path: 'list',
        element: createLazyRoute(() => import('@/pages/app/my-module')),
        meta: { title: 'routes:myModuleList' },
      },
    ],
  },
];
export default myModuleRoutes;
```

### Route Meta 关键字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `title` | `string` | `'routes:xxx'` 格式的 i18n 翻译键 |
| `icon` | `string` | Iconify 图标名，如 `'lucide:layout-dashboard'` |
| `order` | `number` | 菜单排序 |
| `authority` | `string[]` | 角色码和权限码混合数组 |
| `hideInMenu` | `boolean` | 是否隐藏菜单项 |
| `hideInTab` | `boolean` | 是否隐藏标签页 |
| `keepAlive` | `boolean` | 是否缓存页面 |

### 权限模式

- `frontend`（默认）：前端路由 + `meta.authority` 过滤
- `backend`：后端返回菜单 + `pageMap` 动态匹配组件

## 权限系统

### 数据来源（分离存储）

- 角色码 → `useUserStore.userRoles`（来自 `userInfo.roles`）
- 权限码 → `useUserStore.accessCodes`（来自 `GetMyPermissionCode`）
- `meta.authority` → 角色码和权限码的**混合数组**

### 三种鉴权方式

```tsx
// 1. useAccess Hook（推荐，用于条件渲染）
import { useAccess } from '@/core/access';
const { hasAccessByCodes, hasAccessByRoles } = useAccess();
{hasAccessByCodes(['sys:user:create']) && <Button>新建</Button>}

// 2. AccessControl 组件
import { AccessControl } from '@/core/access';
<AccessControl codes={['sys:user:create']} type="code">
  <Button>新建</Button>
</AccessControl>

// 3. 非组件场景（路由生成、工具函数）
import { getAccessStatic } from '@/core/access';
const { hasAccessByCodes } = getAccessStatic();
```

### 权限码格式

`模块:资源:操作`，如 `sys:user:create`、`sys:role:update`

### 超级管理员

拥有 `*:*:*` 角色的用户自动通过所有权限检查。

## 状态管理

### Store 列表

| Store | 文件 | 用途 | 持久化 |
|-------|------|------|--------|
| `useAuthStore` | `stores/auth.ts` | Token、登录/登出 | token 持久化 |
| `useUserStore` | `stores/user.ts` | 用户信息、角色码、权限码 | userInfo 持久化 |
| `usePreferencesStore` | `core/preferences/store/` | 偏好设置 | 全部持久化 |

### 使用方式

```typescript
// React 组件中 — 使用 selector 精确订阅
const token = useAuthStore((s) => s.accessToken);
const isDark = usePreferencesStore((s) => s.preferences.theme.mode === 'dark');

// 非组件环境 — 使用 getState()
const token = useAuthStore.getState().accessToken;
const locale = usePreferencesStore.getState().preferences.app.locale;
```

## 国际化 (i18n)

### 命名空间

| 类别 | 目录 | 命名空间 |
|------|------|----------|
| 核心 | `_core/` | `common`, `auth`, `routes`, `editor` |
| 业务 | `_modules/` | 文件名即为命名空间（如 `user`, `role`） |

### 使用方式

```tsx
import { useI18n } from '@/core/i18n';

// 指定命名空间（推荐）
const { t } = useI18n('user');
t('username');  // 查找 user 命名空间

// 多命名空间
const { t } = useI18n(['user-detail', 'user']);

// 非组件环境
import i18n from 'i18next';
i18n.t('key', { ns: 'common' });
```

### 新增翻译

在 `src/locales/zh-CN/_modules/` 和 `src/locales/en-US/_modules/` 创建同名 JSON 文件，自动被 `import.meta.glob` 收集。

### 翻译键规则

- 插值用 `{{var}}`，**不是** `#{var}`
- 路由标题用 `'routes:xxx'` 格式
- 硬编码文本必须提取到翻译文件

## 偏好设置

```tsx
import { usePreferences } from '@/core/preferences';

const { isDark, theme, app, sidebar, tabbar, updatePreferences, toggleTheme } = usePreferences();
```

## 代码风格

- **Prettier**: 单引号、分号、尾逗号 `all`、行宽 100、2 空格缩进、LF 换行
- **路径别名**: `@/` → `src/`，`#/` → `types/`
- **ESLint**: TypeScript 严格模式，React Hooks 规则强制
- **提交规范**: Conventional Commits（Husky + commitlint）

## 关键注意事项

1. **PaginationQuery 必须用 new**: `new PaginationQuery({ page, pageSize })`
2. **非组件环境禁用 useXxx Hook**: Store/路由守卫/工具函数中只能用 `fetchXxx()` 或直接调 Service 层
3. **国际化插值**: `{{var}}` 而非 `#{var}`
4. **meta.title 格式**: `'routes:xxx'`
5. **禁止修改 generated 目录**: 由工具自动生成
6. **DrawerForm 用 formRef**: 没有 `useForm` 方法
7. **antd v6 变更**: 用 `items` 替代 `TabPane`，Alert 用 `title` 替代 `message`
8. **ProTable scroll.y**: 初始值必须是像素值（数字）
9. **角色码和权限码分离存储**: `userRoles` + `accessCodes`，不混合
10. **不要使用 `userInfo?.permissions`**: 该字段不存在

## 新增完整功能模块清单

当需要新增一个完整的业务模块时，按以下顺序操作：

1. 创建翻译文件: `src/locales/zh-CN/_modules/xxx.json` + `src/locales/en-US/_modules/xxx.json`
2. 创建页面组件: `src/pages/app/xxx/index.tsx`
3. 创建 API: Service 层 (`src/api/service/xxx.ts`) + Hooks 层 (`src/api/hooks/xxx.ts`) + 导出
4. 创建路由: `src/router/modules/xxx.tsx`（自动导入）
5. 如需权限控制：配置 `meta.authority` 并在页面中使用 `useAccess()`
