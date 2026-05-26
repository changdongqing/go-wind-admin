import { ref, reactive } from "vue";
import { ProModalConfig } from "../ProModal/types";

export function useModalState<T>(config: ProModalConfig<T>) {
  const visible = ref(false);
  const mode = ref<"add" | "edit" | "view">("add");
  const formData = reactive<Record<string, any>>({});
  const pk = config.pk ?? "id";

  function open(m: typeof mode.value, row?: T) {
    mode.value = m;
    Object.keys(formData).forEach((k) => delete formData[k]);
    if (row) {
      Object.keys(row).forEach((k) => {
        formData[k] = (row as any)[k];
      });
    }
    visible.value = true;
  }

  async function submit() {
    if (config.submitAction && mode.value !== "view") {
      await config.submitAction(formData as T);
    }
    visible.value = false;
  }

  return { visible, mode, formData, open, submit, pk };
}
