# 模型管理 · React 前端实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为已落地的"模型管理"后端（22 个 REST 接口）落地 React 前端实现：分类管理"行内 Drawer 配置默认模型"入口①，与"产品管理"新菜单入口②（含两步创建向导、Tab 化模型编辑器、特征 Drawer），仅改 `frontend/admin/react`，Vue 两端零变化。

**Architecture:** 严格镜像项目现有 `feature/category` 前端范式——`api/hooks/*.ts` React Query 整合 + `api/client.ts` apiClient 调用生成的 BFF 客户端；页面置于 `src/pages/app/thingmodel/product/`；路由通过 `router/modules/thingmodel.tsx` 自动菜单化。共享子组件（FeaturePicker / OverrideSpecForm / FeatureSourceTag / CategoryPicker）在 `src/pages/app/thingmodel/_shared/` 下，避免污染既有页面目录。

**Tech Stack:** React 19 / Ant Design v6 / @ant-design/pro-components ProTable / React Query / Iconify / react-i18next / TypeScript 5.x

**前置依赖**：后端 23 任务已完成（最新 commit `4183b3c0` 在 `chdq` 分支），生成的 TS 客户端已就位于 `src/api/generated/admin/service/v1/index.ts`（含 `categoryDefaultFeatureService` / `productService` / `productFeatureService` 三个 client 实例属性，132 个 export 引用）。

**范围**：仅 React 前端。Vue Element / Vue Vben 不动。

---

## 文件结构（落点全景）

新增：
```
frontend/admin/react/src/
├── api/hooks/
│   ├── category-default-feature.ts            ← Task 1
│   ├── product.ts                             ← Task 2
│   └── product-feature.ts                     ← Task 3
│
├── pages/app/thingmodel/_shared/              ← 共享子组件命名空间（新增目录）
│   ├── FeatureSourceTag.tsx                   ← Task 4
│   ├── FeaturePicker.tsx                      ← Task 5（含 modal）
│   └── OverrideSpecForm.tsx                   ← Task 6
│
└── pages/app/thingmodel/product/              ← 新增页面目录
    ├── index.tsx                              ← Task 8 (ProductList)
    ├── CreateProductWizard.tsx                ← Task 9
    ├── ProductDetailPage.tsx                  ← Task 10
    ├── ProductFeatureDrawer.tsx               ← Task 11
    └── constants.ts                           ← Task 8（状态/source 枚举映射）

frontend/admin/react/src/locales/
├── zh-CN/_modules/
│   ├── product.json                           ← Task 12
│   ├── product-feature.json                   ← Task 12
│   └── category-default-feature.json          ← Task 12
└── en-US/_modules/  (镜像 zh-CN 三份)        ← Task 12
```

修改：
```
src/pages/app/thingmodel/category/
├── CategoryDefaultFeaturesDrawer.tsx          ← Task 7 (新建)
└── index.tsx                                  ← Task 7（追加 level=4 行操作按钮）

src/router/modules/thingmodel.tsx              ← Task 12（追加 product 子路由）
src/locales/zh-CN/_core/routes.json            ← Task 12（追加 thingmodelProduct key）
src/locales/en-US/_core/routes.json            ← Task 12
```

---

## Task 1: hooks · `category-default-feature.ts`

**Files:**
- Create: `frontend/admin/react/src/api/hooks/category-default-feature.ts`

镜像 `src/api/hooks/feature.ts` 范式。7 个 hook 对应 7 个 BFF RPC（List/Get/Create/BatchAdd/Update/Delete/Reorder）。

- [ ] **Step 1: 写文件骨架（imports + 类型）**

```typescript
/**
 * 物模型 - 分类默认模型条目 hooks。
 * Category default feature management hooks.
 *
 * 镜像 `feature.ts` 的 React Query 整合，封装 7 个 BFF RPC。
 */
import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from '@tanstack/react-query';
import type {
  thingmodelservicev1_CategoryDefaultFeature,
  thingmodelservicev1_ListCategoryDefaultFeatureResponse,
  thingmodelservicev1_GetCategoryDefaultFeatureRequest,
  thingmodelservicev1_CreateCategoryDefaultFeatureRequest,
  thingmodelservicev1_BatchAddCategoryDefaultFeaturesRequest,
  thingmodelservicev1_BatchAddCategoryDefaultFeaturesResponse,
  thingmodelservicev1_UpdateCategoryDefaultFeatureRequest,
  thingmodelservicev1_DeleteCategoryDefaultFeatureRequest,
  thingmodelservicev1_ReorderCategoryDefaultFeaturesRequest,
} from '@/api/generated/admin/service/v1';
import { makeUpdateMask, type PaginationQuery } from '@/core/transport/rest';
import { queryClient } from '@/core';
import { apiClient } from '@/api/client';
```

- [ ] **Step 2: 实现 useListCategoryDefaultFeatures / fetchListCategoryDefaultFeatures**

```typescript
export function useListCategoryDefaultFeatures(
  query: PaginationQuery,
  options?: UseQueryOptions<thingmodelservicev1_ListCategoryDefaultFeatureResponse, Error>,
) {
  return useQuery({
    queryKey: ['listCategoryDefaultFeatures', query],
    queryFn: () => apiClient.categoryDefaultFeatureService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListCategoryDefaultFeatures(query: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ['listCategoryDefaultFeatures', query],
    queryFn: () => apiClient.categoryDefaultFeatureService.List(query.toRawParams()),
    retry: 0,
  });
}
```

- [ ] **Step 3: 实现 Get / Create / BatchAdd**

```typescript
export function useGetCategoryDefaultFeature(
  req: thingmodelservicev1_GetCategoryDefaultFeatureRequest,
  options?: UseQueryOptions<thingmodelservicev1_CategoryDefaultFeature, Error>,
) {
  return useQuery({
    queryKey: ['getCategoryDefaultFeature', req],
    queryFn: () => apiClient.categoryDefaultFeatureService.Get(req),
    enabled: !!req.id,
    ...options,
  });
}

export function useCreateCategoryDefaultFeature(
  options?: UseMutationOptions<unknown, Error, thingmodelservicev1_CreateCategoryDefaultFeatureRequest>,
) {
  return useMutation({
    mutationFn: (req) => apiClient.categoryDefaultFeatureService.Create(req),
    ...options,
  });
}

export function useBatchAddCategoryDefaultFeatures(
  options?: UseMutationOptions<
    thingmodelservicev1_BatchAddCategoryDefaultFeaturesResponse,
    Error,
    thingmodelservicev1_BatchAddCategoryDefaultFeaturesRequest
  >,
) {
  return useMutation({
    mutationFn: (req) => apiClient.categoryDefaultFeatureService.BatchAdd(req),
    ...options,
  });
}
```

- [ ] **Step 4: 实现 Update（含 makeUpdateMask）/ Delete / Reorder**

```typescript
export function useUpdateCategoryDefaultFeature(
  options?: UseMutationOptions<
    unknown,
    Error,
    { id: number; data: Partial<thingmodelservicev1_CategoryDefaultFeature>; mask?: string[] }
  >,
) {
  return useMutation({
    mutationFn: async ({ id, data, mask }) => {
      const updateMask = mask ?? makeUpdateMask(data);
      return apiClient.categoryDefaultFeatureService.Update({
        id,
        data,
        updateMask: { paths: updateMask },
      });
    },
    ...options,
  });
}

export function useDeleteCategoryDefaultFeature(
  options?: UseMutationOptions<unknown, Error, thingmodelservicev1_DeleteCategoryDefaultFeatureRequest>,
) {
  return useMutation({
    mutationFn: (req) => apiClient.categoryDefaultFeatureService.Delete(req),
    ...options,
  });
}

export function useReorderCategoryDefaultFeatures(
  options?: UseMutationOptions<unknown, Error, thingmodelservicev1_ReorderCategoryDefaultFeaturesRequest>,
) {
  return useMutation({
    mutationFn: (req) => apiClient.categoryDefaultFeatureService.Reorder(req),
    ...options,
  });
}
```

- [ ] **Step 5: pnpm typecheck 通过**

Run: `cd frontend/admin/react && pnpm typecheck 2>&1 | tail -10`
Expected: 0 errors

- [ ] **Step 6: Commit**

```bash
git add frontend/admin/react/src/api/hooks/category-default-feature.ts
git commit -m "feat(react): add category-default-feature hooks (7 RPC wrappers)"
```

---

## Task 2: hooks · `product.ts`

**Files:**
- Create: `frontend/admin/react/src/api/hooks/product.ts`

7 个 hook：List/Get/Create/Update/Delete/Publish/Unpublish。

- [ ] **Step 1: 写文件骨架 + List/Get**

