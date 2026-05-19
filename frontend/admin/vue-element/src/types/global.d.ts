/**
 * 全局类型声明
 *
 * @deprecated 请使用 @/types 下的具名导出
 */
declare global {
  type TagView = import("@/types/ui").TagView;

  type ClassType = Array<object | string> | object | string;

  interface BasicOption {
    label: string;
    value: string;
  }

  type SelectOption = BasicOption;

  type TabOption = BasicOption;
}

export {};
