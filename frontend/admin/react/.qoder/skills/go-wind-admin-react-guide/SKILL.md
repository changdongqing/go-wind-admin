---
name: go-wind-admin-react-guide
description: GoWind React Admin 脚手架开发指南。帮助二开人员理解项目架构、遵循开发规范、快速新增功能模块。当用户需要新增页面、新增API、新增路由模块、使用权限控制、使用国际化、理解状态管理、或询问项目架构和开发规范时触发。
---

# GoWind React Admin 脚手架开发指南

## 技术栈

- **框架**: React 19 + TypeScript 6
- **UI 库**: Ant Design v6 + ProComponents 2
- **构建**: Vite 8 + SWC
- **状态管理**: Zustand 5 (persist 中间件)
- **数据请求**: Axios + TanStack React Query 5
- **路由**: React Router v6
- **国际化**: i18next + react-i18next
- **样式**: Less + UnoCSS
- **图标**: Iconify (lucide 图标集)
- **包管理**: pnpm

## 目录结构

```
src/
├── api/                    # API 层（三层架构）
│   ├── generated/          # 自动生成代码（禁止手动修改）
│   ├── service/            # Service 层 - API 调用封装
│   └── hooks/              # Hooks 层 - React Query 集成
├── core/                   # 核心模块
│   ├── access/             # 权限控制
│   ├── i18n/               # 国际化
│   ├── preferences/        # 偏好设置
│   ├── router/             # 路由工厂
│   ├── storage/            # 存储工具
│   └── transport/          # HTTP 传输层
├── hooks/                  # 业务 Hooks
├── layouts/                # 布局组件
├── locales/                # 翻译资源
│   ├── zh-CN/              # 中文（_core/ + _modules/）
│   └── en-US/              # 英文
├── pages/                  # 页面组件
│   ├── app/                # 业务页面
│   └── core/               # 系统页面（错误页等）
├── router/                 # 路由配置
│   ├── config/             # 静态/错误/认证路由
│   ├── guards/             # 路由守卫
│   └── modules/            # 业务路由模块（自动导入）
├── stores/                 # Zustand Stores
├── styles/                 # 全局样式
└── utils/                  # 工具函数
```

## API 三层架构

```
Generated (自动生成) → Service (纯函数封装) → Hooks (React Query 集成)
```

### Service 层 (`src/api/service/*.ts`)

```typescript
// 单例模式 + 纯函数
let _instance: ReturnType<typeof createXxxServiceClient> | null = null;

export function getXxxService() {
  if (!_instance) _instance = createXxxServiceClient(requestApi);
  return _instance;
}

export async function listXxx(query: PaginationQuery) {
  return getXxxService().List(query.toRawParams());
}
```

### Hooks 层 (`src/api/hooks/*.ts`)

```typescript
// React Hook（组件中使用）
export function useListXxx(options?: UseMutationOptions<...>) {
  return useMutation({ mutationFn: (q) => listXxx(q), ...options });
}

// Fetch 方法（Store/工具函数/路由守卫中使用）
export async function fetchListXxx(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ['listXxx', params], queryFn: () => listXxx(params), retry: 0,
  });
}
```

### 使用场景

| 场景 | 方式 | 示例 |
|------|------|------|
| React 组件 | `useXxx()` | `const m = useListUsers(); m.mutateAsync(query)` |
| Zustand Store | `fetchXxx()` | `await fetchUser(id)` |
| 路由守卫 | `fetchXxx()` | `await fetchNavigation()` |

## 新增功能模块指南

### 1. 新增页面

在 `src/pages/app/` 下创建目录，如 `src/pages/app/my-module/index.tsx`。

### 2. 新增路由

在 `src/router/modules/` 下创建路由文件（自动被 `import.meta.glob` 导入）：

```tsx
import type { AppRouteObject } from '@/core/router/types';
import { createLazyRoute } from '@/core/router';

export const myModuleRoutes: AppRouteObject[] = [
  {
    name: 'my-module',
    path: 'my-module',
    meta: {
      title: 'routes:myModule',     // 翻译键，routes 命名空间
      icon: 'lucide:some-icon',     // Iconify 图标
      order: 10,
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

### 3. 新增 API

**步骤**: 生成代码 → 创建 Service → 创建 Hooks → 导出

Service 文件 `src/api/service/my-module.ts`：
```typescript
import { createMyModuleServiceClient } from '@/api/generated/admin/service/v1';
import { type PaginationQuery, requestApi } from '@/core';

let _instance: ReturnType<typeof createMyModuleServiceClient> | null = null;
export function getMyModuleService() {
  if (!_instance) _instance = createMyModuleServiceClient(requestApi);
  return _instance;
}
export async function listMyModules(query: PaginationQuery) {
  return getMyModuleService().List(query.toRawParams());
}
```

Hooks 文件 `src/api/hooks/my-module.ts`：
```typescript
import { useMutation, type UseMutationOptions } from '@tanstack/react-query';
import { listMyModules } from '@/api/service/my-module';
import { type PaginationQuery, queryClient } from '@/core';