```typescript
/**
 * 物模型 - 产品管理 hooks。
 */
import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from '@tanstack/react-query';
import type {
  thingmodelservicev1_Product,
  thingmodelservicev1_ListProductResponse,
  thingmodelservicev1_GetProductRequest,
  thingmodelservicev1_CreateProductRequest,
  thingmodelservicev1_UpdateProductRequest,
  thingmodelservicev1_DeleteProductRequest,
  thingmodelservicev1_PublishProductRequest,
  thingmodelservicev1_UnpublishProductRequest,
} from '@/api/generated/admin/service/v1';
import { makeUpdateMask, type PaginationQuery } from '@/core/transport/rest';
import { queryClient } from '@/core';
import { apiClient } from '@/api/client';

export function useListProducts(
  query: PaginationQuery,
  options?: UseQueryOptions<thingmodelservicev1_ListProductResponse, Error>,
) {
  return useQuery({
    queryKey: ['listProducts', query],
    queryFn: () => apiClient.productService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListProducts(query: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ['listProducts', query],
    queryFn: () => apiClient.productService.List(query.toRawParams()),
    retry: 0,
  });
}

export function useGetProduct(
  req: thingmodelservicev1_GetProductRequest,
  options?: UseQueryOptions<thingmodelservicev1_Product, Error>,
) {
  return useQuery({
    queryKey: ['getProduct', req],
    queryFn: () => apiClient.productService.Get(req),
    enabled: !!(req.id || req.code),
    ...options,
  });
}

export async function fetchGetProduct(req: thingmodelservicev1_GetProductRequest) {
  return queryClient.fetchQuery({
    queryKey: ['getProduct', req],
    queryFn: () => apiClient.productService.Get(req),
    retry: 0,
  });
}
```

- [ ] **Step 2: Create / Update / Delete**

```typescript
export function useCreateProduct(
  options?: UseMutationOptions<unknown, Error, thingmodelservicev1_CreateProductRequest>,
) {
  return useMutation({
    mutationFn: (req) => apiClient.productService.Create(req),
    ...options,
  });
}

export function useUpdateProduct(
  options?: UseMutationOptions<
    unknown,
    Error,
    { id: number; data: Partial<thingmodelservicev1_Product>; mask?: string[] }
  >,
) {
  return useMutation({
    mutationFn: async ({ id, data, mask }) => {
      const updateMask = mask ?? makeUpdateMask(data);
      return apiClient.productService.Update({
        id,
        data,
        updateMask: { paths: updateMask },
      });
    },
    ...options,
  });
}

export function useDeleteProduct(
  options?: UseMutationOptions<unknown, Error, thingmodelservicev1_DeleteProductRequest>,
) {
  return useMutation({
    mutationFn: (req) => apiClient.productService.Delete(req),
    ...options,
  });
}
```

- [ ] **Step 3: Publish / Unpublish**

```typescript
export function usePublishProduct(
  options?: UseMutationOptions<unknown, Error, thingmodelservicev1_PublishProductRequest>,
) {
  return useMutation({
    mutationFn: (req) => apiClient.productService.Publish(req),
    ...options,
  });
}

export function useUnpublishProduct(
  options?: UseMutationOptions<unknown, Error, thingmodelservicev1_UnpublishProductRequest>,
) {
  return useMutation({
    mutationFn: (req) => apiClient.productService.Unpublish(req),
    ...options,
  });
}
```

- [ ] **Step 4: pnpm typecheck 通过**

- [ ] **Step 5: Commit**

```bash
git add frontend/admin/react/src/api/hooks/product.ts
git commit -m "feat(react): add product management hooks (List/Get/CRUD/Publish/Unpublish)"
```

---

## Task 3: hooks · `product-feature.ts`

**Files:**
- Create: `frontend/admin/react/src/api/hooks/product-feature.ts`

8 个 hook：List/Get/Create/PullFromDefault/CloneFromProduct/Update/Delete/Reorder。

- [ ] **Step 1: 写 imports + List/Get（含 effective_spec 备注）**

```typescript
/**
 * 物模型 - 产品下特征条目 hooks。
 *
 * 关键 RPC：PullFromDefault（批量从分类默认模型拉取）、CloneFromProduct（从另一产品克隆）。
 * Get 返回的 effective_spec 由后端合并 feature_snapshot + override_spec 后填充。
 */
import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from '@tanstack/react-query';
import type {
  thingmodelservicev1_ProductFeature,
  thingmodelservicev1_ListProductFeatureResponse,
  thingmodelservicev1_GetProductFeatureRequest,
  thingmodelservicev1_CreateProductFeatureRequest,
  thingmodelservicev1_PullFromDefaultRequest,
  thingmodelservicev1_PullFromDefaultResponse,
  thingmodelservicev1_CloneFromProductRequest,
  thingmodelservicev1_CloneFromProductResponse,
  thingmodelservicev1_UpdateProductFeatureRequest,
  thingmodelservicev1_DeleteProductFeatureRequest,
  thingmodelservicev1_ReorderProductFeaturesRequest,
} from '@/api/generated/admin/service/v1';
import { makeUpdateMask, type PaginationQuery } from '@/core/transport/rest';
import { queryClient } from '@/core';
import { apiClient } from '@/api/client';

export function useListProductFeatures(
  query: PaginationQuery,
  options?: UseQueryOptions<thingmodelservicev1_ListProductFeatureResponse, Error>,
) {
  return useQuery({
    queryKey: ['listProductFeatures', query],
    queryFn: () => apiClient.productFeatureService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListProductFeatures(query: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ['listProductFeatures', query],
    queryFn: () => apiClient.productFeatureService.List(query.toRawParams()),
    retry: 0,
  });
}

export function useGetProductFeature(
  req: thingmodelservicev1_GetProductFeatureRequest,
  options?: UseQueryOptions<thingmodelservicev1_ProductFeature, Error>,
) {
  return useQuery({
    queryKey: ['getProductFeature', req],
    queryFn: () => apiClient.productFeatureService.Get(req),
    enabled: !!req.id,
    ...options,
  });
}
```

- [ ] **Step 2: Create / PullFromDefault / CloneFromProduct**

```typescript
export function useCreateProductFeature(
  options?: UseMutationOptions<unknown, Error, thingmodelservicev1_CreateProductFeatureRequest>,
) {
  return useMutation({
    mutationFn: (req) => apiClient.productFeatureService.Create(req),
    ...options,
  });
}

export function usePullFromDefault(
  options?: UseMutationOptions<
    thingmodelservicev1_PullFromDefaultResponse,
    Error,
    thingmodelservicev1_PullFromDefaultRequest
  >,
) {
  return useMutation({
    mutationFn: (req) => apiClient.productFeatureService.PullFromDefault(req),
    ...options,
  });
}

export function useCloneFromProduct(
  options?: UseMutationOptions<
    thingmodelservicev1_CloneFromProductResponse,
    Error,
    thingmodelservicev1_CloneFromProductRequest
  >,
) {
  return useMutation({
    mutationFn: (req) => apiClient.productFeatureService.CloneFromProduct(req),
    ...options,
  });
}
```

- [ ] **Step 3: Update / Delete / Reorder**

```typescript
export function useUpdateProductFeature(
  options?: UseMutationOptions<
    unknown,
    Error,
    { id: number; data: Partial<thingmodelservicev1_ProductFeature>; mask?: string[] }
  >,
) {
  return useMutation({
    mutationFn: async ({ id, data, mask }) => {
      const updateMask = mask ?? makeUpdateMask(data);
      return apiClient.productFeatureService.Update({
        id,
        data,
        updateMask: { paths: updateMask },
      });
    },
    ...options,
  });
}

export function useDeleteProductFeature(
  options?: UseMutationOptions<unknown, Error, thingmodelservicev1_DeleteProductFeatureRequest>,
) {
  return useMutation({
    mutationFn: (req) => apiClient.productFeatureService.Delete(req),
    ...options,
  });
}

export function useReorderProductFeatures(
  options?: UseMutationOptions<unknown, Error, thingmodelservicev1_ReorderProductFeaturesRequest>,
) {
  return useMutation({
    mutationFn: (req) => apiClient.productFeatureService.Reorder(req),
    ...options,
  });
}
```

- [ ] **Step 4: pnpm typecheck 通过 + Commit**

```bash
git add frontend/admin/react/src/api/hooks/product-feature.ts
git commit -m "feat(react): add product-feature hooks (PullFromDefault, CloneFromProduct, CRUD)"
```

---

## Task 4: 共享组件 · `FeatureSourceTag`

**Files:**
- Create: `frontend/admin/react/src/pages/app/thingmodel/_shared/FeatureSourceTag.tsx`

DEF / GLO / LOC 三色 Tag。

- [ ] **Step 1: 实现组件**

```tsx
/**
 * 产品特征来源标签 / Product feature source tag.
 *
 * DEF (默认模型) = green、GLO (全局特征库) = blue、LOC (本地) = purple。
 */
import { Tag } from 'antd';
import { useTranslation } from 'react-i18next';

export type FeatureSourceValue = 'DEFAULT' | 'GLOBAL' | 'LOCAL';

interface Props {
  source?: FeatureSourceValue | string;
}

const palette: Record<FeatureSourceValue, { color: string; abbr: string }> = {
  DEFAULT: { color: 'green', abbr: 'DEF' },
  GLOBAL: { color: 'blue', abbr: 'GLO' },
  LOCAL: { color: 'purple', abbr: 'LOC' },
};

export const FeatureSourceTag = ({ source }: Props) => {
  const { t } = useTranslation('product-feature');
  if (!source) return null;
  const cfg = palette[source as FeatureSourceValue];
  if (!cfg) return <Tag>{source}</Tag>;
  return (
    <Tag color={cfg.color} title={t(`source.${source}`)}>
      {cfg.abbr}
    </Tag>
  );
};

export default FeatureSourceTag;
```

- [ ] **Step 2: pnpm typecheck 通过 + Commit**

