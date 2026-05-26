import { PaginationProps, TableProps } from "element-plus";
import { ApiRequest, RowAction } from "../types";
import { DefaultRow } from "element-plus/es/components/table/src/table/defaults";

export type TableValueType =
  | "text"
  | "image"
  | "tag"
  | "switch"
  | "date"
  | "link"
  | "tool"
  | "custom";

export interface ProTableColumn<T = any> {
  prop: keyof T & string;
  label: string;
  type?: "selection" | "index" | "expand";
  width?: string | number;
  minWidth?: string | number;
  fixed?: "left" | "right" | boolean;
  sortable?: boolean | "custom";
  valueType?: TableValueType;
  show?: boolean;
  treeNode?: boolean;
  action?: { name: string; text: string; perm?: string; attrs?: Partial<ButtonProps> }[];
  attrs?: Record<string, any>; // 透传 el-table-column 原生属性
  slotName?: string;
}

export interface ProTableConfig<T extends DefaultRow = any, Q = any> {
  permPrefix?: string;
  table?: Partial<TableProps<T>> & { treeConfig?: any };
  pagination?: boolean | Partial<PaginationProps>;
  request: ApiRequest<T, Q>;
  requestParams?: { pageName: string; limitName: string };
  modifyAction?: RowAction<T>;
  deleteAction?: (ids: string) => Promise<void>;
  exportAction?: (params: Q) => Promise<Blob>;
  importAction?: (file: File) => Promise<void>;
  pk?: string;
  toolbar?: string[];
  columns: ProTableColumn<T>[];
}