export function useListMyModules(options?: UseMutationOptions<...>) {
  return useMutation({ mutationFn: (q) => listMyModules(q), ...options });
}
export async function fetchListMyModules(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ['listMyModules', params], queryFn: () => listMyModules(params), retry: 0,
  });
}
```

然后在 `src/api/service/index.ts` 和 `src/api/hooks/index.ts` 中添加导出。

### 4. 新增国际化

在 `src/locales/zh-CN/_modules/` 和 `src/locales/en-US/_modules/` 创建同名 JSON：

```json
// zh-CN/_modules/my-module.json
{ "pageTitle": "我的模块", "name": "名称" }

// en-US/_modules/my-module.json
{ "pageTitle": "My Module", "name": "Name" }
```

组件中使用：
```tsx
const { t } = useI18n('my-module');
<h1>{t('pageTitle')}</h1>
```

## 权限系统

### 数据来源

- 角色码: `useUserStore.userRoles`（来自 userInfo.roles）
- 权限码: `useUserStore.accessCodes`（来自 GetMyPermissionCode）
- `meta.authority`: 角色码和权限码的**混合数组**

### 三种鉴权方式

```tsx
// 1. Hook（推荐）
const { hasAccessByCodes, hasAccessByRoles } = useAccess();
{hasAccessByCodes(['sys:user:create']) && <Button>新建</Button>}

// 2. 组件
<AccessControl codes={['sys:user:create']} type="code">
  <Button>新建</Button>
</AccessControl>

// 3. 非组件场景
const { hasAccessByCodes } = getAccessStatic();
```

### 超级管理员

拥有 `*:*:*` 角色的用户自动通过所有权限检查。

## 状态管理

### Store 列表

| Store | 文件 | 用途 |
|-------|------|------|
| `useAuthStore` | `stores/auth.ts` | Token、登录/登出、注册 |
| `useUserStore` | `stores/user.ts` | 用户信息、角色码、权限码 |
| `usePreferencesStore` | `core/preferences/store/` | 偏好设置 |

### 使用方式

```typescript
// React 组件中
const token = useAuthStore((s) => s.accessToken);

// 非组件环境中
const token = useAuthStore.getState().accessToken;
```

## 路由系统

### 路由类型

| 路由 | 位置 | 说明 |
|------|------|------|
| 静态路由 | `router/config/static.ts` | 主布局 + 根路由 |
| 认证路由 | `router/config/auth.ts` | 登录/注册页 |
| 错误路由 | `router/config/error-routes.ts` | 403/404/500 |
| 业务路由 | `router/modules/*.tsx` | 自动导入 |

### 权限模式

- `frontend`: 前端路由 + `meta.authority` 过滤（默认）
- `backend`: 后端返回菜单 + `pageMap` 动态匹配

### Route Meta 关键字段

```typescript
meta: {
  title: 'routes:xxx',      // i18n 翻译键
  icon: 'lucide:xxx',       // Iconify 图标
  order: 10,                // 排序
  authority: ['code1'],     // 权限码/角色码
  hideInMenu: false,        // 是否隐藏菜单
  hideInTab: false,         // 是否隐藏标签页
  keepAlive: false,         // 是否缓存
}
```

## 关键注意事项

1. **PaginationQuery 必须用 new 实例化**: `new PaginationQuery({ page, pageSize })`
2. **非组件环境用 fetchXxx / service**: 不要在非 React 环境使用 useXxx Hook
3. **国际化插值用 `{{var}}`**: 不是 `#{var}` 或 `${var}`
4. **路由标题用 `routes:` 前缀**: `meta.title` 使用 `'routes:xxx'` 格式
5. **权限码格式**: `模块:资源:操作`，如 `sys:user:create`
6. **不要手动修改 generated 目录**: API 类型由工具自动生成
7. **DrawerForm 用 formRef**: 没有 useForm 方法
8. **antd v6**: 使用 `items` 替代 `TabPane`，`title` 替代 Alert 的 `message`
9. **ProTable scroll.y**: 初始值必须是像素值（数字），不能是百分比
10. **Store 持久化**: auth-store 持久化 token，user-store 持久化 userInfo，偏好设置持久化到 app-preferences

## 代码风格

- Prettier: 单引号、分号、尾逗号、100 字符行宽、2 空格缩进
- 路径别名: `@/` → `src/`，`#/` → `types/`
- 命名: Service 层 `listXxx/getXxx`，Hooks 层 `useListXxx/fetchListXxx`
- 提交: Husky + commitlint 约定式提交