```bash
git add frontend/admin/react/src/pages/app/thingmodel/_shared/FeatureSourceTag.tsx
git commit -m "feat(react): add FeatureSourceTag shared component (DEF/GLO/LOC)"
```

---

## Task 5: 共享组件 · `FeaturePicker` Modal

**Files:**
- Create: `frontend/admin/react/src/pages/app/thingmodel/_shared/FeaturePicker.tsx`

打开后从全局特征库选特征（支持按 type/keyword 过滤），多选返回 feature[]。

- [ ] **Step 1: 实现 FeaturePicker**

```tsx
/**
 * 全局特征库选择器 Modal。
 *
 * 用于：
 *   - 分类默认模型 Drawer 内 "+ 新增特征" 按钮
 *   - 产品详情页 "添加 → 从全局特征库添加"
 *
 * 多选 + 类型/关键字过滤 + 单页 50 条。
 */
import { useMemo, useState } from 'react';
import { Modal, Table, Tabs, Input, Tag } from 'antd';
import { useTranslation } from 'react-i18next';
import { useListFeatures } from '@/api/hooks/feature';
import { PaginationQuery } from '@/core';
import type { thingmodelservicev1_Feature } from '@/api/generated/admin/service/v1';

type FeatureType = 'PROPERTY' | 'EVENT' | 'SERVICE' | 'RELATION';

interface Props {
  open: boolean;
  excludedIds?: number[]; // 已存在的 feature_id，禁选
  onCancel: () => void;
  onConfirm: (features: thingmodelservicev1_Feature[]) => void;
  /** 仅允许选择某些类型；不传则全部允许 */
  allowedTypes?: FeatureType[];
}

const ALL_TYPES: FeatureType[] = ['PROPERTY', 'EVENT', 'SERVICE', 'RELATION'];

export const FeaturePicker = ({
  open,
  excludedIds,
  onCancel,
  onConfirm,
  allowedTypes,
}: Props) => {
  const { t } = useTranslation(['feature', 'common']);
  const types = allowedTypes ?? ALL_TYPES;
  const [activeType, setActiveType] = useState<FeatureType>(types[0]);
  const [keyword, setKeyword] = useState('');
  const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);
  const [selectedRows, setSelectedRows] = useState<thingmodelservicev1_Feature[]>([]);

  const query = useMemo(() => {
    const q = new PaginationQuery({ page: 1, pageSize: 50 });
    q.query = {
      feature_type: { value: activeType, condition: 'equal' },
      ...(keyword
        ? {
            name: { value: keyword, condition: 'icontains' },
          }
        : {}),
    };
    return q;
  }, [activeType, keyword]);

  const { data, isLoading } = useListFeatures(query, { enabled: open });

  const items = (data?.items ?? []).filter((f) => !excludedIds?.includes(f.id ?? 0));

  return (
    <Modal
      open={open}
      title={t('common:select') + t('feature:pageTitle')}
      width={760}
      onCancel={onCancel}
      onOk={() => {
        onConfirm(selectedRows);
        setSelectedKeys([]);
        setSelectedRows([]);
      }}
      okButtonProps={{ disabled: selectedRows.length === 0 }}
    >
      <Tabs
        activeKey={activeType}
        onChange={(k) => setActiveType(k as FeatureType)}
        items={types.map((tp) => ({
          key: tp,
          label: t(`feature:featureTypeMap.${tp}`),
        }))}
      />
      <Input.Search
        allowClear
        placeholder={t('feature:keywordPlaceholder', '搜索名称')}
        onSearch={setKeyword}
        style={{ marginBottom: 12 }}
      />
      <Table
        rowKey="id"
        size="small"
        loading={isLoading}
        dataSource={items}
        rowSelection={{
          selectedRowKeys: selectedKeys,
          onChange: (keys, rows) => {
            setSelectedKeys(keys);
            setSelectedRows(rows);
          },
        }}
        pagination={false}
        scroll={{ y: 360 }}
        columns={[
          { title: t('feature:code'), dataIndex: 'code', width: 140 },
          { title: t('feature:identifier'), dataIndex: 'identifier', width: 160 },
          { title: t('feature:name'), dataIndex: 'name' },
          {
            title: t('feature:dataType'),
            dataIndex: 'dataType',
            width: 100,
            render: (v: string) => (v ? <Tag>{v}</Tag> : null),
          },
        ]}
      />
    </Modal>
  );
};

export default FeaturePicker;
```

- [ ] **Step 2: pnpm typecheck 通过 + Commit**

```bash
git add frontend/admin/react/src/pages/app/thingmodel/_shared/FeaturePicker.tsx
git commit -m "feat(react): add FeaturePicker modal for selecting features from global library"
```

---

## Task 6: 共享组件 · `OverrideSpecForm`

**Files:**
- Create: `frontend/admin/react/src/pages/app/thingmodel/_shared/OverrideSpecForm.tsx`

仅渲染白名单字段：constraints / unit / defaultValue / displayName / description / required。

- [ ] **Step 1: 实现表单**

```tsx
/**
 * 特征覆写表单（白名单字段：constraints/unit/defaultValue/displayName/description/required）。
 *
 * 收口规则：proto FeatureOverrideSpec 本身只允许这 6 个字段；本表单是 UI 层最后一道收口。
 * 非 property 特征上 constraints/unit/defaultValue 灰显（后端会校验拒收）。
 */
import { Form, InputNumber, Input, Switch, Alert } from 'antd';
import { useTranslation } from 'react-i18next';
import type {
  thingmodelservicev1_FeatureOverrideSpec,
  thingmodelservicev1_FeatureSpec,
  thingmodelservicev1_FeatureType,
} from '@/api/generated/admin/service/v1';

interface Props {
  /** 用于决定能否覆写 constraints/unit/defaultValue 的目标 feature 类型 */
  featureType?: thingmodelservicev1_FeatureType;
  /** 当前 snapshot（只读展示原值用） */
  snapshot?: thingmodelservicev1_FeatureSpec;
  /** 受控 override 值 */
  value?: thingmodelservicev1_FeatureOverrideSpec | null;
  onChange?: (v: thingmodelservicev1_FeatureOverrideSpec | null) => void;
  /** 整个表单只读（PUBLISHED 产品的非白名单字段灰显） */
  readonly?: boolean;
}

export const OverrideSpecForm = ({
  featureType,
  snapshot,
  value,
  onChange,
  readonly,
}: Props) => {
  const { t } = useTranslation(['product-feature', 'common']);

  const isProperty = featureType === 'PROPERTY';

  const merge = (patch: Partial<thingmodelservicev1_FeatureOverrideSpec>) => {
    onChange?.({ ...(value ?? {}), ...patch });
  };

  return (
    <Form layout="vertical" disabled={readonly}>
      <Form.Item label={t('override.displayName')}>
        <Input
          value={value?.displayName}
          onChange={(e) => merge({ displayName: e.target.value })}
          placeholder={t('override.displayNamePlaceholder', '作用域内展示别名')}
        />
      </Form.Item>

      <Form.Item label={t('override.description')}>
        <Input.TextArea
          rows={2}
          value={value?.description}
          onChange={(e) => merge({ description: e.target.value })}
        />
      </Form.Item>

      {!isProperty && (
        <Alert
          type="info"
          showIcon
          message={t('override.nonPropertyHint', '非属性类型不支持覆写 constraints / unit / defaultValue')}
          style={{ marginBottom: 12 }}
        />
      )}

      <Form.Item label={t('override.constraints.range', '取值范围 min/max/step')}>
        <Input.Group compact>
          <InputNumber
            placeholder="min"
            disabled={!isProperty}
            value={value?.constraints?.min}
            onChange={(v) =>
              merge({ constraints: { ...(value?.constraints ?? {}), min: v ?? undefined } })
            }
            style={{ width: '33%' }}
          />
          <InputNumber
            placeholder="max"
            disabled={!isProperty}
            value={value?.constraints?.max}
            onChange={(v) =>
              merge({ constraints: { ...(value?.constraints ?? {}), max: v ?? undefined } })
            }
            style={{ width: '33%' }}
          />
          <InputNumber
            placeholder="step"
            disabled={!isProperty}
            value={value?.constraints?.step}
            onChange={(v) =>
              merge({ constraints: { ...(value?.constraints ?? {}), step: v ?? undefined } })
            }
            style={{ width: '34%' }}
          />
        </Input.Group>
      </Form.Item>

      <Form.Item label={t('override.defaultValue')}>
        <Input
          disabled={!isProperty}
          value={value?.defaultValue}
          onChange={(e) => merge({ defaultValue: e.target.value })}
        />
      </Form.Item>

      <Form.Item label={t('override.required', 'required（service 输入参数用）')}>
        <Switch
          checked={value?.required?.value ?? undefined}
          onChange={(checked) => merge({ required: { value: checked } })}
        />
      </Form.Item>

      {snapshot && isProperty && (
        <Alert
          type="default"
          message={
            <span>
              {t('override.snapshotPreview', '原始范围')}：min=
              {snapshot.spec?.property?.constraints?.min ?? '-'} max=
              {snapshot.spec?.property?.constraints?.max ?? '-'}
            </span>
          }
        />
      )}
    </Form>
  );
};

export default OverrideSpecForm;
```

- [ ] **Step 2: pnpm typecheck 通过 + Commit**

