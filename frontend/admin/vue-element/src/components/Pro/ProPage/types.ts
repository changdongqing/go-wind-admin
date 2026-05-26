import type { ProFormField } from "../ProForm/types";
import type { ProTableColumn } from "../ProTable/types";

// 请求列表 API 类型
export type RequestApi<T = any> = (params: any) => Promise<{
  records: T[];
  total: number;
}>;

export interface ProPageProps<T = any> {
  // 查询表单
  queryFields?: ProFormField[];
  // 表格列
  tableColumns: ProTableColumn<T>[];
  // 新增/编辑表单字段
  formFields?: ProFormField[];
  // 请求列表接口
  requestApi: RequestApi<T>;
  // 新增接口
  addApi?: (data: T) => Promise<any>;
  // 编辑接口
  editApi?: (data: T) => Promise<any>;
  // 删除接口
  deleteApi?: (ids: string | number) => Promise<any>;
  // 批量删除接口
  batchDeleteApi?: (ids: string[]) => Promise<any>;
  // 主键名
  rowKey?: string;
  // 标题
  title?: string;
}
