/**
 * 单位选择器 / Unit selector.
 *
 * - 受控组件：`value` 绑定单位编码 `unitCode`，`onChange(code)` 回写表单。
 * - 选中后通过 `onUnitChange(unit)` 通知宿主回填 `unitSymbol`（保持现有数据契约）。
 * - 列表数据来自 `useListUnits`，仅展示 `isEnabled !== false` 的单位（兼容旧数据未显式禁用的情况）。
 * - 搜索支持按 name / nameEn / code / symbol 四个维度模糊匹配。
 * - 如果传入 value 不在选项里（例如历史导入的脏数据），仍按原始 code 显示，避免清空。
 *
 * Mirrors Antd Select API; designed to drop into existing `<Form.Item>` cells.
 */
import { useMemo } from 'react';
import { Select } from 'antd';
import type { SelectProps } from 'antd';
import { useTranslation } from 'react-i18next';

import { useListUnits } from '@/api/hooks/unit';
import { PaginationQuery } from '@/core/transport/rest';
import type { thingmodelservicev1_Unit } from '@/api/generated/admin/service/v1';

export interface UnitSelectProps
  extends Omit<SelectProps<string>, 'options' | 'onChange' | 'value'> {
  /** 当前选中的单位编码 / Current unit code */
  value?: string;
  /** 仅传 code（用于 Form.Item 受控） / Receives the next unit code */
  onChange?: (code?: string) => void;
  /** 选中单位对象的副回调，供宿主回填 unitSymbol/unitId 等 */
  onUnitChange?: (unit?: thingmodelservicev1_Unit) => void;
}

// 一次性拉取常驻缓存：前端按 code 全量过滤，避免对每行单元格重复请求
const ALL_UNITS_QUERY = new PaginationQuery({
  paging: { page: 1, pageSize: 1000 },
  orderBy: ['sort_order', 'id'],
});

const UnitSelect: React.FC<UnitSelectProps> = ({
  value,
  onChange,
  onUnitChange,
  placeholder,
  ...rest
}) => {
  const { t } = useTranslation('feature');
  const { data, isLoading } = useListUnits(ALL_UNITS_QUERY, {
    // 5 分钟内复用同一份 list，避免每行 Select 单独触发请求
    staleTime: 5 * 60 * 1000,
  });

  const units = useMemo<thingmodelservicev1_Unit[]>(() => {
    const items = data?.items ?? [];
    return items.filter((u) => u.isEnabled !== false);
  }, [data]);

  const options = useMemo(() => {
    const opts = units.map((u) => ({
      label: `${u.name ?? u.code}${u.symbol ? ` (${u.symbol})` : ''}`,
      value: u.code ?? '',
      // 自定义查询关键字
      keywords: [u.name, u.nameEn, u.code, u.symbol].filter(Boolean).join(' ').toLowerCase(),
    }));
    // 兼容历史脏数据：如果 value 在选项中不存在，追加一个只读式选项保留原值
    if (value && !opts.some((o) => o.value === value)) {
      opts.unshift({ label: value, value, keywords: value.toLowerCase() });
    }
    return opts;
  }, [units, value]);

  const codeToUnit = useMemo(() => {
    const map = new Map<string, thingmodelservicev1_Unit>();
    for (const u of units) {
      if (u.code) map.set(u.code, u);
    }
    return map;
  }, [units]);

  return (
    <Select
      showSearch
      allowClear
      loading={isLoading}
      placeholder={placeholder ?? t('selectUnit')}
      value={value || undefined}
      options={options}
      optionFilterProp="keywords"
      filterOption={(input, opt) =>
        ((opt?.keywords as string | undefined) ?? '').includes(input.toLowerCase())
      }
      onChange={(next) => {
        onChange?.(next);
        onUnitChange?.(next ? codeToUnit.get(next) : undefined);
      }}
      {...rest}
    />
  );
};

export default UnitSelect;