```bash
git add frontend/admin/react/src/pages/app/thingmodel/_shared/OverrideSpecForm.tsx
git commit -m "feat(react): add OverrideSpecForm whitelist editor"
```

---

## Task 7: 入口① · 分类管理 行内 Drawer

**Files:**
- Create: `frontend/admin/react/src/pages/app/thingmodel/category/CategoryDefaultFeaturesDrawer.tsx`
- Modify: `frontend/admin/react/src/pages/app/thingmodel/category/index.tsx`（追加操作列按钮）

- [ ] **Step 1: 创建 Drawer 组件**

```tsx
/**
 * 配置默认模型 Drawer（入口①）— 仅 level=4 细类调用。
 *
 * 内含 Tab(全部/属性/事件/服务/关系) + ProTable + 行内编辑 Override + + 新增特征。
 */
import { useRef, useState } from 'react';
import { Drawer, Button, Space, Tabs, Popconfirm, App, Modal } from 'antd';
import { ProTable, type ActionType, type ProColumns } from '@ant-design/pro-components';
import { useTranslation } from 'react-i18next';
import { useQueryClient } from '@tanstack/react-query';
import {
  fetchListCategoryDefaultFeatures,
  useBatchAddCategoryDefaultFeatures,
  useDeleteCategoryDefaultFeature,
  useUpdateCategoryDefaultFeature,
} from '@/api/hooks/category-default-feature';
import { PaginationQuery } from '@/core';
import type {
  thingmodelservicev1_Category,
  thingmodelservicev1_CategoryDefaultFeature,
  thingmodelservicev1_Feature,
  thingmodelservicev1_FeatureOverrideSpec,
} from '@/api/generated/admin/service/v1';
import FeaturePicker from '../_shared/FeaturePicker';
import OverrideSpecForm from '../_shared/OverrideSpecForm';

interface Props {
  open: boolean;
  category: thingmodelservicev1_Category | null;
  onClose: () => void;
}

type TabKey = 'ALL' | 'PROPERTY' | 'EVENT' | 'SERVICE' | 'RELATION';

export const CategoryDefaultFeaturesDrawer = ({ open, category, onClose }: Props) => {
  const { t } = useTranslation(['category-default-feature', 'common']);
  const actionRef = useRef<ActionType>(null);
  const qc = useQueryClient();
  const { message } = App.useApp();

  const [tab, setTab] = useState<TabKey>('ALL');
  const [pickerOpen, setPickerOpen] = useState(false);
  const [editing, setEditing] = useState<thingmodelservicev1_CategoryDefaultFeature | null>(null);
  const [editOverride, setEditOverride] = useState<thingmodelservicev1_FeatureOverrideSpec | null>(null);

  const { mutate: batchAdd } = useBatchAddCategoryDefaultFeatures({
    onSuccess: () => {
      message.success(t('addSuccess'));
      actionRef.current?.reload();
      qc.invalidateQueries({ queryKey: ['listCategoryDefaultFeatures'] });
    },
  });
  const { mutate: doDelete } = useDeleteCategoryDefaultFeature({
    onSuccess: () => {
      message.success(t('common:deleteSuccess'));
      actionRef.current?.reload();
    },
  });
  const { mutate: doUpdate } = useUpdateCategoryDefaultFeature({
    onSuccess: () => {
      message.success(t('common:updateSuccess'));
      actionRef.current?.reload();
      setEditing(null);
    },
  });

  const columns: ProColumns<thingmodelservicev1_CategoryDefaultFeature>[] = [
    { title: t('featureCode'), dataIndex: 'featureCode', width: 140 },
    { title: t('featureIdentifier'), dataIndex: 'featureIdentifier', width: 160 },
    { title: t('featureName'), dataIndex: 'featureName' },
    { title: t('featureType'), dataIndex: 'featureType', width: 100 },
    {
      title: t('overridden'),
      dataIndex: 'overrideSpec',
      width: 90,
      render: (_, row) => (row.overrideSpec ? t('yes') : '-'),
    },
    {
      title: t('common:operation'),
      key: 'op',
      width: 160,
      render: (_, row) => (
        <Space size="small">
          <Button
            size="small"
            type="link"
            onClick={() => {
              setEditing(row);
              setEditOverride(row.overrideSpec ?? null);
            }}
          >
            {t('common:edit')}
          </Button>
          <Popconfirm
            title={t('common:confirmDelete')}
            onConfirm={() => doDelete({ ids: [row.id!] })}
          >
            <Button size="small" type="link" danger>
              {t('common:delete')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Drawer
      width={1024}
      open={open}
      onClose={onClose}
      title={category ? `${t('drawerTitle')} — ${category.name} (${category.code})` : ''}
      destroyOnClose
    >
      <Tabs
        activeKey={tab}
        onChange={(k) => setTab(k as TabKey)}
        items={[
          { key: 'ALL', label: t('tab.all') },
          { key: 'PROPERTY', label: t('tab.property') },
          { key: 'EVENT', label: t('tab.event') },
          { key: 'SERVICE', label: t('tab.service') },
          { key: 'RELATION', label: t('tab.relation') },
        ]}
      />

      <ProTable
        actionRef={actionRef}
        rowKey="id"
        search={false}
        columns={columns}
        toolBarRender={() => [
          <Button key="add" type="primary" onClick={() => setPickerOpen(true)}>
            + {t('addFeature')}
          </Button>,
        ]}
        request={async (params) => {
          if (!category?.id) return { data: [], success: true, total: 0 };
          const q = new PaginationQuery({
            page: params.current ?? 1,
            pageSize: params.pageSize ?? 20,
          });
          q.query = {
            category_id: { value: category.id, condition: 'equal' },
            ...(tab !== 'ALL' ? { feature_type: { value: tab, condition: 'equal' } } : {}),
          };
          const resp = await fetchListCategoryDefaultFeatures(q);
          return { data: resp.items ?? [], success: true, total: Number(resp.total ?? 0) };
        }}
      />

      <FeaturePicker
        open={pickerOpen}
        onCancel={() => setPickerOpen(false)}
        onConfirm={(feats) => {
          setPickerOpen(false);
          if (!category?.id || feats.length === 0) return;
          batchAdd({
            categoryId: category.id,
            items: feats.map((f) => ({ featureId: f.id! })),
          });
        }}
      />

      <Modal
        open={!!editing}
        title={t('editOverride')}
        onCancel={() => setEditing(null)}
        onOk={() => {
          if (!editing) return;
          doUpdate({
            id: editing.id!,
            data: { overrideSpec: editOverride ?? undefined },
            mask: ['override_spec'],
          });
        }}
      >
        <OverrideSpecForm
          featureType={editing?.featureType}
          value={editOverride}
          onChange={setEditOverride}
        />
      </Modal>
    </Drawer>
  );
};

export default CategoryDefaultFeaturesDrawer;
```

- [ ] **Step 2: 改 `category/index.tsx`：level=4 行追加"配置默认模型"按钮**

在 ProTable 操作列里追加：

```tsx
// 已有 columns 的操作列里加入：
{row.level === 4 && (
  <Button
    size="small"
    type="link"
    onClick={() => {
      setCdfCategory(row);
      setCdfOpen(true);
    }}
  >
    {t('configDefaultModel')}
  </Button>
)}
```

并在 `CategoryManagement` 组件内 useState：

```tsx
const [cdfOpen, setCdfOpen] = useState(false);
const [cdfCategory, setCdfCategory] = useState<thingmodelservicev1_Category | null>(null);
```

页面底部渲染：

```tsx
<CategoryDefaultFeaturesDrawer
  open={cdfOpen}
  category={cdfCategory}
  onClose={() => setCdfOpen(false)}
/>
```

- [ ] **Step 3: pnpm typecheck + 浏览器手测（启动 dev）**

Run: `cd frontend/admin/react && pnpm typecheck 2>&1 | tail -5`
Expected: 0 errors

可选手测：`pnpm dev`、登录、进入分类管理、找 level=4 行（如 20010100 电动压缩式冷水机组），点 "配置默认模型"，应看到 10 条种子数据。

- [ ] **Step 4: Commit**

```bash
git add frontend/admin/react/src/pages/app/thingmodel/category/CategoryDefaultFeaturesDrawer.tsx \
        frontend/admin/react/src/pages/app/thingmodel/category/index.tsx
git commit -m "feat(react): integrate CategoryDefaultFeaturesDrawer into level=4 rows (entry ①)"
```

---

## Task 8: 入口② · 产品列表 (ProTable)

**Files:**
- Create: `frontend/admin/react/src/pages/app/thingmodel/product/index.tsx`
- Create: `frontend/admin/react/src/pages/app/thingmodel/product/constants.ts`

- [ ] **Step 1: 写 constants.ts**

```typescript
import type { ProSchemaValueEnumObj } from '@ant-design/pro-components';

export const statusValueEnum: ProSchemaValueEnumObj = {
  DRAFT: { text: '草稿', status: 'Default' },
  PUBLISHED: { text: '已发布', status: 'Success' },
};

export const enabledValueEnum: ProSchemaValueEnumObj = {
  true: { text: '启用', status: 'Success' },
  false: { text: '停用', status: 'Default' },
};
```

- [ ] **Step 2: 写 index.tsx (列表页)**

