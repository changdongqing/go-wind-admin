# GoWind React Admin 脚手架开发指南

## 项目概述

GoWind React Admin 是基于 React 19 的企业级后台管理脚手架。

## 技术栈

- React 19 + TypeScript 6
- Ant Design v6 + ProComponents 2
- Vite 8 (SWC)
- Zustand 5 + TanStack React Query 5
- React Router v6
- i18next + react-i18next
- Less + UnoCSS
- Iconify (lucide 图标集)
- pnpm

## 目录结构

```
src/
├── api/                    # API 层（三层架构）
│   ├── generated/          # 自动生成代码（禁止手动修改）
│   ├── service/            # Service 层 - API 调用纯函数封装
│   └── hooks/              # Hooks 层 - React Query 集成
├── core/                   # 核心模块（access/i18n/preferences/router/storage/transport）
├── hooks/                  # 业务 Hooks
├── layouts/                # 布局组件
├── locales/                # 翻译资源（zh-CN/en-US）
├── pages/                  # 页面（app/ 业务，core/ 系统）
├── router/                 # 路由配置（config/guards/modules）
├── stores/                 # Zustand Stores
├── styles/                 # 全局样式
└── utils/                  # 工具函数
```

## API 三层架构

```
Generated (自动生成) → Service (纯函数封装) → Hooks (React Query 集成)
```

### 使用规则

| 场景 | 方式 |
|------|------|
| React 组件 | `useXxx()` Hook（`api/hooks/`） |
| Zustand Store / 路由守卫 / 工具函数 | `fetchXxx()` 方法（`api/hooks/`） |

### 命名规范

- Service 层：`listXxx()`, `getXxx()`, `createXxx()`, `updateXxx()`, `deleteXxx()`
- Hooks 层：`useListXxx()`, `useGetXxx()` + `fetchListXxx()`, `fetchXxx()`

### Service 层模板

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

### Hooks 层模板

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

`src/router/modules/*.tsx` 通过 `import.meta.glob` 自动导入。

### 路由配置模板

```tsx
import type { AppRouteObject } from '@/core/router/types';
import { createLazyRoute } from '@/core/router';

export const myModuleRoutes: AppRouteObject[] = [
  {
    name: 'my-module',
    path: 'my-module',
    meta: {
      title: 'routes:myModule',
      icon: 'lucide:some-icon',
      order: 10,
      authority: ['sys:my_module:view'],
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
| `icon` | `string` | Iconify 图标名 |
| `order` | `number` | 菜单排序 |
| `authority` | `string[]` | 角色码和权限码混合数组 |
| `hideInMenu` | `boolean` | 隐藏菜单项 |
| `hideInTab` | `boolean` | 隐藏标签页 |
| `keepAlive` | `boolean` | 缓存页面 |

### 权限模式

- `frontend`（默认）：前端路由 + `meta.authority` 过滤
- `backend`：后端返回菜单 + `pageMap` 动态匹配

## 权限系统

### 数据来源（分离存储）

- 角色码 → `useUserStore.userRoles`
- 权限码 → `useUserStore.accessCodes`
- `meta.authority` → 角色码和权限码混合数组

### 三种鉴权方式

```tsx
// 1. useAccess Hook（推荐）
const { hasAccessByCodes, hasAccessByRoles } = useAccess();
{hasAccessByCodes(['sys:user:create']) && <Button>新建</Button>}

// 2. AccessControl 组件
<AccessControl codes={['sys:user:create']} type="code">
  <Button>新建</Button>
</AccessControl>

// 3. 非组件场景
const { hasAccessByCodes } = getAccessStatic();
```

- 权限码格式：`模块:资源:操作`（如 `sys:user:create`）
- 超级管理员角色：`*:*:*`，自动通过所有检查

## 状态管理

| Store | 文件 | 用途 | 持久化 |
|-------|------|------|--------|
| `useAuthStore` | `stores/auth.ts` | Token、登录/登出 | token |
| `useUserStore` | `stores/user.ts` | 用户信息、角色码、权限码 | userInfo |
| `usePreferencesStore` | `core/preferences/store/` | 偏好设置 | 全部 |

```typescript
// React 组件中 — selector 精确订阅
const token = useAuthStore((s) => s.accessToken);

// 非组件环境
const token = useAuthStore.getState().accessToken;
```

## 国际化

### 命名空间

| 类别 | 目录 | 示例 |
|------|------|------|
| 核心 | `_core/` | `common`, `auth`, `routes`, `editor` |
| 业务 | `_modules/` | `user`, `role`, `dashboard` 等 |

### 使用方式

```tsx
import { useI18n } from '@/core/i18n';
const { t } = useI18n('user');     // 指定命名空间
t('username');                      // 查找 user 命名空间
```

### 新增翻译

在 `src/locales/zh-CN/_modules/` 和 `en-US/_modules/` 创建同名 JSON，自动收集。

### 翻译键规则

- 插值用 `{{var}}`，**不是** `#{var}`
- 路由标题用 `'routes:xxx'` 格式
- 硬编码文本必须提取到翻译文件

## 代码风格

- Prettier: 单引号、分号、尾逗号 `all`、行宽 100、2 空格缩进、LF 换行
- 路径别名: `@/` → `src/`，`#/` → `types/`
- ESLint: TypeScript 严格模式，React Hooks 规则强制
- 提交: Conventional Commits

## 关键注意事项

1. **PaginationQuery 必须用 new**: `new PaginationQuery({ page, pageSize })`
2. **非组件环境禁用 useXxx Hook**: 只能用 `fetchXxx()` 或 Service 层
3. **国际化插值**: `{{var}}` 而非 `#{var}`
4. **meta.title 格式**: `'routes:xxx'`
5. **禁止修改 generated 目录**: 由工具自动生成
6. **DrawerForm 用 formRef**: 没有 `useForm`
7. **antd v6**: `items` 替代 `TabPane`，Alert 用 `title` 替代 `message`
8. **ProTable scroll.y**: 初始值必须是像素值（数字）
9. **角色码和权限码分离**: `userRoles` + `accessCodes`，不混合
10. **不要使用 `userInfo?.permissions`**: 该字段不存在

## 新增完整功能模块清单

1. 翻译: `src/locales/{zh-CN,en-US}/_modules/xxx.json`
2. 页面: `src/pages/app/xxx/index.tsx`
3. API: `src/api/service/xxx.ts` + `src/api/hooks/xxx.ts` + 导出
4. 路由: `src/router/modules/xxx.tsx`（自动导入）
5. 权限: 配置 `meta.authority` + 页面中使用 `useAccess()`
