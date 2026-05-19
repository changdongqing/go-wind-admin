import type { FieldOptions, FormContext, GenericObject } from "vee-validate";
import type { ZodTypeAny } from "zod";

import type { Component, HtmlHTMLAttributes, Ref } from "vue";

declare global {
  type FormLayout = "horizontal" | "vertical";

  type BaseFormComponentType =
    | "DefaultButton"
    | "PrimaryButton"
    | "Checkbox"
    | "Input"
    | "InputPassword"
    | "PinInput"
    | "Select"
    | (Record<never, never> & string);

  type Breakpoints = "" | "2xl:" | "3xl:" | "lg:" | "md:" | "sm:" | "xl:";

  type GridCols = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 | 13;

  type WrapperClassType = `${Breakpoints}grid-cols-${GridCols}` | (Record<never, never> & string);

  type FormFieldOptions = Partial<
    {
      validateOnBlur?: boolean;
      validateOnChange?: boolean;
      validateOnInput?: boolean;
      validateOnModelUpdate?: boolean;
    } & FieldOptions
  >;

  type RenderComponentContentType = (
    value: Partial<Record<string, any>>,
    api: FormActions
  ) => Record<string, any>;

  type HandleSubmitFn = (values: Record<string, any>) => Promise<void> | void;

  type HandleResetFn = (values: Record<string, any>) => Promise<void> | void;

  type FieldMappingTime = [string, [string, string], ([string, string] | string)?][];

  export interface ActionButtonOptions extends ButtonProps {
    [key: string]: any;
    content?: string;
    show?: boolean;
  }

  interface FormShape {
    /** 默认值 */
    default?: any;
    /** 字段名 */
    fieldName: string;
    /** 是否必填 */
    required?: boolean;
    rules?: ZodTypeAny;
  }

  type MaybeComponentPropKey =
    | "options"
    | "placeholder"
    | "title"
    | keyof HtmlHTMLAttributes
    | (Record<never, never> & string);

  type ExtendedFormApi = {
    useStore: <T = NoInfer<FormProps>>(
      selector?: (state: NoInfer<FormProps>) => T
    ) => Readonly<Ref<T>>;
  } & FormApi;

  export interface FormProps<T extends BaseFormComponentType = BaseFormComponentType> extends Omit<
    FormRenderProps<T>,
    "componentBindEventMap" | "componentMap" | "form"
  > {
    /**
     * 操作按钮是否反转（提交按钮前置）
     */
    actionButtonsReverse?: boolean;
    /**
     * 表单操作区域class
     */
    actionWrapperClass?: ClassType;
    /**
     * 表单字段映射成时间格式
     */
    fieldMappingTime?: FieldMappingTime;
    /**
     * 表单重置回调
     */
    handleReset?: HandleResetFn;
    /**
     * 表单提交回调
     */
    handleSubmit?: HandleSubmitFn;
    /**
     * 表单值变化回调
     */
    handleValuesChange?: (values: Record<string, any>) => void;
    /**
     * 重置按钮参数
     */
    resetButtonOptions?: ActionButtonOptions;
    /**
     * 是否显示默认操作按钮
     * @default true
     */
    showDefaultActions?: boolean;

    /**
     * 提交按钮参数
     */
    submitButtonOptions?: ActionButtonOptions;

    /**
     * 是否在字段值改变时提交表单
     * @default false
     */
    submitOnChange?: boolean;

    /**
     * 是否在回车时提交表单
     * @default false
     */
    submitOnEnter?: boolean;
  }
}

export {};