```tsx
/**
 * 产品管理列表 / Product list (entry ②).
 *
 * ProTable + 搜索（name/code/manufacturer/status）+ "新增产品" 按钮（打开向导）。
 * 点击行进入 ProductDetailPage（路由 :id）。
 */
import { useRef, useState } from 'react';
import { Button, Space, Tag, Popconfirm, App } from 'antd';
import { ProTable, type ActionType, type ProColumns } from '@ant-design/pro-components';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import ContentContainer from '@/layouts/components/PageContainer/ContentContainer';
import { useProTableScrollY } from '@/hooks/useProTableScrollY';
import { PaginationQuery } from '@/core';
import { fetchListProducts, useDeleteProduct } from '@/api/hooks/product';
import type { thingmodelservicev1_Product } from '@/api/generated/admin/service/v1';
import CreateProductWizard from './CreateProductWizard';
import { statusValueEnum, enabledValueEnum } from './constants';

const ProductManagement = () => {
  const { t } = useTranslation(['product', 'common']);
  const navigate = useNavigate();
  const actionRef = useRef<ActionType>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const tableScrollY = useProTableScrollY(containerRef);
  const { message } = App.useApp();
  const [wizardOpen, setWizardOpen] = useState(false);

  const { mutate: del } = useDeleteProduct({
    onSuccess: () => {
      message.success(t('common:deleteSuccess'));
      actionRef.current?.reload();
    },
    onError: (err) => message.error(err.message),
  });

  const columns: ProColumns<thingmodelservicev1_Product>[] = [
    { title: t('code'), dataIndex: 'code', width: 180, copyable: true },
    { title: t('name'), dataIndex: 'name' },
    {
      title: t('category'),
      dataIndex: 'categoryName',
      width: 200,
      render: (_, row) => <span>{row.categoryName} ({row.categoryCode})</span>,
    },
    { title: t('manufacturer'), dataIndex: 'manufacturer', width: 120 },
    { title: t('modelNo'), dataIndex: 'modelNo', width: 140 },
    {
      title: t('status'),
      dataIndex: 'status',
      width: 100,
      valueEnum: statusValueEnum,
    },
    {
      title: t('isEnabled'),
      dataIndex: 'isEnabled',
      width: 90,
      valueEnum: enabledValueEnum,
    },
    {
      title: t('common:operation'),
      key: 'op',
      width: 180,
      render: (_, row) => (
        <Space size="small">
          <Button size="small" type="link" onClick={() => navigate(`/thingmodel/product/${row.id}`)}>
            {t('common:edit')}
          </Button>
          <Popconfirm title={t('common:confirmDelete')} onConfirm={() => del({ ids: [row.id!] })}>
            <Button size="small" type="link" danger>
              {t('common:delete')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <ContentContainer ref={containerRef} heightMode="fixed" padding="16px" bottomMargin={0}>
      <ProTable
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        scroll={{ y: tableScrollY }}
        toolBarRender={() => [
          <Button key="add" type="primary" onClick={() => setWizardOpen(true)}>
            + {t('createProduct')}
          </Button>,
        ]}
        request={async (params) => {
          const q = new PaginationQuery({
            page: params.current ?? 1,
            pageSize: params.pageSize ?? 20,
          });
          const filters: Record<string, unknown> = {};
          if (params.status) filters.status = { value: params.status, condition: 'equal' };
          if (params.manufacturer)
            filters.manufacturer = { value: params.manufacturer, condition: 'icontains' };
          if (params.name) filters.name = { value: params.name, condition: 'icontains' };
          if (params.code) filters.code = { value: params.code, condition: 'icontains' };
          q.query = filters;
          const resp = await fetchListProducts(q);
          return { data: resp.items ?? [], success: true, total: Number(resp.total ?? 0) };
        }}
      />
      <CreateProductWizard
        open={wizardOpen}
        onClose={(reload) => {
          setWizardOpen(false);
          if (reload) actionRef.current?.reload();
        }}
      />
    </ContentContainer>
  );
};

export default ProductManagement;
```

- [ ] **Step 3: pnpm typecheck（CreateProductWizard 未实现先 stub 一下让 typecheck 过；下个 Task 实现真版）**

临时在 `CreateProductWizard.tsx` 占位：

```tsx
import { Modal } from 'antd';
interface Props { open: boolean; onClose: (reload?: boolean) => void; }
const CreateProductWizard = ({ open, onClose }: Props) => (
  <Modal open={open} onCancel={() => onClose(false)} onOk={() => onClose(true)} title="新增产品">
    placeholder - 待 Task 9 实现
  </Modal>
);
export default CreateProductWizard;
```

- [ ] **Step 4: Commit**

```bash
git add frontend/admin/react/src/pages/app/thingmodel/product/index.tsx \
        frontend/admin/react/src/pages/app/thingmodel/product/constants.ts \
        frontend/admin/react/src/pages/app/thingmodel/product/CreateProductWizard.tsx
git commit -m "feat(react): add Product list page (entry ②) with ProTable & filters"
```

---

## Task 9: 入口② · 两步创建向导

**Files:**
- Modify: `frontend/admin/react/src/pages/app/thingmodel/product/CreateProductWizard.tsx`（替换 stub）

- [ ] **Step 1: 实现两步向导**

