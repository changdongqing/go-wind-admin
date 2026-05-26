import type { FormProps, CardProps } from "element-plus";
import type { ProFormField } from "../ProForm/types";

export interface ProSearchConfig<T = any> {
  // 基础配置
  permPrefix?: string;
  colon?: boolean;

  // 布局配置
  grid?: boolean | "left" | "right";
  inline?: boolean;

  // 展开/收起
  isExpandable?: boolean;
  showNumber?: number;

  // 卡片属性
  cardAttrs?: Partial<CardProps>;

  // 表单属性
  form?: Partial<FormProps>;

  // 表单项
  fields: ProFormField<T>[];

  // 按钮配置
  showSearchButton?: boolean;
  showResetButton?: boolean;
  searchButtonText?: string;
  resetButtonText?: string;

  // 权限控制
  searchPerm?: string;
  resetPerm?: string;
}

export interface ProSearchEmits<T = any> {
  search: [queryParams: T];
  reset: [queryParams: T];
  expand: [expanded: boolean];
}
