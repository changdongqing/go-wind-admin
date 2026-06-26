import { useEffect, useMemo, useState } from 'react';
import {
  Card,
  InputNumber,
  Select,
  Button,
  Space,
  Alert,
  Tag,
  Typography,
} from 'antd';
import { SwapOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

import {
  useConvertUnit,
  useListUnitsByCategory,
} from '@/api/hooks/unit';
import type { thingmodelservicev1_Unit } from '@/api/generated/admin/service/v1';
import { conversionTypeColor, getConversionTypeLabel } from './constants';

interface UnitConverterProps {
  /** 初始物理量分类 id（建议来自页面左侧选中分类）。 */
  defaultCategoryId?: number | null;
  /** 初始源单位 id（用于"从单位列表行按钮快速打开换算器"场景）。 */
  defaultSourceUnitId?: number;
}

const { Text } = Typography;

/**
 * 单位换算器 / Unit converter widget.
 *
 * 交互：选分类 → 选源/目标单位 → 输入数值 → 点击或失焦自动换算。
 * 状态非 OK 时（不同分类/不可换算/未找到）展示后端返回的 message。
 *
 * 注意：组件内 debounce 不强制（用户多以点击按钮触发；输入框防抖留待后续优化）。
 */
const UnitConverter: React.FC<UnitConverterProps> = ({
  defaultCategoryId,
  defaultSourceUnitId,
}) => {
  const { t } = useTranslation('unit');

  const [categoryId, setCategoryId] = useState<number | undefined>(
    defaultCategoryId ?? undefined,
  );
  const [sourceId, setSourceId] = useState<number | undefined>(defaultSourceUnitId);
  const [targetId, setTargetId] = useState<number | undefined>(undefined);
  const [value, setValue] = useState<number>(1);
  const [precision, setPrecision] = useState<number | undefined>(undefined);

  // 当父组件传入的分类切换时，重置选择
  useEffect(() => {
    if (defaultCategoryId !== undefined && defaultCategoryId !== null) {
      setCategoryId(defaultCategoryId);
      setSourceId(undefined);
      setTargetId(undefined);
    }
  }, [defaultCategoryId]);

  // 拉取该分类下全部单位（不分页，仅启用）
  const unitsQuery = useListUnitsByCategory(
    {
      categoryId: categoryId ?? 0,
      onlyEnabled: true,
    },
    {
      enabled: Boolean(categoryId),
    } as any,
  );

  const units: thingmodelservicev1_Unit[] = unitsQuery.data?.items || [];

  // 单位选项 / Unit options
  const unitOptions = useMemo(
    () =>
      units.map((u) => ({
        label: `${u.symbol ?? ''} - ${u.name ?? ''}${u.isBase ? ` (${t('baseBadge')})` : ''}`,
        value: u.id as number,
        ext: u,
      })),
    [units, t],
  );

  // 切换分类时自动预选：基准 → 源；任一非基准 → 目标
  useEffect(() => {
    if (!units.length) return;
    if (!sourceId) {
      const base = units.find((u) => u.isBase);
      if (base?.id) setSourceId(base.id);
    }
    if (!targetId) {
      const other = units.find((u) => !u.isBase);
      if (other?.id) setTargetId(other.id);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [units.length]);

  const convertMutation = useConvertUnit();

  const handleConvert = () => {
    if (!sourceId || !targetId) return;
    convertMutation.mutate({
      sourceUnitId: sourceId,
      targetUnitId: targetId,
      value,
      precision,
    });
  };

  const handleSwap = () => {
    setSourceId(targetId);
    setTargetId(sourceId);
  };

  const resp = convertMutation.data;
  const errorMsg = convertMutation.error?.message;

  const sourceUnit = units.find((u) => u.id === sourceId);
  const targetUnit = units.find((u) => u.id === targetId);

  return (
    <Card size="small" title={t('converter.title')}>
      <Space direction="vertical" style={{ width: '100%' }} size="middle">
        {/* 单位下拉（受父组件 categoryId 控制，因此此处只展示当前分类） */}
        {!categoryId && (
          <Alert
            type="info"
            showIcon
            message={t('converter.selectCategory')}
          />
        )}

        <Space wrap>
          <Select
            style={{ width: 220 }}
            placeholder={t('converter.selectSource')}
            value={sourceId}
            options={unitOptions}
            onChange={setSourceId}
            disabled={!categoryId}
            allowClear
            showSearch
            optionFilterProp="label"
          />
          <InputNumber
            value={value}
            onChange={(v) => setValue(typeof v === 'number' ? v : 1)}
            style={{ width: 140 }}
            placeholder={t('converter.inputValue')}
          />
          {sourceUnit?.symbol && (
            <Text type="secondary">{sourceUnit.symbol}</Text>
          )}
          <Button
            icon={<SwapOutlined />}
            onClick={handleSwap}
            disabled={!sourceId || !targetId}
          >
            {t('converter.swap')}
          </Button>
          <Select
            style={{ width: 220 }}
            placeholder={t('converter.selectTarget')}
            value={targetId}
            options={unitOptions}
            onChange={setTargetId}
            disabled={!categoryId}
            allowClear
            showSearch
            optionFilterProp="label"
          />
        </Space>

        <Space wrap>
          <InputNumber
            min={0}
            max={12}
            value={precision}
            onChange={(v) => setPrecision(typeof v === 'number' ? v : undefined)}
            placeholder={t('converter.precision')}
            style={{ width: 140 }}
          />
          <Button
            type="primary"
            icon={<ThunderboltOutlined />}
            onClick={handleConvert}
            loading={convertMutation.isPending}
            disabled={!sourceId || !targetId}
          >
            {t('converter.convert')}
          </Button>
          {sourceUnit?.conversionType && (
            <Tag color={conversionTypeColor[sourceUnit.conversionType as keyof typeof conversionTypeColor] ?? 'default'}>
              src: {getConversionTypeLabel(t, sourceUnit.conversionType)}
            </Tag>
          )}
          {targetUnit?.conversionType && (
            <Tag color={conversionTypeColor[targetUnit.conversionType as keyof typeof conversionTypeColor] ?? 'default'}>
              dst: {getConversionTypeLabel(t, targetUnit.conversionType)}
            </Tag>
          )}
        </Space>

        {/* 结果区 */}
        {errorMsg && <Alert type="error" showIcon message={errorMsg} />}
        {resp && resp.status === 'CONVERT_OK' && (
          <Alert
            type="success"
            showIcon
            message={
              <Space>
                <Text strong>{t('converter.result')}：</Text>
                <Text>{resp.result}</Text>
                {targetUnit?.symbol && <Text>{targetUnit.symbol}</Text>}
              </Space>
            }
            description={
              <Space direction="vertical" size={2}>
                <Text type="secondary">
                  {t('converter.chain')}：{resp.formula}
                </Text>
                {resp.baseValue !== undefined && (
                  <Text type="secondary">base = {resp.baseValue}</Text>
                )}
              </Space>
            }
          />
        )}
        {resp && resp.status === 'CONVERT_DIFFERENT_CATEGORY' && (
          <Alert
            type="warning"
            showIcon
            message={t('converter.messageDifferentCategory')}
            description={resp.message}
          />
        )}
        {resp && resp.status === 'CONVERT_NOT_CONVERTIBLE' && (
          <Alert
            type="warning"
            showIcon
            message={t('converter.messageNotConvertible')}
            description={resp.message}
          />
        )}
        {resp && resp.status === 'CONVERT_NOT_FOUND' && (
          <Alert
            type="error"
            showIcon
            message={t('converter.messageNotFound')}
            description={resp.message}
          />
        )}
      </Space>
    </Card>
  );
};

export default UnitConverter;
