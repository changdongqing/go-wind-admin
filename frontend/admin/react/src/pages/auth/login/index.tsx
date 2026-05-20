import {LockOutlined, UserOutlined} from '@ant-design/icons';
import {LoginForm, ProFormCheckbox, ProFormText} from '@ant-design/pro-components';
import {Helmet, useIntl, useModel} from '@umijs/max';
import {App} from 'antd';
import React from 'react';

import Settings from '../../../../config/defaultSettings';

const Login: React.FC = () => {
  const {message} = App.useApp();
  const intl = useIntl();
  const {login, loginLoading} = useModel('business.authentication');
  const access = useModel('auth.access');

  const handleSubmit = async (values: { username: string; password: string }) => {
    console.log('[Login] Form submitted with values:', values);

    try {
      console.log('[Login] Calling login function...');
      const result = await login(
        {
          username: values.username,
          password: values.password,
          grant_type: 'password',
        },
      );

      console.log('[Login] Login function returned result:', result);

      // 保存令牌到 AccessModel
      if (result.accessToken || result.refreshToken) {
        console.log('[Login] Saving tokens to AccessModel');
        access.setTokens({
          accessToken: result.accessToken ? {
            value: result.accessToken,
            expiresAt: Date.now() + 7200 * 1000, // 默认 2 小时过期
          } : null,
          refreshToken: result.refreshToken ? {
            value: result.refreshToken,
            expiresAt: Date.now() + 30 * 24 * 60 * 60 * 1000, // 默认 30 天过期
          } : null,
        });
        console.log('[Login] Tokens saved to AccessModel');
      } else {
        console.warn('[Login] No tokens in result');
      }

      message.success(intl.formatMessage({id: 'pages.login.success'}));

      // 等待一小段时间确保 localStorage 写入完成，然后跳转
      setTimeout(() => {
        const urlParams = new URL(window.location.href).searchParams;
        const redirect = urlParams.get('redirect') || '/';
        console.log('[Login] Redirecting to:', redirect);
        window.location.href = redirect;
      }, 300);
    } catch (error: any) {
      console.error('[Login] Error occurred:', error);
      // 错误已在 model 中处理
      message.error(error?.message || intl.formatMessage({id: 'pages.login.failure'}));
    }
  };

  return (
    <div
      style={{
        display: 'flex',
        minHeight: '100vh',
        background: '#0a0a0a',
      }}
    >
      <Helmet>
        <title>
          {intl.formatMessage({
            id: 'menu.login',
            defaultMessage: '登录页',
          })}
          {Settings.title && ` - ${Settings.title}`}
        </title>
      </Helmet>
      
      {/* 左侧品牌展示区 */}
      <div
        style={{
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'radial-gradient(ellipse at center, rgba(59, 130, 246, 0.15) 0%, transparent 70%)',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        {/* 背景渐变装饰 */}
        <div
          style={{
            position: 'absolute',
            top: '-50%',
            left: '-50%',
            width: '200%',
            height: '200%',
            background: 'radial-gradient(circle at 30% 50%, rgba(59, 130, 246, 0.1) 0%, transparent 50%)',
            pointerEvents: 'none',
          }}
        />
        
        {/* 品牌图标 */}
        <img 
          alt="logo" 
          src="/logo.svg" 
          style={{
            width: 280,
            height: 280,
            marginBottom: 32,
            filter: 'drop-shadow(0 0 40px rgba(59, 130, 246, 0.3))',
          }}
        />
        
        <h2
          style={{
            color: '#fff',
            fontSize: 24,
            fontWeight: 600,
            marginBottom: 12,
            textAlign: 'center',
          }}
        >
          风行中后台管理系统
        </h2>
        <p
          style={{
            color: 'rgba(255, 255, 255, 0.6)',
            fontSize: 14,
            textAlign: 'center',
          }}
        >
          开箱即用的企业级中后台管理系统
        </p>
      </div>
      
      {/* 右侧登录表单区 */}
      <div
        style={{
          flex: 2,
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'center',
          padding: '64px 48px',
          background: '#141414',
          borderLeft: '1px solid rgba(255, 255, 255, 0.08)',
          position: 'relative',
        }}
      >
        <div style={{maxWidth: 360, margin: '0 auto', width: '100%'}}>
          <h1
            style={{
              color: '#fff',
              fontSize: 28,
              fontWeight: 600,
              marginBottom: 8,
            }}
          >
            欢迎回来 👋
          </h1>
          <p
            style={{
              color: 'rgba(255, 255, 255, 0.5)',
              fontSize: 14,
              marginBottom: 32,
            }}
          >
            请输入您的帐户信息以开始管理您的系统
          </p>
          
          <LoginForm
            loading={loginLoading}
            logo={false}
            title={false}
            subTitle={false}
            initialValues={{
              autoLogin: true,
            }}
            onFinish={handleSubmit}
            submitter={{
              searchConfig: {
                submitText: '登录',
              },
              submitButtonProps: {
                size: 'large',
                style: {
                  width: '100%',
                  background: 'linear-gradient(135deg, #0066ff 0%, #0052cc 100%)',
                  border: 'none',
                  height: 40,
                  fontSize: 15,
                  fontWeight: 500,
                },
              },
            }}
          >
            <ProFormText
              name="username"
              fieldProps={{
                size: 'large',
                placeholder: '请输入用户名',
                className: 'login-input',
              }}
              rules={[
                {
                  required: true,
                  message: '请输入用户名',
                },
              ]}
            />
            <ProFormText.Password
              name="password"
              fieldProps={{
                size: 'large',
                placeholder: '密码',
                className: 'login-input',
              }}
              rules={[
                {
                  required: true,
                  message: '请输入密码',
                },
              ]}
            />
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                marginBottom: 24,
              }}
            >
              <ProFormCheckbox
                name="autoLogin"
                fieldProps={{
                  style: {
                    color: 'rgba(255, 255, 255, 0.6)',
                    fontSize: 13,
                  },
                }}
              >
                记住账号
              </ProFormCheckbox>
            </div>
          </LoginForm>
          
          <div
            style={{
              textAlign: 'center',
              marginTop: 16,
            }}
          >
            <span style={{color: 'rgba(255, 255, 255, 0.5)', fontSize: 13}}>
              还没有账号？{' '}
            </span>
            <a
              href="/auth/register"
              style={{
                color: '#0066ff',
                fontSize: 13,
                textDecoration: 'none',
              }}
            >
              创建账号
            </a>
          </div>
        </div>
        
        {/* 底部版权信息 - 在右侧面板内 */}
        <div
          style={{
            position: 'absolute',
            bottom: 24,
            left: 0,
            right: 0,
            textAlign: 'center',
            color: 'rgba(255, 255, 255, 0.3)',
            fontSize: 12,
          }}
        >
          Copyright © {new Date().getFullYear()} GoWind
        </div>
      </div>
    </div>
  );
};
