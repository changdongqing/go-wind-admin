import {GlobalOutlined, MoonOutlined, SunOutlined} from '@ant-design/icons';
import {Helmet, useIntl, useModel} from '@umijs/max';
import {Button, Tooltip} from 'antd';
import React from 'react';

import SloganIcon from './icons/SloganIcon';
import Settings from '../../../config/defaultSettings';
import './AuthLayout.style.less';
import './AuthLayout.style.less';

/**
 * 认证页面布局属性
 */
export interface AuthLayoutProps {
  /** 页面标题（如：欢迎回来、创建账号、找回密码） */
  title: string;
  /** 页面副标题描述 */
  description: string;
  /** 表单内容（由子页面传入） */
  children: React.ReactNode;
  /** 页面标识（用于 Helmet title） */
  pageKey?: string;
  /** 底部链接区域 */
  footerLink?: {
    text: string;
    linkText: string;
    href: string;
  };
}

/**
 * 认证页面通用布局组件
 * 用于登录、注册、找回密码等页面
 */
const AuthLayout: React.FC<AuthLayoutProps> = ({
                                                 title,
                                                 description,
                                                 children,
                                                 pageKey = 'auth',
                                                 footerLink,
                                               }) => {
  const intl = useIntl();
  const {mode: themeMode, setMode: setThemeMode} = useModel('core.theme');
  const {locale: currentLocale, setLocale: setCurrentLocale} = useModel('core.language');

  // 切换主题
  const toggleTheme = () => {
    const newMode = themeMode === 'light' ? 'dark' : 'light';
    setThemeMode(newMode);
  };

  // 根据主题模式判断当前是否为亮色模式
  const isLightMode = React.useMemo(() => {
    if (themeMode === 'system') {
      return window.matchMedia('(prefers-color-scheme: light)').matches;
    }
    return themeMode === 'light';
  }, [themeMode]);

  // 监听系统主题变化
  React.useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handleChange = () => {
      if (themeMode === 'system') {
        // 强制重新渲染以更新 isLightMode
        setThemeMode('system');
      }
    };
    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, [themeMode]);

  // 切换语言
  const toggleLanguage = () => {
    setCurrentLocale(currentLocale === 'zh-CN' ? 'en-US' : 'zh-CN');
  };

  return (
    <div className={`auth-layout-wrapper${isLightMode ? ' light-mode' : ''}`}>
      {/* 右上角工具栏 */}
      <div className="auth-toolbar">
        <Tooltip title={currentLocale === 'zh-CN' ? '切换语言' : 'Switch Language'}>
          <Button
            type="text"
            icon={<GlobalOutlined/>}
            onClick={toggleLanguage}
            className={isLightMode ? 'auth-toolbar-btn-light' : 'auth-toolbar-btn-dark'}
          >
            {currentLocale === 'zh-CN' ? 'EN' : '中文'}
          </Button>
        </Tooltip>
        <Tooltip title={themeMode === 'light' ? '切换暗黑模式' : '切换亮色模式'}>
          <Button
            type="text"
            icon={themeMode === 'light' ? <MoonOutlined/> : <SunOutlined/>}
            onClick={toggleTheme}
            className={isLightMode ? 'auth-toolbar-btn-light' : 'auth-toolbar-btn-dark'}
          />
        </Tooltip>
      </div>

      <Helmet>
        <title>
          {intl.formatMessage({
            id: `menu.${pageKey}`,
            defaultMessage: title,
          })}
          {Settings.title && ` - ${Settings.title}`}
        </title>
      </Helmet>

      {/* 左侧品牌展示区 */}
      <div className="auth-brand-section">
        {/* 背景装饰 - 多层渐变 */}
        <div className="auth-brand-overlay" />

        {/* 装饰圆形 */}
        <div className="auth-brand-circle circle-1" />
        <div className="auth-brand-circle circle-2" />

        {/* 品牌图标 */}
        <div className="auth-brand-icon">
          <SloganIcon />
        </div>

        <h2 className="auth-brand-title">
          风行中后台管理系统
        </h2>
        <p className="auth-brand-description">
          开箱即用的企业级中后台管理系统
        </p>
      </div>

      {/* 右侧表单区 */}
      <div className="auth-form-section">
        <div className="auth-form-content">
          {/* 页面标题 */}
          <h1 className="auth-form-title">
            {title}
          </h1>

          {/* 页面描述 */}
          <p className="auth-form-description">
            {description}
          </p>

          {/* 表单内容（由子页面传入） */}
          {children}

          {/* 底部链接 */}
          {footerLink && (
            <div className="auth-footer-link">
              <span className="auth-footer-text">
                {footerLink.text}{' '}
              </span>
              <a
                href={footerLink.href}
                className="auth-footer-anchor"
              >
                {footerLink.linkText}
              </a>
            </div>
          )}
        </div>

        {/* 底部版权信息 */}
        <div className="auth-copyright">
          Copyright © {new Date().getFullYear()} GoWind
        </div>
      </div>
    </div>
  );
};

export default AuthLayout;