```tsx
/**
 * 新增产品两步向导 / Two-step product creation wizard.
 *
 * Step 1: 基本信息（含 level=4 分类级联选择）
 * Step 2: 拉取默认模型（默认全选，可反选 + SKIP/REPLACE）
 *
 * 提交流程：CreateProduct → 拿到 id → PullFromDefault。
 */
import { useEffect, useState } from 'react';
import { Modal, Steps, Form, Input, Select, Cascader, Checkbox, Radio, App, Spin } from 'antd';
import { useTranslation } from 'react-i18next';
import { useCreateProduct, fetchGetProduct } from '@/api/hooks/product';
import { usePullFromDefault, fetchListProductFeatures } from '@/api/hooks/product-feature';
import { fetchListCategoryDefaultFeatures } from '@/api/hooks/category-default-feature';
import { fetchListCategories } from '@/api/hooks/category';
import { PaginationQuery } from '@/core';
import type { thingmodelservicev1_CategoryDefaultFeature } from '@/api/generated/admin/service/v1';

interface Props {
  open: boolean;
  onClose: (reload?: boolean) => void;
}

const CreateProductWizard = ({ open, onClose }: Props) => {
  const { t } = useTranslation(['product', 'common']);
  const { message } = App.useApp();
  const [step, setStep] = useState(0);
  const [form] = Form.useForm();
  const [productId, setProductId] = useState<number | null>(null);
  const [defaults, setDefaults] = useState<thingmodelservicev1_CategoryDefaultFeature[]>([]);
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [onConflict, setOnConflict] = useState<'SKIP' | 'REPLACE'>('SKIP');
  const [loadingDefaults, setLoadingDefaults] = useState(false);
  const [categoryOptions, setCategoryOptions] = useState<any[]>([]);

  const { mutate: doCreate, isPending: creating } = useCreateProduct();
  const { mutate: doPull, isPending: pulling } = usePullFromDefault();

  // 进入向导时拉一次完整分类树
  useEffect(() => {
    if (!open) return;
    (async () => {
      const q = new PaginationQuery({ page: 1, pageSize: 1000 });
      q.query = { kind: { value: 'FACILITY', condition: 'equal' } };
      const resp = await fetchListCategories(q);
      // 简化：按 code 排序的扁平列表，前端构建级联（kind=FACILITY 一棵树）
      setCategoryOptions(buildCascade(resp.items ?? []));
    })();
  }, [open]);

  const handleNext = async () => {
    try {
      const v = await form.validateFields();
      doCreate(
        {
          data: {
            code: v.code,
            name: v.name,
            nameEn: v.nameEn,
            categoryId: Number(v.categoryId),
            manufacturer: v.manufacturer,
            modelNo: v.modelNo,
            description: v.description,
          },
        },
        {
          onSuccess: async () => {
            // 拿新产品 id（按 code 查）
            const p = await fetchGetProduct({ code: v.code });
            setProductId(p.id ?? null);
            // 拿该分类的默认条目
            setLoadingDefaults(true);
            const cdfQ = new PaginationQuery({ page: 1, pageSize: 500 });
            cdfQ.query = { category_id: { value: v.categoryId, condition: 'equal' } };
            const cdfResp = await fetchListCategoryDefaultFeatures(cdfQ);
            const items = cdfResp.items ?? [];
            setDefaults(items);
            setSelectedIds(items.map((x) => x.id!));
            setLoadingDefaults(false);
            setStep(1);
          },
          onError: (err) => message.error(err.message),
        },
      );
    } catch {
      /* form invalid */
    }
  };

  const handlePullAndFinish = () => {
    if (!productId) return;
    doPull(
      {
        productId,
        defaultFeatureIds: selectedIds,
        onConflict: onConflict as any,
      },
      {
        onSuccess: () => {
          message.success(t('createSuccess'));
          reset();
          onClose(true);
        },
        onError: (err) => message.error(err.message),
      },
    );
  };

  const handleSkipPull = () => {
    message.success(t('createSuccess'));
    reset();
    onClose(true);
  };

  const reset = () => {
    setStep(0);
    setProductId(null);
    setDefaults([]);
    setSelectedIds([]);
    form.resetFields();
  };

  return (
    <Modal
      open={open}
      title={t('createProduct')}
      width={760}
      onCancel={() => {
        reset();
        onClose(false);
      }}
      footer={null}
      destroyOnClose
    >
      <Steps current={step} items={[{ title: t('wizard.step1') }, { title: t('wizard.step2') }]} />
      <div style={{ marginTop: 24 }}>
        {step === 0 && (
          <Form form={form} layout="vertical">
            <Form.Item
              label={t('category')}
              name="categoryId"
              rules={[{ required: true, message: t('categoryRequired') }]}
            >
              <Cascader
                options={categoryOptions}
                placeholder={t('categoryPlaceholder')}
                showSearch
                changeOnSelect={false}
                displayRender={(labels) => labels.join(' / ')}
                onChange={(val) => {
                  // 取叶子（最深一层）作为 categoryId
                  if (val?.length) form.setFieldValue('categoryId', val[val.length - 1]);
                }}
              />
            </Form.Item>
            <Form.Item label={t('code')} name="code" rules={[{ required: true }]}>
              <Input placeholder="GREE-LSBLG320" />
            </Form.Item>
            <Form.Item label={t('name')} name="name" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item label={t('nameEn')} name="nameEn">
              <Input />
            </Form.Item>
            <Form.Item label={t('manufacturer')} name="manufacturer">
              <Input />
            </Form.Item>
            <Form.Item label={t('modelNo')} name="modelNo">
              <Input />
            </Form.Item>
            <Form.Item label={t('description')} name="description">
              <Input.TextArea rows={2} />
            </Form.Item>
            <div style={{ textAlign: 'right' }}>
              <button
                type="button"
                className="ant-btn ant-btn-primary"
                onClick={handleNext}
                disabled={creating}
              >
                {t('common:next')}
              </button>
            </div>
          </Form>
        )}
        {step === 1 && (
          <Spin spinning={loadingDefaults}>
            <p>{t('wizard.pullDescription', '勾选要从默认模型拷贝的特征（默认全选）')}</p>
            <Checkbox.Group
              value={selectedIds}
              onChange={(v) => setSelectedIds(v as number[])}
              style={{ display: 'flex', flexDirection: 'column', gap: 8, maxHeight: 320, overflowY: 'auto' }}
            >
              {defaults.map((d) => (
                <Checkbox key={d.id} value={d.id}>
                  <strong>{d.featureCode}</strong> {d.featureName}
                  {d.overrideSpec && (
                    <span style={{ marginLeft: 8, color: '#1677ff' }}>{t('hasOverride')}</span>
                  )}
                </Checkbox>
              ))}
            </Checkbox.Group>
            <div style={{ marginTop: 16 }}>
              <span>{t('wizard.onConflict')}：</span>
              <Radio.Group value={onConflict} onChange={(e) => setOnConflict(e.target.value)}>
                <Radio value="SKIP">{t('wizard.skip')}</Radio>
                <Radio value="REPLACE">{t('wizard.replace')}</Radio>
              </Radio.Group>
            </div>
            <div style={{ textAlign: 'right', marginTop: 16 }}>
              <button type="button" className="ant-btn" onClick={handleSkipPull} disabled={pulling}>
                {t('wizard.skipPull')}
              </button>
              <button
                type="button"
                className="ant-btn ant-btn-primary"
                style={{ marginLeft: 8 }}
                onClick={handlePullAndFinish}
                disabled={pulling}
              >
                {t('wizard.createAndPull')}
              </button>
            </div>
          </Spin>
        )}
      </div>
    </Modal>
  );
};

/**
 * 按 code 前缀构建 FACILITY 分类树（level 1→2→3→4）。
 */
function buildCascade(items: Array<{ id?: number; code?: string; level?: number; name?: string; parentId?: number }>) {
  type Node = { value: number; label: string; children?: Node[] };
  const byId = new Map<number, Node>();
  items.forEach((it) => {
    byId.set(it.id!, { value: it.id!, label: `${it.code} ${it.name ?? ''}` });
  });
  const roots: Node[] = [];
  items.forEach((it) => {
    const n = byId.get(it.id!)!;
    const parent = it.parentId ? byId.get(it.parentId) : undefined;
    if (parent) {
      (parent.children ??= []).push(n);
    } else {
      roots.push(n);
    }
  });
  return roots;
}

export default CreateProductWizard;
```

- [ ] **Step 2: pnpm typecheck 通过 + Commit**

```bash
git add frontend/admin/react/src/pages/app/thingmodel/product/CreateProductWizard.tsx
git commit -m "feat(react): impl two-step CreateProductWizard with PullFromDefault"
```

---

## Task 10: 入口② · 产品详情页

**Files:**
- Create: `frontend/admin/react/src/pages/app/thingmodel/product/ProductDetailPage.tsx`

- [ ] **Step 1: 实现详情页（顶部信息 + Tab 化模型编辑器）**

```tsx
/**
 * 产品详情页 / Product detail page.
 *
 * - 顶部：基本信息卡片 + 发布/取消发布/删除按钮
 * - 主体：Tab(属性/事件/服务/关系) 列表 + "+ 添加特征" 下拉（拉取默认 / 加全局 / 加本地）
 */
import { useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Button, Card, Descriptions, Tabs, Space, Popconfirm, App, Modal, Tag, Dropdown } from 'antd';
import { ProTable, type ActionType, type ProColumns } from '@ant-design/pro-components';
import { useTranslation } from 'react-i18next';

import ContentContainer from '@/layouts/components/PageContainer/ContentContainer';
import { useGetProduct, usePublishProduct, useUnpublishProduct, useDeleteProduct } from '@/api/hooks/product';
import {
  fetchListProductFeatures,
  useDeleteProductFeature,
  usePullFromDefault,
} from '@/api/hooks/product-feature';
import { fetchListCategoryDefaultFeatures } from '@/api/hooks/category-default-feature';
import { PaginationQuery } from '@/core';
import type { thingmodelservicev1_ProductFeature } from '@/api/generated/admin/service/v1';
import FeatureSourceTag from '../_shared/FeatureSourceTag';
import ProductFeatureDrawer from './ProductFeatureDrawer';

type Tab = 'PROPERTY' | 'EVENT' | 'SERVICE' | 'RELATION';

const ProductDetailPage = () => {
  const { id } = useParams();
  const productId = Number(id);
  const { t } = useTranslation(['product', 'product-feature', 'common']);
  const navigate = useNavigate();
  const actionRef = useRef<ActionType>(null);
  const { message } = App.useApp();
  const [tab, setTab] = useState<Tab>('PROPERTY');
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [drawerMode, setDrawerMode] = useState<'edit' | 'create-local' | 'create-global'>('edit');
  const [editing, setEditing] = useState<thingmodelservicev1_ProductFeature | null>(null);

  const { data: product, refetch: refetchProduct } = useGetProduct({ id: productId });
  const { mutate: doPublish } = usePublishProduct();
  const { mutate: doUnpublish } = useUnpublishProduct();
  const { mutate: doDeleteProduct } = useDeleteProduct();
  const { mutate: doDeletePF } = useDeleteProductFeature({
    onSuccess: () => {
      message.success(t('common:deleteSuccess'));
      actionRef.current?.reload();
    },
  });
  const { mutate: doPull } = usePullFromDefault({
    onSuccess: () => {
      message.success(t('product:pullSuccess'));
      actionRef.current?.reload();
    },
  });

  const published = product?.status === 'PUBLISHED';

  const handlePublish = () => {
    Modal.confirm({
      title: t('product:confirmPublish'),
      content: t('product:publishHint'),
      onOk: () =>
        doPublish(
          { id: productId },
          {
            onSuccess: () => {
              message.success(t('product:publishSuccess'));
              refetchProduct();
            },
          },
        ),
    });
  };
  const handleUnpublish = () =>
    doUnpublish({ id: productId }, { onSuccess: () => refetchProduct() });
  const handleDelete = () =>
    doDeleteProduct(
      { ids: [productId] },
      {
        onSuccess: () => {
          message.success(t('common:deleteSuccess'));
          navigate('/thingmodel/product');
        },
      },
    );

  const handlePullMore = async () => {
    if (!product?.categoryId) return;
    const q = new PaginationQuery({ page: 1, pageSize: 500 });
    q.query = { category_id: { value: product.categoryId, condition: 'equal' } };
    const resp = await fetchListCategoryDefaultFeatures(q);
    const ids = (resp.items ?? []).map((x) => x.id!);
    if (ids.length === 0) {
      message.info(t('product:noDefaultModel'));
      return;
    }
    doPull({ productId, defaultFeatureIds: ids, onConflict: 'SKIP' as any });
  };

  const columns: ProColumns<thingmodelservicev1_ProductFeature>[] = [
    {
      title: t('product-feature:source'),
      dataIndex: 'source',
      width: 80,
      render: (_, row) => <FeatureSourceTag source={row.source as any} />,
    },
    { title: t('product-feature:code'), dataIndex: 'code', width: 160, copyable: true },
    { title: t('product-feature:identifier'), dataIndex: 'identifier', width: 160 },
    { title: t('product-feature:name'), dataIndex: 'name' },
    { title: t('product-feature:dataType'), dataIndex: 'dataType', width: 100,
      render: (v: any) => (v ? <Tag>{v}</Tag> : null) },
    { title: t('product-feature:accessMode'), dataIndex: 'accessMode', width: 90 },
    {
      title: t('product-feature:overridden'),
      width: 90,
      render: (_, row) => (row.overrideSpec ? t('product-feature:yes') : '-'),
    },
    {
      title: t('common:operation'),
      width: 160,
      render: (_, row) => (
        <Space size="small">
          <Button
            size="small"
            type="link"
            onClick={() => {
              setEditing(row);
              setDrawerMode('edit');
              setDrawerOpen(true);
            }}
          >
            {t('common:edit')}
          </Button>
          {!published && (
            <Popconfirm title={t('common:confirmDelete')} onConfirm={() => doDeletePF({ ids: [row.id!] })}>
              <Button size="small" type="link" danger>
                {t('common:delete')}
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  if (!product) return null;

  return (
    <ContentContainer heightMode="fixed" padding="16px" bottomMargin={0}>
      <Space style={{ marginBottom: 12 }}>
        <Button onClick={() => navigate('/thingmodel/product')}>← {t('common:back')}</Button>
        <span style={{ fontSize: 18, fontWeight: 600 }}>{product.name}</span>
        {published ? (
          <Tag color="green">{t('product:status.PUBLISHED')}</Tag>
        ) : (
          <Tag>{t('product:status.DRAFT')}</Tag>
        )}
      </Space>
      <Space style={{ float: 'right' }}>
        {!published && (
          <Button type="primary" onClick={handlePublish}>
            {t('product:publish')}
          </Button>
        )}
        {published && <Button onClick={handleUnpublish}>{t('product:unpublish')}</Button>}
        <Popconfirm title={t('common:confirmDelete')} onConfirm={handleDelete}>
          <Button danger>{t('common:delete')}</Button>
        </Popconfirm>
      </Space>
      <Card style={{ marginBottom: 12 }}>
        <Descriptions column={2} size="small">
          <Descriptions.Item label={t('product:code')}>{product.code}</Descriptions.Item>
          <Descriptions.Item label={t('product:category')}>
            {product.categoryName} ({product.categoryCode})
          </Descriptions.Item>
          <Descriptions.Item label={t('product:manufacturer')}>
            {product.manufacturer || '-'}
          </Descriptions.Item>
          <Descriptions.Item label={t('product:modelNo')}>{product.modelNo || '-'}</Descriptions.Item>
          <Descriptions.Item label={t('product:description')} span={2}>
            {product.description || '-'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Tabs activeKey={tab} onChange={(k) => setTab(k as Tab)}
        items={(['PROPERTY', 'EVENT', 'SERVICE', 'RELATION'] as Tab[]).map((tp) => ({
          key: tp, label: t(`product-feature:type.${tp}`),
        }))}
      />

      <ProTable
        actionRef={actionRef}
        rowKey="id"
        search={false}
        columns={columns}
        toolBarRender={() => [
          <Dropdown
            key="add"
            menu={{
              items: [
                { key: 'pull', label: t('product:addFrom.pull'), onClick: handlePullMore },
                {
                  key: 'global',
                  label: t('product:addFrom.global'),
                  onClick: () => {
                    setEditing(null);
                    setDrawerMode('create-global');
                    setDrawerOpen(true);
                  },
                },
                {
                  key: 'local',
                  label: t('product:addFrom.local'),
                  onClick: () => {
                    setEditing(null);
                    setDrawerMode('create-local');
                    setDrawerOpen(true);
                  },
                },
              ],
            }}
          >
            <Button type="primary" disabled={published}>
              + {t('product:addFeature')}
            </Button>
          </Dropdown>,
        ]}
        request={async (params) => {
          const q = new PaginationQuery({
            page: params.current ?? 1,
            pageSize: params.pageSize ?? 50,
          });
          q.query = {
            product_id: { value: productId, condition: 'equal' },
            feature_type: { value: tab, condition: 'equal' },
          };
          const resp = await fetchListProductFeatures(q);
          return { data: resp.items ?? [], success: true, total: Number(resp.total ?? 0) };
        }}
      />

      <ProductFeatureDrawer
        open={drawerOpen}
        productId={productId}
        feature={editing}
        mode={drawerMode}
        readonly={published && drawerMode === 'edit' ? 'partial' : false}
        onClose={(reload) => {
          setDrawerOpen(false);
          if (reload) actionRef.current?.reload();
        }}
      />
    </ContentContainer>
  );
};

export default ProductDetailPage;
```

