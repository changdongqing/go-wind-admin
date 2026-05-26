import type { PaginationProps } from "element-plus";

export interface ProPaginationProps extends Partial<PaginationProps> {
  modelValue?: {
    currentPage?: number;
    pageSize?: number;
  };
  total?: number;
  showTotal?: boolean;
  showSizes?: boolean;
  showJump?: boolean;
  layout?: string;
  pageSizes?: number[];
  disabled?: boolean;
  background?: boolean;
  hideOnSinglePage?: boolean;
}

export interface PaginationEmits {
  "update:modelValue": [value: { currentPage: number; pageSize: number }];
  "current-change": [currentPage: number];
  "size-change": [pageSize: number];
  "prev-click": [currentPage: number];
  "next-click": [currentPage: number];
}
