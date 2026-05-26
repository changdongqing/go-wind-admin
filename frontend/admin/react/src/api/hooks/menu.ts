import {
  type permissionservicev1_CreateMenuRequest,
  type permissionservicev1_DeleteMenuRequest,
  type permissionservicev1_GetMenuRequest,
  type permissionservicev1_ListMenuResponse,
  type permissionservicev1_Menu,
} from '@/api/generated/admin/service/v1';
import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from '@tanstack/react-query';
import { makeUpdateMask, type PaginationQuery, queryClient } from '@/core';
import { listMenus, getMenu, createMenu, updateMenu, deleteMenu } from '@/api/service/menu';

// ==============================
// 菜单管理
// ==============================

export function useListMenus(
  query: PaginationQuery,
  options?: UseQueryOptions<permissionservicev1_ListMenuResponse, Error>,
) {
  return useQuery({
    queryKey: ['listMenus', query],
    queryFn: () => listMenus(query),
    ...options,
  });
}

export async function fetchListMenus(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ['listMenus', params],
    queryFn: () => listMenus(params),
    retry: 0,
  });
}

export function useGetMenu(
  req: permissionservicev1_GetMenuRequest,
  options?: UseQueryOptions<permissionservicev1_Menu, Error>,
) {
  return useQuery({
    queryKey: ['getMenu', req],
    queryFn: () => getMenu(req),
    ...options,
  });
}

export function useCreateMenu(
  options?: UseMutationOptions<{}, Error, permissionservicev1_CreateMenuRequest>,
) {
  return useMutation({
    mutationFn: (data) => createMenu(data),
    ...options,
  });
}

export function useUpdateMenu(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>,
) {
  return useMutation({
    mutationFn: ({ id, values }: { id: number; values: Record<string, any> }) =>
      updateMenu({
        id,
        data: { ...values } as any,
        updateMask: makeUpdateMask(Object.keys(values ?? {})),
      }),
    ...options,
  });
}

export function useDeleteMenu(
  options?: UseMutationOptions<{}, Error, permissionservicev1_DeleteMenuRequest>,
) {
  return useMutation({
    mutationFn: (data) => deleteMenu(data),
    ...options,
  });
}
