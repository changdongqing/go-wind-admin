// ================= 基础泛型约束 =================
export type ApiRequest<T = any, Q = any> = (params: Q) => Promise<{ list: T[]; total: number }>;
export type RowAction<T = any> = (data: { row: T; field: string; value: any }) => Promise<void>;