- [ ] **Step 2: pnpm typecheck 通过（ProductFeatureDrawer 先 stub）**

临时 stub `ProductFeatureDrawer.tsx`：

```tsx
import { Drawer } from 'antd';
interface Props {
  open: boolean;
  productId: number;
  feature: any;
  mode: 'edit' | 'create-local' | 'create-global';
  readonly?: 'partial' | false;
  onClose: (reload?: boolean) => void;
}
const ProductFeatureDrawer = ({ open, onClose }: Props) => (
  <Drawer open={open} onClose={() => onClose(false)} title="编辑特征 (placeholder)">
    placeholder - 待 Task 11 实现
  </Drawer>
);
export default ProductFeatureDrawer;
```

- [ ] **Step 3: Commit**

```bash
git add frontend/admin/react/src/pages/app/thingmodel/product/ProductDetailPage.tsx \
        frontend/admin/react/src/pages/app/thingmodel/product/ProductFeatureDrawer.tsx
git commit -m "feat(react): add ProductDetailPage with feature tabs and lifecycle actions"
```

---

## Task 11: 入口② · 产品特征编辑 Drawer

**Files:**
- Modify: `frontend/admin/react/src/pages/app/thingmodel/product/ProductFeatureDrawer.tsx`（替换 stub）

- [ ] **Step 1: 实现 Drawer**

```tsx
/**
 * 产品特征编辑 Drawer（edit / create-global / create-local）。
 *
 * - edit：基础字段灰显（仅白名单可改）；PUBLISHED 产品 readonly=partial
 * - create-global：弹 FeaturePicker 选 1 条全局特征，落库 source=GLOBAL
 * - create-local：完全自定义 + 选 dataType（BOOL/INT/DOUBLE/TEXT），落库 source=LOCAL
 */
import { useEffect, useMemo, useState } from 'react';
import { Drawer, Form, Input, Select, Alert, Button, App } from 'antd';
import { useTranslation } from 'react-i18next';
import OverrideSpecForm from '../_shared/OverrideSpecForm';
import FeaturePicker from '../_shared/FeaturePicker';
import {
  useCreateProductFeature,
  useUpdateProductFeature,
} from '@/api/hooks/product-feature';
import type {
  thingmodelservicev1_FeatureOverrideSpec,
  thingmodelservicev1_ProductFeature,
} from '@/api/generated/admin/service/v1';

interface Props {
  open: boolean;
  productId: number;
  feature: thingmodelservicev1_ProductFeature | null;
  mode: 'edit' | 'create-local' | 'create-global';
  readonly?: 'partial' | false;
  onClose: (reload?: boolean) => void;
}

const ProductFeatureDrawer = ({ open, productId, feature, mode, readonly, onClose }: Props) => {
  const { t } = useTranslation(['product-feature', 'common']);
  const { message } = App.useApp();
  const [form] = Form.useForm();
  const [override, setOverride] = useState<thingmodelservicev1_FeatureOverrideSpec | null>(null);
  const [pickerOpen, setPickerOpen] = useState(mode === 'create-global');

  const { mutate: doCreate, isPending: creating } = useCreateProductFeature();
  const { mutate: doUpdate, isPending: updating } = useUpdateProductFeature();

  useEffect(() => {
    if (mode === 'edit' && feature) {
      form.setFieldsValue({
        code: feature.code,
        identifier: feature.identifier,
        name: feature.name,
        description: feature.description,
      });
      setOverride(feature.overrideSpec ?? null);
    } else if (mode === 'create-local') {
      form.resetFields();
      setOverride(null);
    } else if (mode === 'create-global') {
      setPickerOpen(true);
    }
  }, [feature, mode, form]);

  const handleSubmit = async () => {
    const v = await form.validateFields();
    if (mode === 'edit' && feature) {
      doUpdate(
        {
          id: feature.id!,
          data: { ...v, overrideSpec: override ?? undefined },
          mask: ['name', 'description', 'override_spec'],
        },
        {
          onSuccess: () => {
            message.success(t('common:updateSuccess'));
            onClose(true);
          },
        },
      );
      return;
    }
    if (mode === 'create-local') {
      doCreate(
        {
          data: {
            productId,
            source: 'LOCAL' as any,
            featureType: 'PROPERTY' as any,
            code: v.code,
            identifier: v.identifier,
            name: v.name,
            description: v.description,
            featureSnapshot: {
              spec: {
                property: {
                  dataType: v.dataType,
                  accessMode: v.accessMode ?? 'RW',
                },
              },
            },
            dataType: v.dataType,
            accessMode: v.accessMode ?? 'RW',
          },
        },
        {
          onSuccess: () => {
            message.success(t('common:createSuccess'));
            onClose(true);
          },
        },
      );
    }
  };

  const isEditing = mode === 'edit';
  const isPropertyLocal = mode === 'create-local';

  return (
    <>
      <Drawer
        width={720}
        open={open && (mode !== 'create-global' || !!feature)}
        onClose={() => onClose(false)}
        title={
          isEditing
            ? `${t('editFeature')} — ${feature?.name}`
            : mode === 'create-local'
            ? t('createLocalFeature')
            : t('createGlobalFeature')
        }
        destroyOnClose
        extra={
          <Button type="primary" onClick={handleSubmit} loading={creating || updating}>
            {t('common:save')}
          </Button>
        }
      >
        {readonly === 'partial' && (
          <Alert
            type="warning"
            message={t('publishedReadonlyWarning')}
            showIcon
            style={{ marginBottom: 12 }}
          />
        )}
        <Form form={form} layout="vertical" disabled={readonly === 'partial' && isEditing}>
          <Form.Item label={t('code')} name="code" rules={[{ required: !isEditing }]}>
            <Input disabled={isEditing} />
          </Form.Item>
          <Form.Item label={t('identifier')} name="identifier" rules={[{ required: !isEditing }]}>
            <Input disabled={isEditing} />
          </Form.Item>
          <Form.Item label={t('name')} name="name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label={t('description')} name="description">
            <Input.TextArea rows={2} />
          </Form.Item>

          {isPropertyLocal && (
            <>
              <Form.Item label={t('dataType')} name="dataType" rules={[{ required: true }]}>
                <Select
                  options={['INT', 'DOUBLE', 'BOOL', 'TEXT', 'ENUM', 'DATE'].map((v) => ({
                    label: v,
                    value: v,
                  }))}
                />
              </Form.Item>
              <Form.Item label={t('accessMode')} name="accessMode" initialValue="RW">
                <Select options={[{ label: 'R', value: 'R' }, { label: 'RW', value: 'RW' }]} />
              </Form.Item>
            </>
          )}
        </Form>

        {(isEditing || isPropertyLocal) && (
          <>
            <h4>{t('overrideSection')}</h4>
            <OverrideSpecForm
              featureType={feature?.featureType ?? ('PROPERTY' as any)}
              snapshot={feature?.featureSnapshot}
              value={override}
              onChange={setOverride}
              readonly={readonly === 'partial' ? false : undefined}
            />
          </>
        )}
      </Drawer>

      {mode === 'create-global' && (
        <FeaturePicker
          open={pickerOpen}
          onCancel={() => {
            setPickerOpen(false);
            onClose(false);
          }}
          onConfirm={(feats) => {
            setPickerOpen(false);
            if (feats.length === 0) {
              onClose(false);
              return;
            }
            // 多选时逐个创建
            let done = 0;
            feats.forEach((f) => {
              doCreate(
                {
                  data: {
                    productId,
                    source: 'GLOBAL' as any,
                    refFeatureId: f.id,
                  },
                },
                {
                  onSuccess: () => {
                    done++;
                    if (done === feats.length) {
                      message.success(t('common:createSuccess'));
                      onClose(true);
                    }
                  },
                  onError: (err) => message.error(err.message),
                },
              );
            });
          }}
        />
      )}
    </>
  );
};

export default ProductFeatureDrawer;
```

