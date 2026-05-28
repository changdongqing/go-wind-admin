<template>
  <component
    :is="props.config.component === 'drawer' ? 'ElDrawer' : 'ElDialog'"
    v-model="visible"
    v-bind="containerProps"
    @close="handleClose"
  >
    <ProForm
      ref="formRef"
      :model-value="formData"
      :fields="fields"
      :disabled="mode === 'view'"
      :colon="config.colon"
      v-bind="formProps"
    >
      <template v-for="(_, name) in $slots" #[name]="slotProps">
        <slot :name="name" v-bind="slotProps" />
      </template>
    </ProForm>

    <template #footer>
      <slot name="footer">
        <ElButton @click="handleClose">{{ t("common.button.cancel") }}</ElButton>
        <ElButton v-if="mode !== 'view'" type="primary" :loading="submitting" @click="handleSubmit">
          {{ t("common.button.confirm") }}
        </ElButton>
      </slot>
    </template>
  </component>
</template>

<script setup lang="ts" generic="T extends Record<string, any>">
import { computed, ref } from "vue";
import { ElButton } from "element-plus";
import { useI18n } from '@/core/i18n';
import ProForm from "../ProForm/index.vue";
import type { ProModalConfig, ModalMode } from "./types";

defineOptions({ inheritAttrs: false });

const props = withDefaults(
  defineProps<{
    visible: boolean;
    mode?: ModalMode;
    config: ProModalConfig<T>;
    formData: T;
  }>(),
  { mode: "add" }
);
const emit = defineEmits<{
  "update:visible": [boolean];
  submit: [];
}>();

const { t } = useI18n();

const visible = computed({
  get: () => props.visible,
  set: (v) => emit("update:visible", v),
});
const containerProps = computed(() =>
  props.config.component === "drawer"
    ? { destroyOnClose: true, appendToBody: true, lockScroll: false, ...props.config.drawer }
    : { destroyOnClose: true, alignCenter: true, appendToBody: true, lockScroll: false, ...props.config.dialog }
);
const formProps = computed(() => ({
  labelWidth: "auto",
  ...props.config.form,
}));
const fields = computed(() => props.config.fields);

const formRef = ref<any>(null);
const submitting = ref(false);

function handleClose() {
  visible.value = false;
  formRef.value?.resetFields();
}

async function handleSubmit() {
  if (props.mode === "view") return;
  try {
    await formRef.value?.validate();
    submitting.value = true;
    if (typeof props.config.beforeSubmit === "function") {
      props.config.beforeSubmit(props.formData);
    }
    await props.config.submitAction?.(props.formData);
    emit("submit");
    visible.value = false;
  } finally {
    submitting.value = false;
  }
}

defineExpose({ formRef });
</script>
