import React from 'react';
import {Button, Input, InputNumber, Segmented, Select, Space, Switch} from 'antd';
import {MinusOutlined, PlusOutlined, QuestionCircleOutlined} from '@ant-design/icons';
import {usePreferencesStore} from '../../store';
import type {ContentCompactType, LayoutType} from '../../types';
import './LayoutPanel.style.less';

/** 布局选项 */
const LAYOUT_OPTIONS = [
  {
    label: '垂直',
    value: 'sidebar-nav',
    icon: '📋',
  },
  {
    label: '双列菜单',
    value: 'sidebar-mixed-nav',
    icon: '📑',
  },
  {
    label: '水平',
    value: 'header-nav',
    icon: '',
  },
  {
    label: '混合菜单',
    value: 'mixed-nav',
    icon: '📊',
  },
  {
    label: '内容全屏',
    value: 'full-content',
    icon: '📱',
  },
];

/** 内容宽度选项 */
const CONTENT_COMPACT_OPTIONS = [
  { label: '流式', value: 'wide' },
  { label: '定宽', value: 'compact' },
];

export const LayoutPanel: React.FC = () => {
  const { preferences, setPreferences } = usePreferencesStore();

  const handleLayoutChange = (layout: LayoutType) => {
    setPreferences({ app: { layout } });
  };

  const handleContentCompactChange = (compact: ContentCompactType) => {
    setPreferences({ app: { contentCompact: compact } });
  };

  const handleSidebarWidthChange = (width: number | null) => {
    if (width && width >= 180 && width <= 320) {
      setPreferences({ sidebar: { width } });
    }
  };

  return (
    <div className="layout-panel">
      {/* 布局选择 */}
      <section className="layout-section">
        <h3 className="section-title">布局</h3>
        <div className="layout-grid">
          {LAYOUT_OPTIONS.map((option) => (
            <div
              key={option.value}
              className={`layout-item ${preferences.app.layout === option.value ? 'active' : ''}`}
              onClick={() => handleLayoutChange(option.value as LayoutType)}
            >
              <div className="layout-preview">
                {/* 渲染真实的布局示意图 */}
                {option.value === 'sidebar-nav' && (
                  <div className="layout-visual sidebar-nav">
                    <div className="sidebar" />
                    <div className="main-content">
                      <div className="header-bar gray" />
                      <div className="content-area">
                        <div className="block-row">
                          <div className="block" />
                          <div className="block" />
                        </div>
                        <div className="block" />
                      </div>
                    </div>
                  </div>
                )}
                {option.value === 'sidebar-mixed-nav' && (
                  <div className="layout-visual sidebar-mixed-nav">
                    <div className="sidebar thin primary" />
                    <div className="sidebar thin gray" />
                    <div className="main-content">
                      <div className="header-bar gray" />
                      <div className="content-area">
                        <div className="block-row">
                          <div className="block" />
                          <div className="block" />
                        </div>
                        <div className="block" />
                      </div>
                    </div>
                  </div>
                )}
                {option.value === 'header-nav' && (
                  <div className="layout-visual header-nav">
                    <div className="content-wrapper">
                      <div className="header-bar full-width">
                        <div className="menu-item" />
                        <div className="menu-item" />
                        <div className="menu-item" />
                        <div className="menu-item" />
                        <div className="menu-item" />
                      </div>
                      <div className="content-area">
                        <div className="block-row">
                          <div className="block" />
                          <div className="block" />
                        </div>
                        <div className="block" />
                      </div>
                    </div>
                  </div>
                )}
                {option.value === 'mixed-nav' && (
                  <div className="layout-visual mixed-nav">
                    <div className="content-wrapper">
                      <div className="header-bar full-width">
                        <div className="menu-item" />
                        <div className="menu-item" />
                        <div className="menu-item" />
                      </div>
                      <div className="main-content">
                        <div className="sidebar thin gray" />
                        <div className="content-area">
                          <div className="block-row">
                            <div className="block" />
                            <div className="block" />
                          </div>
                          <div className="block" />
                        </div>
                      </div>
                    </div>
                  </div>
                )}
                {option.value === 'full-content' && (
                  <div className="layout-visual full-content">
                    <div className="content-area full-width">
                      <div className="block-row">
                        <div className="block" />
                        <div className="block" />
                      </div>
                      <div className="block" />
                    </div>
                  </div>
                )}
              </div>
              <div className="layout-label">
                <span>{option.label}</span>
                <QuestionCircleOutlined className="help-icon" />
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* 内容宽度 */}
      <section className="layout-section">
        <h3 className="section-title">内容</h3>
        <div className="content-compact-grid">
          {CONTENT_COMPACT_OPTIONS.map((option) => (
            <div
              key={option.value}
              className={`content-compact-item ${preferences.app.contentCompact === option.value ? 'active' : ''}`}
              onClick={() => handleContentCompactChange(option.value as ContentCompactType)}
            >
              <div className="content-preview">
                <div className={`preview-bar ${option.value === 'compact' ? 'narrow' : 'wide'}`} />
              </div>
              <span>{option.label}</span>
            </div>
          ))}
        </div>
      </section>

      {/* 侧边栏设置 */}
      <section className="layout-section">
        <h3 className="section-title">侧边栏</h3>
        <div className="preference-item">
          <span>显示侧边栏</span>
          <Switch
            checked={preferences.sidebar.enable}
            onChange={(checked) => setPreferences({ sidebar: { enable: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>折叠菜单</span>
          <Switch
            checked={preferences.sidebar.collapsed}
            onChange={(checked) => setPreferences({ sidebar: { collapsed: checked } })}
          />
        </div>
        <div className={`preference-item ${preferences.sidebar.collapsed ? 'disabled' : ''}`}>
          <span>折叠显示菜单名</span>
          <Switch
            disabled={!preferences.sidebar.collapsed}
            checked={preferences.sidebar.collapsedShowTitle}
            onChange={(checked) => setPreferences({ sidebar: { collapsedShowTitle: checked } })}
          />
        </div>
        <div className="preference-item width-control">
          <span>宽度</span>
          <Space size={4}>
            <Button
              size="small"
              icon={<MinusOutlined/>}
              onClick={() => handleSidebarWidthChange(preferences.sidebar.width - 8)}
              disabled={preferences.sidebar.width <= 180}
            />
            <InputNumber
              min={180}
              max={320}
              value={preferences.sidebar.width}
              onChange={handleSidebarWidthChange}
              style={{width: 70}}  // 从80压缩到70
            />
            <Button
              size="small"
              icon={<PlusOutlined/>}
              onClick={() => handleSidebarWidthChange(preferences.sidebar.width + 8)}
              disabled={preferences.sidebar.width >= 320}
            />
          </Space>
        </div>
      </section>

      {/* 顶栏设置 */}
      <section className="layout-section">
        <h3 className="section-title">顶栏</h3>
        <div className="preference-item">
          <span>显示顶栏</span>
          <Switch
            checked={preferences.header.enable}
            onChange={(checked) => setPreferences({ header: { enable: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>模式</span>
          <Segmented
            options={[
              { label: '固定', value: 'fixed' },
              { label: '自动', value: 'auto' },
              { label: '静态', value: 'static' },
            ]}
            value={preferences.header.mode}
            onChange={(value) => setPreferences({ header: { mode: value as any } })}
          />
        </div>
      </section>

      {/* 导航菜单 */}
      <section className="layout-section">
        <h3 className="section-title">导航菜单</h3>
        <div className="preference-item">
          <span>导航菜单风格</span>
          <Segmented
            options={[
              { label: '圆润', value: 'rounded' },
              { label: '朴素', value: 'plain' },
            ]}
            value={preferences.navigation.styleType}
            onChange={(value) => setPreferences({ navigation: { styleType: value as any } })}
          />
        </div>
        <div className="preference-item">
          <span>导航菜单分离</span>
          <Switch
            checked={preferences.navigation.split}
            onChange={(checked) => setPreferences({ navigation: { split: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>侧边导航菜单手风琴模式</span>
          <Switch
            checked={preferences.navigation.accordion}
            onChange={(checked) => setPreferences({ navigation: { accordion: checked } })}
          />
        </div>
      </section>

      {/* 面包屑导航 */}
      <section className="layout-section">
        <h3 className="section-title">面包屑导航</h3>
        <div className="preference-item">
          <span>开启面包屑导航</span>
          <Switch
            checked={preferences.breadcrumb.enable}
            onChange={(checked) => setPreferences({ breadcrumb: { enable: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>仅有一个时隐藏</span>
          <Switch
            checked={preferences.breadcrumb.hideOnlyOne}
            onChange={(checked) => setPreferences({ breadcrumb: { hideOnlyOne: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>显示面包屑图标</span>
          <Switch
            checked={preferences.breadcrumb.showIcon}
            onChange={(checked) => setPreferences({ breadcrumb: { showIcon: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>显示首页按钮</span>
          <Switch
            checked={preferences.breadcrumb.showHome}
            onChange={(checked) => setPreferences({ breadcrumb: { showHome: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>面包屑风格</span>
          <Segmented
            options={[
              { label: '常规', value: 'normal' },
              { label: '背景', value: 'background' },
            ]}
            value={preferences.breadcrumb.styleType}
            onChange={(value) => setPreferences({ breadcrumb: { styleType: value as any } })}
          />
        </div>
      </section>

      {/* 标签栏 */}
      <section className="layout-section">
        <h3 className="section-title">标签栏</h3>
        <div className="preference-item">
          <span>启用标签栏</span>
          <Switch
            checked={preferences.tabbar.enable}
            onChange={(checked) => setPreferences({ tabbar: { enable: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>持久化标签页</span>
          <Switch
            checked={preferences.tabbar.persist}
            onChange={(checked) => setPreferences({ tabbar: { persist: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>启动拖拽排序</span>
          <Switch
            checked={preferences.tabbar.draggable}
            onChange={(checked) => setPreferences({ tabbar: { draggable: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>显示标签栏图标</span>
          <Switch
            checked={preferences.tabbar.showIcon}
            onChange={(checked) => setPreferences({ tabbar: { showIcon: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>显示更多按钮</span>
          <Switch
            checked={preferences.tabbar.showMore}
            onChange={(checked) => setPreferences({ tabbar: { showMore: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>显示最大化按钮</span>
          <Switch
            checked={preferences.tabbar.showMaximize}
            onChange={(checked) => setPreferences({ tabbar: { showMaximize: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>标签页风格</span>
          <Segmented
            options={[
              { label: '谷歌', value: 'chrome' },
              { label: '卡片', value: 'card' },
              { label: '极简', value: 'minimal' },
            ]}
            value={preferences.tabbar.styleType}
            onChange={(value) => setPreferences({ tabbar: { styleType: value as any } })}
          />
        </div>
      </section>

      {/* 小部件 */}
      <section className="layout-section">
        <h3 className="section-title">小部件</h3>
        <div className="preference-item">
          <span>启用全局搜索</span>
          <Switch
            checked={preferences.widget.globalSearch}
            onChange={(checked) => setPreferences({ widget: { globalSearch: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>启用主题切换</span>
          <Switch
            checked={preferences.widget.themeToggle}
            onChange={(checked) => setPreferences({ widget: { themeToggle: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>启用语言切换</span>
          <Switch
            checked={preferences.widget.languageToggle}
            onChange={(checked) => setPreferences({ widget: { languageToggle: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>启用全屏</span>
          <Switch
            checked={preferences.widget.fullscreen}
            onChange={(checked) => setPreferences({ widget: { fullscreen: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>启用通知</span>
          <Switch
            checked={preferences.widget.notification}
            onChange={(checked) => setPreferences({ widget: { notification: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>启用锁屏</span>
          <Switch
            checked={preferences.widget.lockScreen}
            onChange={(checked) => setPreferences({ widget: { lockScreen: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>启用侧边栏切换</span>
          <Switch
            checked={preferences.widget.sidebarToggle}
            onChange={(checked) => setPreferences({ widget: { sidebarToggle: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>启用刷新</span>
          <Switch
            checked={preferences.widget.refresh}
            onChange={(checked) => setPreferences({ widget: { refresh: checked } })}
          />
        </div>
        <div className="preference-item">
          <span>偏好设置位置</span>
          <Select
            style={{width: 120}}
            options={[
              { label: '自动', value: 'auto' },
              { label: '固定', value: 'fixed' },
              { label: '隐藏', value: 'hidden' },
            ]}
            value={preferences.app.preferencesButtonPosition}
            onChange={(value) => setPreferences({ app: { preferencesButtonPosition: value as any } })}
          />
        </div>
      </section>

      {/* 底栏 */}
      <section className="layout-section">
        <h3 className="section-title">底栏</h3>
        <div className="preference-item">
          <span>显示底栏</span>
          <Switch
            checked={preferences.footer.enable}
            onChange={(checked) => setPreferences({ footer: { enable: checked } })}
          />
        </div>
        <div className={`preference-item ${!preferences.footer.enable ? 'disabled' : ''}`}>
          <span>固定在底部</span>
          <Switch
            disabled={!preferences.footer.enable}
            checked={preferences.footer.fixed}
            onChange={(checked) => setPreferences({ footer: { fixed: checked } })}
          />
        </div>
      </section>

      {/* 版权 */}
      <section className="layout-section">
        <h3 className="section-title">版权</h3>
        <div className="preference-item">
          <span>启用版权</span>
          <Switch
            checked={preferences.copyright.enable}
            onChange={(checked) => setPreferences({ copyright: { enable: checked } })}
          />
        </div>
        <div className={`preference-item ${!preferences.copyright.enable ? 'disabled' : ''}`}>
          <span>公司名</span>
          <Input
            style={{width: 200}}
            value={preferences.copyright.companyName}
            onChange={(e) => setPreferences({ copyright: { companyName: e.target.value } })}
            disabled={!preferences.copyright.enable}
          />
        </div>
        <div className={`preference-item ${!preferences.copyright.enable ? 'disabled' : ''}`}>
          <span>公司主页</span>
          <Input
            style={{width: 200}}
            value={preferences.copyright.companySiteLink}
            onChange={(e) => setPreferences({ copyright: { companySiteLink: e.target.value } })}
            disabled={!preferences.copyright.enable}
          />
        </div>
        <div className={`preference-item ${!preferences.copyright.enable ? 'disabled' : ''}`}>
          <span>日期</span>
          <Input
            style={{width: 200}}
            value={preferences.copyright.date}
            onChange={(e) => setPreferences({ copyright: { date: e.target.value } })}
            disabled={!preferences.copyright.enable}
          />
        </div>
        <div className={`preference-item ${!preferences.copyright.enable ? 'disabled' : ''}`}>
          <span>ICP备案号</span>
          <Input
            style={{width: 200}}
            value={preferences.copyright.icp}
            onChange={(e) => setPreferences({ copyright: { icp: e.target.value } })}
            disabled={!preferences.copyright.enable}
          />
        </div>
        <div className={`preference-item ${!preferences.copyright.enable ? 'disabled' : ''}`}>
          <span>ICP网站链接</span>
          <Input
            style={{width: 200}}
            value={preferences.copyright.icpLink}
            onChange={(e) => setPreferences({ copyright: { icpLink: e.target.value } })}
            disabled={!preferences.copyright.enable}
          />
        </div>
      </section>
    </div>
  );
};