- [ ] **Step 2: pnpm typecheck 通过 + Commit**

```bash
git add frontend/admin/react/src/pages/app/thingmodel/product/ProductFeatureDrawer.tsx
git commit -m "feat(react): impl ProductFeatureDrawer (edit/create-local/create-global)"
```

---

## Task 12: 路由 + i18n

**Files:**
- Modify: `frontend/admin/react/src/router/modules/thingmodel.tsx`
- Modify: `frontend/admin/react/src/locales/zh-CN/_core/routes.json`
- Modify: `frontend/admin/react/src/locales/en-US/_core/routes.json`
- Create: `frontend/admin/react/src/locales/{zh-CN,en-US}/_modules/product.json`
- Create: `frontend/admin/react/src/locales/{zh-CN,en-US}/_modules/product-feature.json`
- Create: `frontend/admin/react/src/locales/{zh-CN,en-US}/_modules/category-default-feature.json`

- [ ] **Step 1: 改 thingmodel.tsx，追加 product 路由（含 :id）**

```tsx
{
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
  name: 'thingmodel-product-detail',
  path: 'product/:id',
  element: createLazyRoute(() => import('@/pages/app/thingmodel/product/ProductDetailPage')),
  meta: {
    title: 'routes:thingmodelProduct',
    hideInMenu: true,
    activePath: '/thingmodel/product',
  },
},
```

- [ ] **Step 2: routes.json 追加 thingmodelProduct key（zh + en）**

```json
"thingmodelProduct": "产品管理"
// en: "thingmodelProduct": "Product"
```

- [ ] **Step 3: 写 3 份 _modules JSON（中英镜像）**

参考 feature.json 风格，每份 zh 文件至少含 `pageTitle / moduleName / pull*/wizard.* / source.* / addFrom.* / override.* / status.* / tab.*` 等业务 key。完整 key 列表在设计文档 [05 §4](../../thingmodel/sheji/模型管理/05-前端实现设计.md)。

- [ ] **Step 4: pnpm typecheck + i18n key 覆盖检查**

Run: `cd frontend/admin/react && pnpm typecheck && pnpm lint --quiet 2>&1 | tail -10`
Expected: 0 errors, 0 lint warnings tied to thingmodel/* files

- [ ] **Step 5: Commit**

```bash
git add frontend/admin/react/src/router/modules/thingmodel.tsx \
        frontend/admin/react/src/locales/
git commit -m "feat(react): add product routes & 3 i18n module files (zh-CN + en-US)"
```

---

## Task 13: 全量验收 (build + 接口巡检 + 手测)

- [ ] **Step 1: pnpm build 通过**

Run: `cd frontend/admin/react && pnpm build 2>&1 | tail -10`
Expected: built successfully

- [ ] **Step 2: 启动 dev 服务器手测核心路径（后端已起 + 种子已入库）**

Run: `cd frontend/admin/react && pnpm dev`

打开 `http://localhost:7000` → 登录 admin/admin。

| # | 路径 | 期望 |
|---|------|------|
| 1 | 进入"物模型 / 产品管理" | 路由存在，列表显示种子产品 GREE-LSBLG320 (PUBLISHED) |
| 2 | 点击产品行 → 详情 | 4 个 Tab (属性/事件/服务/关系)；属性页 5 条（含 LOCAL 夜间静音） |
| 3 | PUBLISHED 标记可见、"添加"按钮置灰 | ✓ |
| 4 | 点击行内"编辑"，打开 Drawer，warning 黄条提示 PUBLISHED | ✓ |
| 5 | 退回列表 → "+ 新增产品" 打开向导 Step 1 → 选 FACILITY 细类 | ✓ |
| 6 | 填基本信息 → Step 2 默认勾选 10 条 → 创建 | 列表多出新产品 |
| 7 | 进入新产品 → 列表显示 10 条 DEFAULT 特征 | ✓ |
| 8 | "+ 添加特征 → 新建产品本地特征"，填 BOOL 字段，保存 | LOCAL 行出现 |
| 9 | 编辑某 property 行，改 constraints.max=99，保存 | "覆写"列变"已" |
| 10 | 删除新产品 | 列表删除 |
| 11 | 回到分类管理 → 找 20010100 行 → "配置默认模型" | Drawer 打开，10 条种子条目 |
| 12 | 在 Drawer 中编辑某条 override，改 displayName，保存 | UI 列表刷新 |

- [ ] **Step 3: 收尾 commit（如有手测发现的小修补）**

```bash
git status   # 应 clean，或仅微调
```

---

## 完成标准

实施完成 = 满足以下全部：

1. ✅ Tasks 1-13 全部 checkbox 完成
2. ✅ `pnpm typecheck` 0 errors / `pnpm build` 成功
3. ✅ 路由 `/thingmodel/product` 与 `/thingmodel/product/:id` 可访问
4. ✅ 13 条核心手测路径全部通过
5. ✅ Vue Element / Vue Vben 两个目录**未做任何变更**（`git diff` 仅 react/）
6. ✅ 后端 22 个接口在前端 fully covered (4 hooks 文件全部走 BFF)

---

## 风险与缓解

| 风险 | 缓解 |
|------|------|
| BFF 生成 TS 类型枚举值序列化（"PROPERTY" 字符串 vs 数字）不一致 | apiClient 已统一；hooks 透传，遇到类型断言用 `as any` 收窄；运行时 dev 工具验证 |
| Cascader 全分类加载性能 | 限制 kind=FACILITY + pageSize=1000；如未来 >1000 条改 lazyLoad |
| 中英 i18n key 漏配 | Task 12 Step 4 lint 检查；项目 i18n 缺 key 时 fallback 为 key 本身，不阻断 |
| PUBLISHED 产品的 Drawer readonly UX | Task 11 Step 1 顶部黄条 + Form `disabled` 显式提示 |
| FeaturePicker 多选后逐个 Create 时部分失败 | Task 11 Step 1 done 计数 + 错误聚合；本期不做事务回滚 |

---

## 不在本计划内

- vue-element / vue-vben 前端 — 用户明确仅 React
- 拖拽排序 UI（用列表 `sort_order` 字段即可，体验优化留到下期）
- 产品克隆 UI（CloneFromProduct RPC 已就位，前端不暴露按钮）
- 产品级 simulator / 联调 — 本期不做
- E2E 自动化测试 — 项目现有 React 端无 E2E 框架，手测为主
