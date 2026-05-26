<template>
  <el-form v-bind="$attrs" ref="formRef" :model="modelValue">
    <el-row :gutter="20">
      <el-col
        v-for="field in fields"
        :key="String(field.field)"
        :span="field.span ?? 24"
        v-show="!field.hidden"
        v-bind="field.col"
      >
        <el-form-item :label="field.label" :prop="String(field.field)" :rules="field.rules">
          <template #label>
            <span>{{ field.label }}</span>
            <slot :name="`label-${String(field.field)}`" />
          </template>

          <slot v-if="field.slotName" :name="field.slotName" :model="modelValue" />
          <component
            v-else
            :is="resolveComponent(field.type)"
            v-model="modelValue[field.field]"
            v-bind="field.attrs"
          >
            <template v-if="['select', 'radio', 'checkbox'].includes(field.type ?? '')">
              <component
                :is="
                  field.type === 'select'
                    ? 'el-option'
                    : field.type === 'radio'
                      ? 'el-radio'
                      : 'el-checkbox'
                "
                v-for="opt in field.options"
                :key="opt.value"
                :label="opt.label"
                :value="opt.value"
              />
            </template>
          </component>
        </el-form-item>
      </el-col>
    </el-row>
  </el-form>
</template>

<script setup lang="ts" generic="T extends Record<string, any>">
import { ref } from "vue";
import {
  ElForm,
  ElFormItem,
  ElInput,
  ElSelect,
  ElOption,
  ElSwitch,
  ElDatePicker,
} from "element-plus";
import type { ProFormField } from "./types";

defineOptions({ inheritAttrs: false });
const props = defineProps<{ modelValue: T; fields: ProFormField<T>[] }>();
const emit = defineEmits<{ "update:modelValue": [T] }>();
const formRef = ref<InstanceType<typeof ElForm>>();

defineExpose({
  formRef,
  validate: () => formRef.value?.validate(),
  resetFields: () => formRef.value?.resetFields(),
});

const resolveComponent = (type?: string) => {
  const map: Record<string, any> = {
    input: ElInput,
    select: ElSelect,
    switch: ElSwitch,
    date: ElDatePicker,
    number: () => h(ElInput, { type: "number" }),
  };
  return map[type ?? "input"] || ElInput;
};
</script>
