import type { Router } from "vue-router";

import { setupCommonGuard } from "./common";
import { setupAccessGuard } from "./access";

/**
 * 项目守卫配置
 * @param router
 */
function createRouterGuard(router: Router) {
  /** 通用 */
  setupCommonGuard(router);
  /** 权限访问 */
  setupAccessGuard(router);
}

export { createRouterGuard };
