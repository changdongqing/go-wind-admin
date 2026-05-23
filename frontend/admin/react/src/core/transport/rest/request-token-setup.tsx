/**
 * Request Token 配置
 *
 * 用于在应用初始化时设置 token 获取方法和错误消息处理
 */

import React, {useEffect, useRef} from 'react';
import { App } from 'antd';

import {setGetTokenCallback, setErrorHandlerCallback} from './rest-client';
import {useAuthStore} from '@/stores/auth';

/**
 * RequestTokenSetup 组件
 * 在应用初始化时设置 token 获取回调和错误消息处理
 */
export function RequestTokenSetup({children}: { children: React.ReactNode }) {
  const initialized = useRef(false);
  const { message } = App.useApp();

  useEffect(() => {
    // StrictMode 下 useEffect 会执行两次，确保只初始化一次
    if (initialized.current) return;
    initialized.current = true;

    // 设置获取 Token 的回调函数
    setGetTokenCallback(() => {
      const token = useAuthStore.getState().accessToken;
      console.log('🔑 Getting access token from store:', {
        hasToken: !!token,
        tokenPreview: token ? '***' + token.slice(-8) : 'null',
        fullTokenLength: token?.length || 0,
      });
      return token || null;
    });

    // 设置错误消息处理的回调函数
    setErrorHandlerCallback((msg: string) => {
      message.error({
        content: msg,
        key: msg, // 去重：相同的消息只显示一次
      });
    });

    console.log('Request client initialized with access token getter and error handler');
  }, [message]);

  return <>{children}</>;
}
