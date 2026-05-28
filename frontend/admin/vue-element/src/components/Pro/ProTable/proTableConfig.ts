import type { ProTableConfig } from "./types";
import { DEFAULT_PAGE_SIZE, DEFAULT_PAGE_SIZES } from "@/components/Pro";

export const proTableGlobalConfig: ProTableConfig = {
  engine: "vxe",
  rowKey: "id",
  rowHeight: 42,
  headerHeight: 44,
  border: true,
  stripe: true,
  borderRadius: 6,
  fontSize: 13,
  emptyText: "暂无数据",

  pagination: {
    show: true,
    pageSizes: DEFAULT_PAGE_SIZES,
    pageSize: DEFAULT_PAGE_SIZE,
    layout: "total, sizes, prev, pager, next, jumper",
    background: true,
  },

  column: {
    align: "center",
    resizable: true,
    showOverflowTooltip: true,
    sortable: false,
  },

  toolButton: {
    size: "small",
    link: true,
    iconSize: 16,
  },
};

export function mergeTableConfig(
  global: ProTableConfig,
  component?: Partial<ProTableConfig>
): ProTableConfig {
  return {
    ...global,
    ...component,
    pagination: { ...global.pagination, ...component?.pagination },
    column: { ...global.column, ...component?.column },
    toolButton: { ...global.toolButton, ...component?.toolButton },
  };
}
