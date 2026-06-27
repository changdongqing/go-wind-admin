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
      {
        name: 'thingmodel-feature',
        path: 'feature',
        element: createLazyRoute(() => import('@/pages/app/thingmodel/feature')),
        meta: {
          title: 'routes:thingmodelFeature',
          icon: 'lucide:boxes',
          order: 2,
          // 权限码由后端 SyncPermissions 自动从菜单 path 派生：
          //   menu /thingmodel/feature (MENU) → "feature:view"
          // authority: ['feature:view'],
        },
      },
      {
        name: 'thingmodel-category',
        path: 'category',
        element: createLazyRoute(() => import('@/pages/app/thingmodel/category')),
        meta: {
          title: 'routes:thingmodelCategory',
          icon: 'lucide:layers',
          order: 3,
          // 权限码由后端 SyncPermissions 自动从菜单 path 派生：
          //   menu /thingmodel/category (MENU) → "category:view"
          // authority: ['category:view'],
        },
      },
      {
        // 产品管理（模型管理入口②）/ Product management (Model management entry ②)
        // 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/05-前端实现设计.md §3.1
        // 注意：分类默认模型（入口①）不单独建菜单，复用 thingmodel:category:edit 权限码
        name: 'thingmodel-product',
        path: 'product',
        element: createLazyRoute(() => import('@/pages/app/thingmodel/product')),
        meta: {
          title: 'routes:thingmodelProduct',
          icon: 'carbon:product',
          order: 4,
        },
      },
      {
        // 产品详情页（隐藏菜单，通过列表点入）
        // Hide from menu; reached by clicking a row in the list.
        name: 'thingmodel-product-detail',
        path: 'product/:id',
        element: createLazyRoute(() => import('@/pages/app/thingmodel/product/ProductDetailPage')),
        meta: {
          title: 'routes:thingmodelProduct',
          hideInMenu: true,
          activePath: '/thingmodel/product',
        },
      },
    ],
  },
];

export default thingmodelRoutes;
