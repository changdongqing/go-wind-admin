/**
 * 物模型 - 产品管理 hooks。
 * Product management hooks.
 *
 * 封装 7 个 BFF RPC：List/Get/Create/Update/Delete/Publish/Unpublish。
 * 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/05-前端实现设计.md §2.1
 */
import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from '@tanstack/react-query';
import {
  type thingmodelservicev1_Product,
  type thingmodelservicev1_ListProductResponse,
  type thingmodelservicev1_GetProductRequest,
  type thingmodelservicev1_CreateProductRequest,
  type thingmodelservicev1_DeleteProductRequest,
  type thingmodelservicev1_PublishProductRequest,
  type thingmodelservicev1_UnpublishProductRequest,
} from '@/api/generated/admin/service/v1';
import { makeUpdateMask, type PaginationQuery } from '@/core/transport/rest';
import { queryClient } from '@/core';
import { apiClient } from '@/api/client';

// ==============================
// 分页 / 列表
// ==============================

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

export async function fetchListProducts(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ['listProducts', params],
    queryFn: () => apiClient.productService.List(params.toRawParams()),
    retry: 0,
  });
}

// ==============================
// 详情（支持 by id 或 by code）
// ==============================

export function useGetProduct(
  req: thingmodelservicev1_GetProductRequest,
  options?: UseQueryOptions<thingmodelservicev1_Product, Error>,
) {
  return useQuery({
    queryKey: ['getProduct', req],
    queryFn: () => apiClient.productService.Get(req),
    enabled: Boolean(req.id || req.code),
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

// ==============================
// 增 / 改 / 删
// ==============================

export function useCreateProduct(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_CreateProductRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.productService.Create(data),
    ...options,
  });
}

export function useUpdateProduct(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>,
) {
  return useMutation({
    mutationFn: ({ id, values }) =>
      apiClient.productService.Update({
        id,
        data: { ...values } as any,
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteProduct(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_DeleteProductRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.productService.Delete(data),
    ...options,
  });
}

// ==============================
// 生命周期：发布 / 撤销发布
// ==============================

export function usePublishProduct(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_PublishProductRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.productService.Publish(data),
    ...options,
  });
}

export function useUnpublishProduct(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_UnpublishProductRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.productService.Unpublish(data),
    ...options,
  });
}
