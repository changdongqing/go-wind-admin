/**
 * 全局特征库选择器 Modal / Global feature picker modal.
 *
 * 用于：
 *   - 分类默认模型 Drawer 内 "+ 新增特征" 按钮
 *   - 产品详情页 "添加 → 从全局特征库添加"
 *
 * 多选 + 按 type Tab + 关键字搜索 + 单页 50 条。
 * 过滤走项目通用 formValues 后缀范式（feature_type=equal / name___icontains）。
 *
 * 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/05-前端实现设计.md §5.1
 */
import { useEffect, useMemo, useState } from 'react';
import { Modal, Table, Tabs, Input, Tag } from 'antd';
import { useTranslation } from 'react-i18next';
import { useListFeatures } from '@/api/hooks/feature';
import { PaginationQuery } from '@/core/transport/rest';
import type { thingmodelservicev1_Feature } from '@/api/generated/admin/service/v1';

type FeatureType = 'PROPERTY' | 'EVENT' | 'SERVICE' | 'RELATION';

interface Props {
  open: boolean;
  excludedIds?: number[]; // 已存在的 feature_id，禁选
  onCancel: () => void;
  onConfirm: (features: thingmodelservicev1_Feature[]) => void;
  /** 仅允许选择某些类型；不传则全部允许 */
  allowedTypes?: FeatureType[];
}

const ALL_TYPES: FeatureType[] = ['PROPERTY', 'EVENT', 'SERVICE', 'RELATION'];

export const FeaturePicker = ({
  open,
  excludedIds,
  onCancel,
  onConfirm,
  allowedTypes,
}: Props) => {
  const { t } = useTranslation(['feature', 'common']);
  const types = allowedTypes ?? ALL_TYPES;
  const [activeType, setActiveType] = useState<FeatureType>(types[0]);
  const [keyword, setKeyword] = useState('');
  const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);
  const [selectedRows, setSelectedRows] = useState<thingmodelservicev1_Feature[]>([]);

  // Modal 关闭时清空状态
  useEffect(() => {
    if (!open) {
      setSelectedKeys([]);
      setSelectedRows([]);
      setKeyword('');
    }
  }, [open]);

  const query = useMemo(() => {
    const formValues: Record<string, unknown> = {
      feature_type: activeType,
    };
    if (keyword) {
      formValues.name___icontains = keyword;
    }
    return new PaginationQuery({
      paging: { page: 1, pageSize: 50 },
      formValues,
      orderBy: ['sort_order', 'id'],
    });
  }, [activeType, keyword]);

  const { data, isLoading } = useListFeatures(query, { enabled: open });

  const items = (data?.items ?? []).filter((f) => !excludedIds?.includes(f.id ?? 0));

  return (
    <Modal
      open={open}
      title={`${t('common:select')}${t('pageTitle')}`}
      width={760}
      onCancel={onCancel}
      onOk={() => {
        onConfirm(selectedRows);
      }}
      okButtonProps={{ disabled: selectedRows.length === 0 }}
      destroyOnHidden
    >
      <Tabs
        activeKey={activeType}
        onChange={(k) => {
          setActiveType(k as FeatureType);
          setSelectedKeys([]);
          setSelectedRows([]);
        }}
        items={types.map((tp) => ({
          key: tp,
          label: t(`featureTypeMap.${tp}`),
        }))}
      />
      <Input.Search
        allowClear
        placeholder={t('common:search')}
        onSearch={setKeyword}
        style={{ marginBottom: 12 }}
      />
      <Table
        rowKey="id"
        size="small"
        loading={isLoading}
        dataSource={items}
        rowSelection={{
          selectedRowKeys: selectedKeys,
          onChange: (keys, rows) => {
            setSelectedKeys(keys);
            setSelectedRows(rows);
          },
        }}
        pagination={false}
        scroll={{ y: 360 }}
        columns={[
          { title: t('code'), dataIndex: 'code', width: 140 },
          { title: t('identifier'), dataIndex: 'identifier', width: 160 },
          { title: t('name'), dataIndex: 'name' },
          {
            title: t('dataType'),
            dataIndex: 'dataType',
            width: 100,
            render: (v: string) => (v ? <Tag>{v}</Tag> : null),
          },
        ]}
      />
    </Modal>
  );
};

export default FeaturePicker;
