/**
 * 物模型 - 分类默认模型条目 hooks。
 * Category default feature management hooks.
 *
 * 镜像 `feature.ts` 的 React Query 整合，封装 7 个 BFF RPC：
 * List/Get/Create/BatchAdd/Update/Delete/Reorder。
 *
 * 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/05-前端实现设计.md §2.1
 */
import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from '@tanstack/react-query';
import {
  type thingmodelservicev1_CategoryDefaultFeature,
  type thingmodelservicev1_ListCategoryDefaultFeatureResponse,
  type thingmodelservicev1_GetCategoryDefaultFeatureRequest,
  type thingmodelservicev1_CreateCategoryDefaultFeatureRequest,
  type thingmodelservicev1_BatchAddCategoryDefaultFeaturesRequest,
  type thingmodelservicev1_BatchAddCategoryDefaultFeaturesResponse,
  type thingmodelservicev1_DeleteCategoryDefaultFeatureRequest,
  type thingmodelservicev1_ReorderCategoryDefaultFeaturesRequest,
} from '@/api/generated/admin/service/v1';
import { makeUpdateMask, type PaginationQuery } from '@/core/transport/rest';
import { queryClient } from '@/core';
import { apiClient } from '@/api/client';

// ==============================
// 分页 / 列表
// ==============================

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

export async function fetchListCategoryDefaultFeatures(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ['listCategoryDefaultFeatures', params],
    queryFn: () => apiClient.categoryDefaultFeatureService.List(params.toRawParams()),
    retry: 0,
  });
}

// ==============================
// 详情
// ==============================

export function useGetCategoryDefaultFeature(
  req: thingmodelservicev1_GetCategoryDefaultFeatureRequest,
  options?: UseQueryOptions<thingmodelservicev1_CategoryDefaultFeature, Error>,
) {
  return useQuery({
    queryKey: ['getCategoryDefaultFeature', req],
    queryFn: () => apiClient.categoryDefaultFeatureService.Get(req),
    enabled: Boolean(req.id),
    ...options,
  });
}

// ==============================
// 增 / 改 / 删 / 批量 / 排序
// ==============================

export function useCreateCategoryDefaultFeature(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_CreateCategoryDefaultFeatureRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.categoryDefaultFeatureService.Create(data),
    ...options,
  });
}

/**
 * 批量添加：一次性把多个 feature 绑定到 category。
 * 后端会跳过已存在的 (category, feature) 对，返回 skipped_duplicate_feature_codes。
 */
export function useBatchAddCategoryDefaultFeatures(
  options?: UseMutationOptions<
    thingmodelservicev1_BatchAddCategoryDefaultFeaturesResponse,
    Error,
    thingmodelservicev1_BatchAddCategoryDefaultFeaturesRequest
  >,
) {
  return useMutation({
    mutationFn: (data) => apiClient.categoryDefaultFeatureService.BatchAdd(data),
    ...options,
  });
}

/**
 * 更新（FieldMask）：传 values（含变更字段），按 keys 派生 updateMask。
 * 调用方常用：useUpdateCategoryDefaultFeature().mutate({ id, values: { overrideSpec: ... } })
 */
export function useUpdateCategoryDefaultFeature(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>,
) {
  return useMutation({
    mutationFn: ({ id, values }) =>
      apiClient.categoryDefaultFeatureService.Update({
        id,
        data: { ...values } as any,
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteCategoryDefaultFeature(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_DeleteCategoryDefaultFeatureRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.categoryDefaultFeatureService.Delete(data),
    ...options,
  });
}

/**
 * 拖拽排序：一次性提交多条 (id, sort_order)。
 */
export function useReorderCategoryDefaultFeatures(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_ReorderCategoryDefaultFeaturesRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.categoryDefaultFeatureService.Reorder(data),
    ...options,
  });
}
