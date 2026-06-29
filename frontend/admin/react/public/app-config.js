/**
 * GoWind Admin 运行时配置文件
 * ----------------------------------------------------------------------------
 * 本文件供【实施交付人员】在现场部署时修改。修改后刷新浏览器即生效，
 * 【不需要】重新编译打包，也【不需要】开发人员介入。
 *
 * 修改方法：用记事本（或任意文本编辑器，如 VS Code）打开本文件，
 *           修改下方花括号里对应的值即可，注意不要删除引号和逗号。
 *
 * 除本文件外，以下两项也可在部署目录直接【替换同名文件】免编译修改：
 *   - logo.png      站点 Logo 图片（左上角标题旁的小图标）
 *   - favicon.ico   浏览器页签图标（浏览器标签页上显示的小图标）
 *
 * 详细图文说明见：docs/set/deployment-config.md
 * ----------------------------------------------------------------------------
 */
window.__APP_CONFIG__ = {
  /**
   * 站点 / 应用名称。
   * 显示位置：浏览器页签标题、后台界面左上角标题。
   * 示例：'GoWind Admin'、'风行管理平台'、'XX集团后台' 等。
   */
  appName: "万物有名",
  /**
   * 登录页左侧品牌大标题与描述（可选）。
   * 留空或不配置时，使用代码内置的默认文案。
   */
  systemTitle: "万物有名管理平台",
  systemDescription: "企业级中后台管理系统",
  /**
   * 版权信息（显示在登录页底部及后台页脚，可选）。
   * 不需要的项留空字符串 "" 即可。
   */
  copyrightCompanyName: "万物有名科技有限公司",
  copyrightYear: "2026",
  copyrightSiteLink: "https://example.com",
  copyrightIcp: "",

  /**
   * 后端服务地址（可选）————————————————————————————————————————————
   * 桌面客户端 / 现场部署可在此【免编译】指定后端地址；留空（或不写）则
   * 使用打包时内置的默认地址（开发期见 .env.development，生产期见 .env.production）。
   *
   *   apiBaseUrl:      "https://api.customer-a.com",            // API 基础地址
   *   sseUrl:          "https://sse.customer-a.com/events",     // SSE 推送地址
   *   updateServerUrl: "https://updates.customer-a.com/desktop/", // 桌面端自动更新源（仅桌面端生效）
   */
};

// 尽早设置浏览器页签标题，避免页面加载过程中短暂显示旧的默认标题
document.title = window.__APP_CONFIG__.appName;
