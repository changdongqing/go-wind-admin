<template>
  <ElForm v-bind="$attrs" ref="formRef" :model="modelValue" :disabled="disabled">
    <ElRow :gutter="20">
      <ElCol
        v-for="field in resolvedFields"
        v-show="!field.hidden"
        :key="String(field.field)"
        :span="field.span ?? 24"
        v-bind="field.col"
      >
        <ElFormItem :label="field.label" :prop="String(field.field)" :rules="field.rules">
          <template #label>
            <span class="flex items-center gap-1">
              {{ field.label }}
              <ElTooltip v-if="field.tips" v-bind="getTooltipProps(field.tips)">
                <ElIcon class="text-gray-400"><QuestionFilled /></ElIcon>
              </ElTooltip>
              <span v-if="colon" class="ml-0.5">:</span>
            </span>
          </template>

          <!-- 自定义插槽 -->
          <slot
            v-if="field.slotName || field.type === 'custom'"
            :name="field.slotName ?? String(field.field)"
            :model="modelValue"
            :field="field.field"
            :attrs="{ style: { width: '100%' }, ...field.attrs }"
          />

          <!-- api-tree-select -->
          <ElTreeSelect
            v-else-if="field.type === 'api-tree-select'"
            v-model="modelValue[field.field]"
            v-bind="{ style: { width: '100%' }, ...field.attrs }"
            v-on="field.events || {}"
          />

          <!-- 动态组件 -->
          <component
            :is="resolveComponent(field.type)"
            v-else
            v-model="modelValue[field.field]"
            v-bind="{ style: { width: '100%' }, ...field.attrs }"
            v-on="field.events || {}"
          >
            <template v-if="['select', 'radio', 'checkbox'].includes(field.type ?? '')">
              <component
                :is="getChildrenComponent(field.type!)"
                v-for="opt in field.options"
                :key="opt.value"
                :label="opt.label"
                :value="opt.value"
                :disabled="opt.disabled"
              />
            </template>
          </component>
        </ElFormItem>
      </ElCol>
    </ElRow>
  </ElForm>
</template>

<script setup lang="ts" generic="T extends Record<string, any>">
import { ref, markRaw, onMounted, reactive } from "vue";
import {
  ElForm,
  ElFormItem,
  ElInput,
  ElSelect,
  ElOption,
  ElSwitch,
  ElDatePicker,
  ElTimePicker,
  ElTimeSelect,
  ElInputNumber,
  ElCascader,
  ElTreeSelect,
  ElRadioGroup,
  ElRadio,
  ElCheckboxGroup,
  ElCheckbox,
  ElRow,
  ElCol,
  ElTooltip,
  ElIcon,
} from "element-plus";
import { QuestionFilled } from "@element-plus/icons-vue";
import InputTag from "@/components/InputTag/index.vue";
import IconSelect from "@/components/IconSelect/index.vue";
import type { ProFormField } from "./types";

defineOptions({ inheritAttrs: false });

const props = withDefaults(
  defineProps<{
    modelValue: T;
    fields: ProFormField<T>[];
    colon?: boolean;
    disabled?: boolean;
  }>(),
  { colon: false, disabled: false }
);

defineEmits<{ "update:modelValue": [T] }>();
const formRef = ref<InstanceType<typeof ElForm>>();

// 解析字段（支持 initFn）
const resolvedFields = reactive([...props.fields]) as ProFormField<T>[];

onMounted(() => {
  resolvedFields.forEach((field) => {
    if (field.initFn) {
      field.initFn(field as any as Record<string, any>);
    }
    // api-tree-select: 异步加载数据
    if (field.type === "api-tree-select" && typeof field.api === "function") {
      field.api().then((data) => {
        if (!field.attrs) field.attrs = {};
        field.attrs.data = data;
      });
    }
  });
});

// 组件映射表
const componentMap: Record<string, any> = {
  input: markRaw(ElInput),
  textarea: markRaw(ElInput),
  select: markRaw(ElSelect),
  switch: markRaw(ElSwitch),
  "date-picker": markRaw(ElDatePicker),
  date: markRaw(ElDatePicker),
  "time-picker": markRaw(ElTimePicker),
  "time-select": markRaw(ElTimeSelect),
  "input-number": markRaw(ElInputNumber),
  number: markRaw(ElInputNumber),
  cascader: markRaw(ElCascader),
  "tree-select": markRaw(ElTreeSelect),
  "input-tag": markRaw(InputTag),
  "custom-tag": markRaw(InputTag),
  "icon-select": markRaw(IconSelect),
  radio: markRaw(ElRadioGroup),
  checkbox: markRaw(ElCheckboxGroup),
  custom: "",
};

// 子组件映射
const childrenMap: Record<string, any> = {
  select: markRaw(ElOption),
  radio: markRaw(ElRadio),
  checkbox: markRaw(ElCheckbox),
};

const resolveComponent = (type?: string) => {
  if (!type) return ElInput;
  if (type === "textarea") return markRaw(ElInput);
  return componentMap[type] || ElInput;
};

const getChildrenComponent = (type: string) => {
  return childrenMap[type] || ElOption;
};

const getTooltipProps = (tips: string | Record<string, any>) => {
  return typeof tips === "string" ? { content: tips, placement: "top" } : tips;
};

defineExpose({
  formRef,
  validate: () => formRef.value?.validate(),
  resetFields: () => formRef.value?.resetFields(),
});
</script>
