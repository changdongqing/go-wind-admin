/**
 * 物模型 - 产品下特征条目 hooks。
 * Product feature hooks.
 *
 * 封装 8 个 BFF RPC：List/Get/Create/PullFromDefault/CloneFromProduct/Update/Delete/Reorder。
 *
 * 关键 RPC：
 *   - PullFromDefault：批量从分类默认模型拷贝到产品（值复制）
 *   - CloneFromProduct：从另一产品克隆全部特征（含 LOCAL）
 *
 * Get 返回的 effective_spec 由后端合并 feature_snapshot + override_spec 后填充。
 * 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/05-前端实现设计.md §2.1
 */
import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from '@tanstack/react-query';
import {
  type thingmodelservicev1_ProductFeature,
  type thingmodelservicev1_ListProductFeatureResponse,
  type thingmodelservicev1_GetProductFeatureRequest,
  type thingmodelservicev1_CreateProductFeatureRequest,
  type thingmodelservicev1_PullFromDefaultRequest,
  type thingmodelservicev1_PullFromDefaultResponse,
  type thingmodelservicev1_CloneFromProductRequest,
  type thingmodelservicev1_CloneFromProductResponse,
  type thingmodelservicev1_DeleteProductFeatureRequest,
  type thingmodelservicev1_ReorderProductFeaturesRequest,
} from '@/api/generated/admin/service/v1';
import { makeUpdateMask, type PaginationQuery } from '@/core/transport/rest';
import { queryClient } from '@/core';
import { apiClient } from '@/api/client';

// ==============================
// 分页 / 列表
// ==============================

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

export async function fetchListProductFeatures(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ['listProductFeatures', params],
    queryFn: () => apiClient.productFeatureService.List(params.toRawParams()),
    retry: 0,
  });
}

// ==============================
// 详情
// ==============================

export function useGetProductFeature(
  req: thingmodelservicev1_GetProductFeatureRequest,
  options?: UseQueryOptions<thingmodelservicev1_ProductFeature, Error>,
) {
  return useQuery({
    queryKey: ['getProductFeature', req],
    queryFn: () => apiClient.productFeatureService.Get(req),
    enabled: Boolean(req.id),
    ...options,
  });
}

// ==============================
// 增 / 改 / 删
// ==============================

export function useCreateProductFeature(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_CreateProductFeatureRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.productFeatureService.Create(data),
    ...options,
  });
}

export function useUpdateProductFeature(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>,
) {
  return useMutation({
    mutationFn: ({ id, values }) =>
      apiClient.productFeatureService.Update({
        id,
        data: { ...values } as any,
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteProductFeature(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_DeleteProductFeatureRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.productFeatureService.Delete(data),
    ...options,
  });
}

// ==============================
// 批量从默认模型拉取 / 从另一产品克隆 / 拖拽排序
// ==============================

export function usePullFromDefault(
  options?: UseMutationOptions<
    thingmodelservicev1_PullFromDefaultResponse,
    Error,
    thingmodelservicev1_PullFromDefaultRequest
  >,
) {
  return useMutation({
    mutationFn: (data) => apiClient.productFeatureService.PullFromDefault(data),
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
    mutationFn: (data) => apiClient.productFeatureService.CloneFromProduct(data),
    ...options,
  });
}

export function useReorderProductFeatures(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_ReorderProductFeaturesRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.productFeatureService.Reorder(data),
    ...options,
  });
}
