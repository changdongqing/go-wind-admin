import { DialogProps, DrawerProps } from "element-plus";
import { ProFormField } from "../ProForm/types";

export type ModalType = "dialog" | "drawer";

export interface ProModalConfig<T = any> {
  component?: ModalType;
  dialog?: Partial<Omit<DialogProps, "modelValue">>;
  drawer?: Partial<Omit<DrawerProps, "modelValue">>;
  form?: Partial<FormProps>;
  fields: ProFormField<T>[];
  pk?: string;
  submitAction?: (data: T) => Promise<void>;
}
