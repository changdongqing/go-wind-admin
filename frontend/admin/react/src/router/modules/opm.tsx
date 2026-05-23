import type { AppRouteObject } from '@/core/router/types';
import { createLazyRoute } from '@/router/utils/lazy';

/**
 * 组织人员管理路由配置
 * 包括组织架构、职位管理、用户管理等页面
 */
export const opmRoutes: AppRouteObject[] = [
  {
    name: 'opm',
    path: 'opm', // 相对路径，会自动拼接到父路由 '/'
    meta: {
      title: '组织人员管理',
      icon: 'lucide:users', // Iconify 格式
      order: 2001,
      keepAlive: true, // 保持组件状态
      permission: 'sys:platform_admin', // 平台管理员或租户管理员权限
    },
    children: [
      {
        name: 'org-units',
        path: 'org-units', // 相对路径，最终为 /opm/org-units
        element: createLazyRoute(() => import('@/pages/app/opm/org-unit')),
        meta: {
          title: '组织架构',
          icon: 'lucide:layers', // Iconify 格式
          order: 1,
          permission: 'sys:platform_admin', // 平台管理员或租户管理员权限
        },
      },
      {
        name: 'positions',
        path: 'positions', // 相对路径，最终为 /opm/positions
        element: createLazyRoute(() => import('@/pages/app/opm/position')),
        meta: {
          title: '职位管理',
          icon: 'lucide:briefcase', // Iconify 格式
          order: 2,
          permission: 'sys:platform_admin', // 平台管理员或租户管理员权限
        },
      },
      {
        name: 'users',
        path: 'users', // 相对路径，最终为 /opm/users
        element: createLazyRoute(() => import('@/pages/app/opm/user')),
        meta: {
          title: '用户管理',
          icon: 'lucide:user', // Iconify 格式
          order: 3,
          permission: 'sys:platform_admin', // 平台管理员或租户管理员权限
        },
      },
      {
        name: 'user-detail',
        path: 'users/detail/:id', // 动态路由，最终为 /opm/users/detail/:id
        element: createLazyRoute(() => import('@/pages/app/opm/user/detail')),
        meta: {
          title: '用户详情',
          hideInMenu: true, // 隐藏在菜单中，仅通过编程导航访问
          permission: 'sys:platform_admin', // 平台管理员或租户管理员权限
        },
      },
    ],
  },
];

export default opmRoutes;
