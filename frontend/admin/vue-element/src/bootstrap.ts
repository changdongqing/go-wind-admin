import { createApp } from "vue";

// ===== 样式导入 =====
import "element-plus/dist/index.css";
import "element-plus/theme-chalk/dark/css-vars.css";
import "vxe-table/lib/style.css";
import "@/styles/index.scss";
import "uno.css";
import "animate.css";

import { setupDirective } from "@/directives";
import { setupI18n } from "@/i18n";
import { setupRouter } from "@/router";
import { initStores } from "@/stores/setup";
import { registerGlobComp } from "@/registerGlobComp";
import { initPreferences } from "@/core/preferences";

import App from "./App.vue";
import { setupVueQuery } from "@/plugins/vue-query";

async function bootstrap(namespace: string) {
  const app = createApp(App);

  // 初始化偏好设置
  await initPreferences({
    namespace,
    overrides: {
      app: {
        name: import.meta.env.VITE_APP_TITLE || "GoWind Admin",
        version: import.meta.env.VITE_APP_VERSION || "0.0.0",
        enableTenant: import.meta.env.VITE_APP_TENANT_ENABLED === "true",
      },
    },
  });

  // 注册全局组件
  registerGlobComp(app);

  setupVueQuery(app);

  // 注册自定义指令
  setupDirective(app);

  // 配置 pinia-tore
  await initStores(app, { namespace });

  // 配置路由及路由守卫
  setupRouter(app);

  // 国际化 i18n 配置
  await setupI18n(app);

  // 挂载应用
  app.mount("#app");
}

export { bootstrap };
