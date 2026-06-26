import type { AppRouteObject } from '@/core/router/types';
import { createLazyRoute } from '@/core/router';

/**
 * 物模型路由配置 / Thing-model routes
 *
 * 物模型为未来大模块，新建一级路由节点（`import.meta.glob` 自动导入）。
 * 当前仅落地"单位管理"子项，后续将扩展属性 / 物模型实例 / 设备数据归一化等。
 */
export const thingmodelRoutes: AppRouteObject[] = [
  {
    name: 'thingmodel',
    path: 'thingmodel',
    meta: {
      title: 'routes:thingmodel',
      icon: 'lucide:cpu',
      order: 1500, // 紧随业务模块之后、系统管理之前
      keepAlive: true,
    },
    children: [
      {
        name: 'thingmodel-unit',
        path: 'unit',
        element: createLazyRoute(() => import('@/pages/app/thingmodel/unit')),
        meta: {
          title: 'routes:thingmodelUnit',
          icon: 'lucide:ruler',
          order: 1,
          // 权限码由后端 SyncPermissions 自动从菜单 path 派生（见 backend/pkg/utils/converter/menu.go）：
          //   menu /thingmodel/unit (MENU) → "unit:view"
          // 平台/租户管理员通过 sys:platform_admin / sys:tenant_manager 角色码覆盖。
          // Permission codes are auto-derived from menu path by backend SyncPermissions.
          // authority: ['unit:view'],
        },
      },
    ],
  },
];

export default thingmodelRoutes;
