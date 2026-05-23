import { useState, useEffect, useMemo } from 'react';
import { RouterProvider } from 'react-router-dom';

import { createAccessibleRouter } from '@/core/router/factory';
import { staticRoutes } from './config/static';
import { useAuthStore, useUserStore } from '@/stores';

import { Forbidden } from '@/pages/core/error';

export const AppRouter = () => {
  const [router, setRouter] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  const { accessToken } = useAuthStore();
  const { userInfo } = useUserStore();

  // 计算属性：是否已认证、权限列表（使用 useMemo 稳定化）
  const isAuthenticated = !!accessToken;
  const permissions = useMemo(() => userInfo?.permissions || [], [userInfo?.permissions]);

  useEffect(() => {
    const initRouter = async () => {
      setLoading(true);

      try {
        // 生成完整路由（包含 AuthGuard 和 GuestGuard）
        const appRouter = await createAccessibleRouter({
          routes: staticRoutes,
          permissions,
          forbiddenElement: <Forbidden />,
          autoInjectRedirect: true,
          autoSort: true,
        });
        setRouter(appRouter);
      } catch (err) {
        console.error('Router init failed:', err);
      } finally {
        setLoading(false);
      }
    };

    initRouter();
  }, [isAuthenticated, permissions]);

  if (loading || !router)
    return <div className="h-screen flex items-center justify-center">初始化中...</div>;

  return <RouterProvider router={router} />;
};
