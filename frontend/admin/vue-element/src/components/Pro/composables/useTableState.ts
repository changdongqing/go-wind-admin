import { ref, reactive, computed } from "vue";
import { ProTableConfig } from "../ProTable/types";

export function useTableState<T, Q>(config: ProTableConfig<T, Q>) {
  const data = ref<T[]>([]);
  const loading = ref(false);
  const selection = ref<T[]>([]);
  const pagination = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
    ...((typeof config.pagination === "object" ? config.pagination : {}) as any),
  });

  const pk = config.pk ?? "id";
  const reqParams = config.requestParams ?? { pageName: "page", limitName: "pageSize" };

  async function fetch(queryParams: Q = {}, resetPage = false) {
    loading.value = true;
    if (resetPage) pagination.currentPage = 1;

    const params = {
      [reqParams.pageName]: pagination.currentPage,
      [reqParams.limitName]: pagination.pageSize,
      ...queryParams,
    } as any;

    try {
      const res = await config.request(params);
      data.value = res.list;
      pagination.total = res.total;
    } finally {
      loading.value = false;
    }
  }

  function handleSelectionChange(rows: T[]) {
    selection.value = rows;
  }
  function getSelectionIds() {
    return selection.value.map((r) => (r as any)[pk]);
  }

  return {
    data,
    loading,
    pagination,
    selection,
    fetch,
    handleSelectionChange,
    getSelectionIds,
  };
}
