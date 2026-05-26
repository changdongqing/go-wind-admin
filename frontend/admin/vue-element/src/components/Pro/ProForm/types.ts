import { ColProps, FormItemRule } from "element-plus";

export type FormValueType =
  | "input"
  | "textarea"
  | "select"
  | "radio"
  | "checkbox"
  | "switch"
  | "date"
  | "number"
  | "custom";

export interface ProFormField<T = any> {
  field: keyof T & string;
  label: string;
  type?: FormValueType;
  span?: number;
  rules?: FormItemRule[];
  initialValue?: any;
  attrs?: Record<string, any>;
  options?: { label: string; value: any }[];
  hidden?: boolean;
  col?: Partial<ColProps>;
  slotName?: string; // 自定义插槽
}
