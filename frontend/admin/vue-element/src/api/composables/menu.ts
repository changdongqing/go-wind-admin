import { computed } from "vue";
import type {
  permissionservicev1_DeleteMenuRequest,
  permissionservicev1_GetMenuRequest,
  permissionservicev1_ListMenuResponse,
  permissionservicev1_Menu,
  permissionservicev1_Menu as Menu,
  permissionservicev1_Menu_Type as Menu_Type,
} from "@/api/generated/admin/service/v1";
import {
  useMutation,
  type UseMutationOptions,
  useQuery,
  type UseQueryOptions,
} from "@tanstack/vue-query";
import { makeUpdateMask, type PaginationQuery } from "@/core/transport/rest";
import { listMenus, getMenu, createMenu, updateMenu, deleteMenu } from "@/api/service/menu";
import { queryClient } from "@/plugins/vue-query";

import { i18n } from '@/core/i18n';

const t = i18n.global.t;

// ==============================
// 菜单管理
// ==============================

export function useListMenus(
  query: PaginationQuery,
  options?: UseQueryOptions<permissionservicev1_ListMenuResponse, Error>
) {
  return useQuery({
    queryKey: ["listMenus", query],
    queryFn: () => listMenus(query),
    ...options,
  });
}

export async function fetchListMenus(params: PaginationQuery) {
  return queryClient.fetchQuery({
    queryKey: ["listMenus", params],
    queryFn: () => listMenus(params),
    retry: 0,
  });
}

export function useGetMenu(
  req: permissionservicev1_GetMenuRequest,
  options?: UseQueryOptions<permissionservicev1_Menu, Error>
) {
  return useQuery({
    queryKey: ["getMenu", req],
    queryFn: () => getMenu(req),
    ...options,
  });
}

export function useCreateMenu(options?: UseMutationOptions<{}, Error, Record<string, any>>) {
  return useMutation({
    mutationFn: (values) => createMenu({ data: { ...values } as Menu }),
    ...options,
  });
}

export function useUpdateMenu(
  options?: UseMutationOptions<{}, Error, { id: number; values: Record<string, any> }>
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
  options?: UseMutationOptions<{}, Error, permissionservicev1_DeleteMenuRequest>
) {
  return useMutation({
    mutationFn: (data) => deleteMenu(data),
    ...options,
  });
}

// ==============================
// 菜单枚举与工具函数
// ==============================

export const menuTypeList = computed(() => [
  { value: "CATALOG", label: t("enum.menu.type.CATALOG") },
  { value: "MENU", label: t("enum.menu.type.MENU") },
  { value: "BUTTON", label: t("enum.menu.type.BUTTON") },
  { value: "EMBEDDED", label: t("enum.menu.type.EMBEDDED") },
  { value: "LINK", label: t("enum.menu.type.LINK") },
]);

export function menuTypeToName(menuType: any): string {
  const values = menuTypeList.value;
  const matchedItem = values.find((item) => item.value === menuType);
  return matchedItem ? matchedItem.label : "";
}

export function menuTypeToColor(menuType: Menu_Type) {
  switch (menuType) {
    case "BUTTON":
      return "#F56C6C";
    case "CATALOG":
      return "#27AE60";
    case "EMBEDDED":
      return "#4096FF";
    case "LINK":
      return "#9B59B6";
    case "MENU":
      return "#165DFF";
    default:
      return "#86909C";
  }
}

export const isCatalog = (type: string) => type === "CATALOG";
export const isMenu = (type: string) => type === "MENU";
export const isButton = (type: string) => type === "BUTTON";
export const isEmbedded = (type: string) => type === "EMBEDDED";
export const isLink = (type: string) => type === "LINK";

export function travelMenuChild(nodes: Menu[] | undefined, parent: Menu): boolean {
  if (nodes === undefined) return false;
  if (parent.parentId === 0 || parent.parentId === undefined) {
    if (parent?.meta?.title) parent.meta.title = t(parent?.meta?.title ?? "");
    nodes.push(parent);
    return true;
  }
  for (const node of nodes) {
    if (node === undefined) continue;
    if (node.id === parent.parentId) {
      if (parent?.meta?.title) parent.meta.title = t(parent?.meta?.title ?? "");
      if (node.children !== undefined) node.children.push(parent);
      return true;
    }
    if (travelMenuChild(node.children, parent)) return true;
  }
  return false;
}

export function buildMenuTree(menus: Menu[]): Menu[] {
  const tree: Menu[] = [];
  for (const menu of menus) {
    if (!menu) continue;
    if (menu.parentId !== 0 && menu.parentId !== undefined) continue;
    if (menu?.meta?.title) menu.meta.title = t(menu?.meta?.title ?? "");
    tree.push(menu);
  }
  for (const menu of menus) {
    if (!menu) continue;
    if (menu.parentId === 0 || menu.parentId === undefined) continue;
    if (travelMenuChild(tree, menu)) continue;
    if (menu?.meta?.title) menu.meta.title = t(menu?.meta?.title ?? "");
    tree.push(menu);
  }
  return tree;
}
