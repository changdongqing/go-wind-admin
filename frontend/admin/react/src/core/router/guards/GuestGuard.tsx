import { Navigate } from 'react-router-dom';
import React from 'react';
import { useAuthStore } from '@/stores';

interface GuestGuardProps {
  isAuthenticated?: boolean;
  children: React.ReactNode;
  redirectPath?: string;
}

/**
 * 访客守卫：已登录用户不能访问（如登录页、注册页）
 * 如果已登录，重定向到指定路径
 */
export const GuestGuard = ({
  isAuthenticated: isAuthenticatedProp,
  children,
  redirectPath = '/',
}: GuestGuardProps) => {
  // 若未显式传入 isAuthenticated，则从 store 读取
  const isAuthenticated = isAuthenticatedProp ?? !!useAuthStore.getState().accessToken;

  if (isAuthenticated) {
    return <Navigate to={redirectPath} replace />;
  }

  return <>{children}</>;
};
