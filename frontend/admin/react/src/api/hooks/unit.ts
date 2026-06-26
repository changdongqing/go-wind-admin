/**
 * 物模型-单位管理 hooks（物理量分类 + 单位 + 单位换算）
 * Thing-model unit management hooks (categories + units + conversion)
 *
 * 镜像 `dict.ts` 的 React Query 整合，新增 `useConvertUnit` 换算 mutation。
 */
import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from '@tanstack/react-query';
import {
  type thingmodelservicev1_UnitCategory,
  type thingmodelservicev1_ListUnitCategoryResponse,
  type thingmodelservicev1_GetUnitCategoryRequest,
  type thingmodelservicev1_CreateUnitCategoryRequest,
  type thingmodelservicev1_DeleteUnitCategoryRequest,
  type thingmodelservicev1_Unit,
  type thingmodelservicev1_ListUnitResponse,
  type thingmodelservicev1_GetUnitRequest,
  type thingmodelservicev1_CreateUnitRequest,
  type thingmodelservicev1_DeleteUnitRequest,
  type thingmodelservicev1_ListUnitByCategoryRequest,
  type thingmodelservicev1_ConvertUnitRequest,
  type thingmodelservicev1_ConvertUnitResponse,
} from '@/api/generated/admin/service/v1';
import { makeUpdateMask, type PaginationQuery } from '@/core/transport/rest';
import { queryClient } from '@/core';
import { apiClient } from '@/api/client';

// ==============================
// 物理量分类（UnitCategory）
// ==============================

export function useListUnitCategories(
  query: PaginationQuery,
  options?: UseQueryOptions<thingmodelservicev1_ListUnitCategoryResponse, Error>,
) {
  return useQuery({
    queryKey: ['listUnitCategories', query],
    queryFn: () => apiClient.unitCategoryService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListUnitCategories(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ['listUnitCategories', params],
    queryFn: () => apiClient.unitCategoryService.List(params.toRawParams()),
    retry: 0,
  });
}

export function useGetUnitCategory(
  req: thingmodelservicev1_GetUnitCategoryRequest,
  options?: UseQueryOptions<thingmodelservicev1_UnitCategory, Error>,
) {
  return useQuery({
    queryKey: ['getUnitCategory', req],
    queryFn: () => apiClient.unitCategoryService.Get(req),
    ...options,
  });
}

export function useCreateUnitCategory(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_CreateUnitCategoryRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.unitCategoryService.Create(data),
    ...options,
  });
}

export function useUpdateUnitCategory(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>,
) {
  return useMutation({
    mutationFn: ({ id, values }) =>
      apiClient.unitCategoryService.Update({
        id,
        data: { ...values },
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteUnitCategory(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_DeleteUnitCategoryRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.unitCategoryService.Delete(data),
    ...options,
  });
}

// ==============================
// 单位（Unit）
// ==============================

export function useListUnits(
  query: PaginationQuery,
  options?: UseQueryOptions<thingmodelservicev1_ListUnitResponse, Error>,
) {
  return useQuery({
    queryKey: ['listUnits', query],
    queryFn: () => apiClient.unitService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListUnits(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ['listUnits', params],
    queryFn: () => apiClient.unitService.List(params.toRawParams()),
    retry: 0,
  });
}

export function useGetUnit(
  req: thingmodelservicev1_GetUnitRequest,
  options?: UseQueryOptions<thingmodelservicev1_Unit, Error>,
) {
  return useQuery({
    queryKey: ['getUnit', req],
    queryFn: () => apiClient.unitService.Get(req),
    ...options,
  });
}

/**
 * 按物理量分类查询单位（不分页，供属性选单位下拉框/换算器使用）
 * List units by category (no paging) — for property selector / converter dropdowns.
 */
export function useListUnitsByCategory(
  req: thingmodelservicev1_ListUnitByCategoryRequest,
  options?: UseQueryOptions<thingmodelservicev1_ListUnitResponse, Error>,
) {
  return useQuery({
    queryKey: ['listUnitsByCategory', req],
    queryFn: () => apiClient.unitService.ListByCategory(req),
    enabled: Boolean(req.categoryId || req.categoryCode),
    ...options,
  });
}

export function useCreateUnit(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_CreateUnitRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.unitService.Create(data),
    ...options,
  });
}

export function useUpdateUnit(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>,
) {
  return useMutation({
    mutationFn: ({ id, values }) =>
      apiClient.unitService.Update({
        id,
        data: { ...values } as any,
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteUnit(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_DeleteUnitRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.unitService.Delete(data),
    ...options,
  });
}

// ==============================
// 单位换算（Convert）
// ==============================

export function useConvertUnit(
  options?: UseMutationOptions<
    thingmodelservicev1_ConvertUnitResponse,
    Error,
    thingmodelservicev1_ConvertUnitRequest
  >,
) {
  return useMutation({
    mutationFn: (data) => apiClient.unitService.Convert(data),
    ...options,
  });
}
