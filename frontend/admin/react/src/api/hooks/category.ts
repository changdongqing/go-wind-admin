/**
 * 物模型-分类管理 hooks（CRUD）
 * Thing-model category management hooks (CRUD)
 *
 * 设计依据 / Design ref: docs/thingmodel/sheji/分类管理/05-前端实现设计.md
 * 镜像 `unit.ts` 中 UnitCategory 部分的 React Query 整合。
 */
import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from '@tanstack/react-query';
import {
  type thingmodelservicev1_Category,
  type thingmodelservicev1_ListCategoryResponse,
  type thingmodelservicev1_GetCategoryRequest,
  type thingmodelservicev1_CreateCategoryRequest,
  type thingmodelservicev1_DeleteCategoryRequest,
} from '@/api/generated/admin/service/v1';
import { makeUpdateMask, type PaginationQuery } from '@/core/transport/rest';
import { queryClient } from '@/core';
import { apiClient } from '@/api/client';

// ==============================
// 分类（Category）
// ==============================

export function useListCategories(
  query: PaginationQuery,
  options?: UseQueryOptions<thingmodelservicev1_ListCategoryResponse, Error>,
) {
  return useQuery({
    queryKey: ['listCategories', query],
    queryFn: () => apiClient.categoryService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListCategories(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ['listCategories', params],
    queryFn: () => apiClient.categoryService.List(params.toRawParams()),
    retry: 0,
  });
}

export function useGetCategory(
  req: thingmodelservicev1_GetCategoryRequest,
  options?: UseQueryOptions<thingmodelservicev1_Category, Error>,
) {
  return useQuery({
    queryKey: ['getCategory', req],
    queryFn: () => apiClient.categoryService.Get(req),
    ...options,
  });
}

export function useCreateCategory(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_CreateCategoryRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.categoryService.Create(data),
    ...options,
  });
}

export function useUpdateCategory(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>,
) {
  return useMutation({
    mutationFn: ({ id, values }) =>
      apiClient.categoryService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteCategory(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_DeleteCategoryRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.categoryService.Delete(data),
    ...options,
  });
}
