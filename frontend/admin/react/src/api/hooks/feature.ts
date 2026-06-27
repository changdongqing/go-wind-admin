/**
 * 物模型-特征管理 hooks（属性 / 事件 / 服务 / 关系 统一）
 * Thing-model feature management hooks (property/event/service/relation unified).
 *
 * 镜像 `unit.ts` 的 React Query 整合，新增按类型查询与 spec 校验 mutation。
 */
import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from '@tanstack/react-query';
import {
  type thingmodelservicev1_Feature,
  type thingmodelservicev1_ListFeatureResponse,
  type thingmodelservicev1_GetFeatureRequest,
  type thingmodelservicev1_CreateFeatureRequest,
  type thingmodelservicev1_DeleteFeatureRequest,
  type thingmodelservicev1_ListFeatureByTypeRequest,
  type thingmodelservicev1_ValidateFeatureSpecRequest,
  type thingmodelservicev1_ValidateFeatureSpecResponse,
  type thingmodelservicev1_ImportFeaturesRequest,
  type thingmodelservicev1_ImportFeaturesResponse,
  type thingmodelservicev1_ImportFeatureRow,
} from '@/api/generated/admin/service/v1';
import { makeUpdateMask, type PaginationQuery } from '@/core/transport/rest';
import { queryClient } from '@/core';
import { apiClient } from '@/api/client';

// ==============================
// 分页 / 列表
// ==============================

export function useListFeatures(
  query: PaginationQuery,
  options?: UseQueryOptions<thingmodelservicev1_ListFeatureResponse, Error>,
) {
  return useQuery({
    queryKey: ['listFeatures', query],
    queryFn: () => apiClient.featureService.List(query.toRawParams()),
    ...options,
  });
}

export async function fetchListFeatures(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ['listFeatures', params],
    queryFn: () => apiClient.featureService.List(params.toRawParams()),
    retry: 0,
  });
}

/**
 * 按特征类型查询（不分页，供左侧树联动 / 关系选择器使用）
 * List features by type (no paging).
 */
export function useListFeaturesByType(
  req: thingmodelservicev1_ListFeatureByTypeRequest,
  options?: UseQueryOptions<thingmodelservicev1_ListFeatureResponse, Error>,
) {
  return useQuery({
    queryKey: ['listFeaturesByType', req],
    queryFn: () => apiClient.featureService.ListByType(req),
    enabled: Boolean(req.featureType),
    ...options,
  });
}

// ==============================
// 详情
// ==============================

export function useGetFeature(
  req: thingmodelservicev1_GetFeatureRequest,
  options?: UseQueryOptions<thingmodelservicev1_Feature, Error>,
) {
  return useQuery({
    queryKey: ['getFeature', req],
    queryFn: () => apiClient.featureService.Get(req),
    enabled: Boolean(req.id || req.code || req.identifier),
    ...options,
  });
}

// ==============================
// 增 / 改 / 删
// ==============================

export function useCreateFeature(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_CreateFeatureRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.featureService.Create(data),
    ...options,
  });
}

export function useUpdateFeature(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>,
) {
  return useMutation({
    mutationFn: ({ id, values }) =>
      apiClient.featureService.Update({
        id,
        data: { ...values } as any,
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteFeature(
  options?: UseMutationOptions<{}, Error, thingmodelservicev1_DeleteFeatureRequest>,
) {
  return useMutation({
    mutationFn: (data) => apiClient.featureService.Delete(data),
    ...options,
  });
}

// ==============================
// spec 校验（前端表单实时校验，不落库）
// ==============================

export function useValidateFeatureSpec(
  options?: UseMutationOptions<
    thingmodelservicev1_ValidateFeatureSpecResponse,
    Error,
    thingmodelservicev1_ValidateFeatureSpecRequest
  >,
) {
  return useMutation({
    mutationFn: (data) => apiClient.featureService.ValidateSpec(data),
    ...options,
  });
}

// ==============================
// 批量导入（保底方案：种子未初始化或需补录时用 Excel 导入）
// Batch import (fallback when seed didn't initialize or for bulk补录)
// ==============================

/**
 * 批量导入特征。按 code 幂等 upsert，失败行汇总返回。
 * rows 由前端从 Excel 解析得到（列顺序见 ImportFeaturesModal）。
 */
export function useImportFeatures(
  options?: UseMutationOptions<
    thingmodelservicev1_ImportFeaturesResponse,
    Error,
    thingmodelservicev1_ImportFeatureRow[]
  >,
) {
  return useMutation({
    // 默认 skipInvalid=true：单行失败不阻断整批，适合大批量导入
    mutationFn: (rows) =>
      apiClient.featureService.ImportFeatures({ rows, skipInvalid: true }),
    ...options,
  });
}
