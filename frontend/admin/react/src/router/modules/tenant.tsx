import type { AppRouteObject } from '@/core/router/types';
import { createLazyRoute } from '@/router/utils/lazy';

/**
 * 租户管理路由配置
 * 包括租户成员管理等页面
 */
export const tenantRoutes: AppRouteObject[] = [
  {
    name: 'tenant',
    path: 'tenant', // 相对路径，会自动拼接到父路由 '/'
    meta: {
      title: '租户管理',
      icon: 'lucide:building-2', // Iconify 格式
      order: 2000,
      permission: 'sys:platform_admin', // 平台管理员权限
    },
    children: [
      {
        name: 'tenant-members',
        path: 'members', // 相对路径，最终为 /tenant/members
        element: createLazyRoute(() => import('@/pages/app/tenant/tenant')),
        meta: {
          title: '租户成员',
          icon: 'lucide:users', // Iconify 格式
          order: 1,
          permission: 'sys:platform_admin', // 平台管理员权限
        },
      },
    ],
  },
];

export default tenantRoutes;
