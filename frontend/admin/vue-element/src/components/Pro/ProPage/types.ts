import type { DialogProps, DrawerProps } from "element-plus";
import type { ProFormField } from "../ProForm/types";
import type { ProTableColumn, TableEngine } from "../ProTable/types";

export type ToolbarLeft = "add" | "delete" | "import" | "export";
export type ToolbarRight = "refresh" | "filter" | "search" | "zoom";

export interface ToolsButton {
  name: string;
  label?: string;
  auth?: string | string[];
  attrs?: Record<string, any>;
  visible?: (...args: any[]) => boolean;
}

export type ListAction<TItem = any, TQuery = any> = (
  queryParams: TQuery & { page?: number; limit?: number }
) => Promise<{ items: TItem[]; total: number } | TItem[]>;

// 全局 ProPage 配置
export interface ProPageGlobalConfig {
  engine?: TableEngine;
  rowKey?: string;
  pagination?: {
    enabled?: boolean;
    pageSizes?: number[];
    pageSize?: number;
  };
  toolbar?: {
    defaultLeft?: Array<ToolbarLeft | ToolsButton>;
    defaultRight?: Array<ToolbarRight | ToolsButton>;
  };
  message?: {
    fetchSuccess?: boolean;
    fetchError?: boolean;
    deleteSuccess?: string;
    deleteConfirm?: string;
    submitSuccess?: string;
  };
}

// 页面配置
export interface ProPageConfig<T = any, Q = any> {
  exportFilename?: string;
  engine?: TableEngine;
  rowKey?: string;
  tableId?: string;
  pageKey?: string;

  search?: {
    fields?: ProFormField[];
    isExpandable?: boolean;
    showNumber?: number;
    colon?: boolean;
    grid?: boolean | "left" | "right";
  };

  table: {
    columns: ProTableColumn<T>[];
    tableAttrs?: Record<string, any>;
    pagination?: boolean;
    toolbar?: Array<ToolbarLeft | ToolsButton>;
    toolbarRight?: Array<ToolbarLeft | ToolsButton>;
    defaultToolbar?: Array<ToolbarRight | ToolsButton>;
    listAction: ListAction<T, Q>;
    request?: { pageName: string; limitName: string };
    modifyAction?: (data: { [key: string]: any; field: string; value: any }) => Promise<any>;
    deleteAction?: (ids: string) => Promise<any>;
    exportsAction?: (queryParams: Q) => Promise<any[]>;
    importsAction?: (data: any[]) => Promise<any>;
    importAction?: (file: File) => Promise<any>;
    importTemplate?: string | (() => Promise<any>);
  };

  modal?: {
    component?: "dialog" | "drawer";
    dialog?: Partial<Omit<DialogProps, "modelValue">>;
    drawer?: Partial<Omit<DrawerProps, "modelValue">>;
    form?: Record<string, any>;
    colon?: boolean;
    fields: ProFormField<T>[];
    beforeSubmit?: (data: T) => Promise<T> | T;
    submitAction?: (data: T) => Promise<any>;
    afterSubmit?: () => void;
  };
}
