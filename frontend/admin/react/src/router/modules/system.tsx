import type { AppRouteObject } from '@/core/router/types';
import { createLazyRoute } from '@/router/utils/lazy';

/**
 * 系统管理路由配置
 * 包括菜单管理、API管理、字典管理、文件管理、任务管理、登录策略、语言管理等页面
 */
export const systemRoutes: AppRouteObject[] = [
  {
    name: 'system',
    path: 'system', // 相对路径，会自动拼接到父路由 '/'
    meta: {
      title: '系统管理',
      icon: 'lucide:settings', // Iconify 格式
      order: 2005,
      keepAlive: true, // 保持组件状态
      permission: 'sys:platform_admin', // 平台管理员或租户管理员权限
    },
    children: [
      {
        name: 'menus',
        path: 'menus', // 相对路径，最终为 /system/menus
        element: createLazyRoute(() => import('@/pages/app/system/menu')),
        meta: {
          title: '菜单管理',
          icon: 'lucide:square-menu', // Iconify 格式
          order: 1,
          permission: 'sys:platform_admin', // 仅平台管理员权限
        },
      },
      {
        name: 'apis',
        path: 'apis', // 相对路径，最终为 /system/apis
        element: createLazyRoute(() => import('@/pages/app/system/api')),
        meta: {
          title: 'API管理',
          icon: 'lucide:route', // Iconify 格式
          order: 2,
          permission: 'sys:platform_admin', // 仅平台管理员权限
        },
      },
      {
        name: 'dict',
        path: 'dict', // 相对路径，最终为 /system/dict
        element: createLazyRoute(() => import('@/pages/app/system/dict')),
        meta: {
          title: '字典管理',
          icon: 'lucide:library-big', // Iconify 格式
          order: 3,
          permission: 'sys:platform_admin', // 仅平台管理员权限
        },
      },
      {
        name: 'files',
        path: 'files', // 相对路径，最终为 /system/files
        element: createLazyRoute(() => import('@/pages/app/system/file')),
        meta: {
          title: '文件管理',
          icon: 'lucide:file-search', // Iconify 格式
          order: 4,
          permission: 'sys:platform_admin', // 平台管理员或租户管理员权限
        },
      },
      {
        name: 'tasks',
        path: 'tasks', // 相对路径，最终为 /system/tasks
        element: createLazyRoute(() => import('@/pages/app/system/task')),
        meta: {
          title: '任务管理',
          icon: 'lucide:list-todo', // Iconify 格式
          order: 5,
          permission: 'sys:platform_admin', // 平台管理员或租户管理员权限
        },
      },
      {
        name: 'login-policies',
        path: 'login-policies', // 相对路径，最终为 /system/login-policies
        element: createLazyRoute(() => import('@/pages/app/system/login-policy')),
        meta: {
          title: '登录策略',
          icon: 'lucide:shield-x', // Iconify 格式
          order: 6,
          permission: 'sys:platform_admin', // 仅平台管理员权限
        },
      },
      {
        name: 'languages',
        path: 'languages', // 相对路径，最终为 /system/languages
        element: createLazyRoute(() => import('@/pages/app/system/language')),
        meta: {
          title: '语言管理',
          icon: 'lucide:globe', // Iconify 格式
          order: 7,
          permission: 'sys:platform_admin', // 仅平台管理员权限
        },
      },
    ],
  },
];

export default systemRoutes;
