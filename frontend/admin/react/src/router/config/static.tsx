import React from 'react';

import { type AppRouteObject, AuthGuard, GuestGuard } from '@/core/router';

import MainLayout from '@/layouts/MainLayout';
import UserLayout from '@/layouts/UserLayout';
import BlankLayout from '@/layouts/BlankLayout';
import RouteErrorFallback from '@/layouts/components/ErrorFallback/RouteErrorFallback.tsx';
import PageSkeleton from '@/layouts/components/LoadingSkeleton/presets/PageSkeleton';

import Login from '@/pages/core/auth/login';
import Register from '@/pages/core/auth/register';
import { Forbidden, NotFound } from '@/pages/core/error';

/**
 * 创建懒加载路由（自动包裹 Suspense + 骨架屏）
 */
const createLazyRoute = (loader: () => Promise<{ default: React.ComponentType<any> }>) => {
  const Component = React.lazy(loader);
  return (
    <Suspense fallback={<PageSkeleton delay={100} />}>
      <Component />
    </Suspense>
  );
};

/**
 * 静态基础路由配置
 * 规则：
 * 1. 公开路由（登录/404）放前面
 * 2. 需要登录的路由用 AuthGuard 包裹
 * 3. path: '*' 的 404 必须放在最后
 * 4. 懒加载组件用 createLazyRoute 包裹
 */
export const staticRoutes: AppRouteObject[] = [
  // ========== 公开路由（无需登录） ==========
  {
    name: 'auth',
    path: '/auth',
    element: <UserLayout requireAuth={false} />,
    errorElement: <RouteErrorFallback />,
    meta: { title: '认证', ignoreAccess: true, hideInMenu: true, hideInTab: true },
    children: [
      {
        name: 'login',
        path: 'login',
        element: (
          <GuestGuard>
            <Login />
          </GuestGuard>
        ),
        meta: { title: '登录', ignoreAccess: true },
      },
      {
        name: 'register',
        path: 'register',
        element: (
          <GuestGuard>
            <Register />
          </GuestGuard>
        ),
        meta: { title: '注册', ignoreAccess: true },
      },
    ],
  },

  // ========== 主应用路由（需登录 + 权限校验） ==========
  {
    name: 'app',
    path: '/',
    element: (
      <AuthGuard>
        <MainLayout />
      </AuthGuard>
    ),
    errorElement: <RouteErrorFallback />,
    meta: {
      title: '概览',
      icon: 'DashboardOutlined',
      requiresAuth: true,
      hideInMenu: false,
    },
    redirect: '/dashboard',
    children: [
      {
        name: 'dashboard',
        path: 'dashboard',
        // 懒加载 + 自动 Suspense
        element: createLazyRoute(() => import('@/pages/app/dashboard')),
        meta: {
          title: '仪表盘',
          icon: 'DashboardOutlined',
          order: 1,
          ignoreAccess: true, // 临时移除权限要求，用于调试
        },
      },
    ],
  },

  // ========== 错误页面（用 BlankLayout 包裹） ==========
  {
    name: 'error',
    path: '/error',
    element: <BlankLayout />,
    errorElement: <RouteErrorFallback />,
    meta: { title: '错误', hideInMenu: true, hideInTab: true },
    children: [
      {
        name: 'forbidden',
        path: '403',
        element: <Forbidden />,
        meta: { title: '无权限', ignoreAccess: true },
      },
    ],
  },

  // ========== 404 通配符路由（必须放在最后！） ==========
  {
    name: 'not-found',
    path: '*',
    element: <NotFound />,
    meta: { title: '404', ignoreAccess: true, hideInMenu: true, hideInTab: true },
  },
];
