export type ApiRequest<T = Record<string, any>> = (
  params?: Record<string, any>
) => Promise<Record<string, any>>;

export function useProTable<T>(requestFn: ApiRequest<T>) {
  const data = ref<T[]>([]);
  const loading = ref(false);
  const pagination = reactive({ page: 1, size: 20, total: 0 });

  async function fetch(params = {}) {
    loading.value = true;
    try {
      const res = await requestFn({ ...pagination, ...params });
      data.value = res.list;
      pagination.total = res.total;
    } finally {
      loading.value = false;
    }
  }

  return { data, loading, pagination, fetch };
}
