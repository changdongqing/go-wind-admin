<template>
  <component
    :is="props.component === 'drawer' ? 'el-drawer' : 'el-dialog'"
    v-model="visible"
    v-bind="containerProps"
    @close="emit('update:visible', false)"
  >
    <ProForm
      ref="formRef"
      v-model="formData"
      :fields="fields"
      :disabled="mode === 'view'"
      v-bind="formProps"
    >
      <template v-for="(_, name) in $slots" #[name]="slotProps">
        <slot :name="name" v-bind="slotProps" />
      </template>
    </ProForm>

    <template #footer>
      <slot name="footer">
        <el-button @click="visible = false">{{ t("common.cancel") }}</el-button>
        <el-button
          v-if="mode !== 'view'"
          type="primary"
          @click="handleSubmit"
          :loading="submitting"
        >
          {{ t("common.confirm") }}
        </el-button>
      </slot>
    </template>
  </component>
</template>

<script setup lang="ts" generic="T extends Record<string, any>">
import { computed, ref } from "vue";
import { ElDialog, ElDrawer } from "element-plus";
import ProForm from "./ProForm.vue";
import type { ProModalConfig } from "./types";

defineOptions({ inheritAttrs: false });
const props = defineProps<{
  visible: boolean;
  mode: "add" | "edit" | "view";
  config: ProModalConfig<T>;
  formData: T;
}>();
const emit = defineEmits<{ "update:visible": [boolean]; submit: [] }>();

const visible = computed({ get: () => props.visible, set: (v) => emit("update:visible", v) });
const containerProps = computed(() =>
  props.component === "drawer" ? props.config.drawer : props.config.dialog
);
const formProps = computed(() => props.config.form ?? {});
const fields = computed(() => props.config.fields);

const formRef = ref<InstanceType<typeof ProForm>>();
const submitting = ref(false);

async function handleSubmit() {
  if (props.mode === "view") return;
  try {
    await formRef.value?.validate();
    submitting.value = true;
    await props.config.submitAction?.(props.formData);
    emit("submit");
    visible.value = false;
  } finally {
    submitting.value = false;
  }
}

defineExpose({ formRef });
</script>
