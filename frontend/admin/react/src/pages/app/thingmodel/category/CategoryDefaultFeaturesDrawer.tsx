/**
 * 配置默认模型 Drawer（入口①）— 仅 level=4 细类调用。
 * Category default features drawer (entry ①).
 *
 * 内含 Tab(全部/属性/事件/服务/关系) + ProTable + 行内编辑 Override + + 新增特征（FeaturePicker 多选）。
 *
 * 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/05-前端实现设计.md §1.1
 */
import { useRef, useState } from 'react';
import { Drawer, Button, Space, Tabs, Popconfirm, App, Modal } from 'antd';
import { ProTable, type ActionType, type ProColumns } from '@ant-design/pro-components';
import { useTranslation } from 'react-i18next';
import { useQueryClient } from '@tanstack/react-query';
import {
  fetchListCategoryDefaultFeatures,
  useBatchAddCategoryDefaultFeatures,
  useDeleteCategoryDefaultFeature,
  useUpdateCategoryDefaultFeature,
} from '@/api/hooks/category-default-feature';
import { PaginationQuery } from '@/core/transport/rest';
import type {
  thingmodelservicev1_Category,
  thingmodelservicev1_CategoryDefaultFeature,
  thingmodelservicev1_FeatureOverrideSpec,
} from '@/api/generated/admin/service/v1';
import FeaturePicker from '../_shared/FeaturePicker';
import OverrideSpecForm from '../_shared/OverrideSpecForm';

interface Props {
  open: boolean;
  category: thingmodelservicev1_Category | null;
  onClose: () => void;
}

type TabKey = 'ALL' | 'PROPERTY' | 'EVENT' | 'SERVICE' | 'RELATION';

export const CategoryDefaultFeaturesDrawer = ({ open, category, onClose }: Props) => {
  const { t } = useTranslation(['category-default-feature', 'common']);
  const actionRef = useRef<ActionType>(null);
  const qc = useQueryClient();
  const { message } = App.useApp();

  const [tab, setTab] = useState<TabKey>('ALL');
  const [pickerOpen, setPickerOpen] = useState(false);
  const [editing, setEditing] = useState<thingmodelservicev1_CategoryDefaultFeature | null>(null);
  const [editOverride, setEditOverride] = useState<thingmodelservicev1_FeatureOverrideSpec | null>(
    null,
  );

  const { mutate: batchAdd } = useBatchAddCategoryDefaultFeatures({
    onSuccess: () => {
      message.success(t('addSuccess'));
      actionRef.current?.reload();
      qc.invalidateQueries({ queryKey: ['listCategoryDefaultFeatures'] });
    },
  });
  const { mutate: doDelete } = useDeleteCategoryDefaultFeature({
    onSuccess: () => {
      message.success(t('common:deleteSuccess'));
      actionRef.current?.reload();
    },
  });
  const { mutate: doUpdate } = useUpdateCategoryDefaultFeature({
    onSuccess: () => {
      message.success(t('common:updateSuccess'));
      actionRef.current?.reload();
      setEditing(null);
    },
  });

  const columns: ProColumns<thingmodelservicev1_CategoryDefaultFeature>[] = [
    { title: t('featureCode'), dataIndex: 'featureCode', width: 140 },
    { title: t('featureIdentifier'), dataIndex: 'featureIdentifier', width: 160 },
    { title: t('featureName'), dataIndex: 'featureName' },
    { title: t('featureType'), dataIndex: 'featureType', width: 100 },
    {
      title: t('overridden'),
      dataIndex: 'overrideSpec',
      width: 90,
      render: (_, row) => (row.overrideSpec ? t('yes') : '-'),
    },
    {
      title: t('common:table.action'),
      key: 'op',
      width: 160,
      render: (_, row) => (
        <Space size="small">
          <Button
            size="small"
            type="link"
            onClick={() => {
              setEditing(row);
              setEditOverride(row.overrideSpec ?? null);
            }}
          >
            {t('common:edit')}
          </Button>
          <Popconfirm
            title={t('common:deleteConfirm')}
            onConfirm={() => row.id && doDelete({ ids: [row.id] })}
          >
            <Button size="small" type="link" danger>
              {t('common:delete')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Drawer
      width={1024}
      open={open}
      onClose={onClose}
      title={category ? `${t('drawerTitle')} — ${category.name} (${category.code})` : ''}
      destroyOnClose
    >
      <Tabs
        activeKey={tab}
        onChange={(k) => setTab(k as TabKey)}
        items={[
          { key: 'ALL', label: t('tab.all') },
          { key: 'PROPERTY', label: t('tab.property') },
          { key: 'EVENT', label: t('tab.event') },
          { key: 'SERVICE', label: t('tab.service') },
          { key: 'RELATION', label: t('tab.relation') },
        ]}
      />

      <ProTable<thingmodelservicev1_CategoryDefaultFeature>
        actionRef={actionRef}
        rowKey="id"
        search={false}
        columns={columns}
        toolBarRender={() => [
          <Button key="add" type="primary" onClick={() => setPickerOpen(true)}>
            + {t('addFeature')}
          </Button>,
        ]}
        request={async (params) => {
          if (!category?.id) return { data: [], success: true, total: 0 };
          const formValues: Record<string, unknown> = {
            category_id: category.id,
          };
          if (tab !== 'ALL') formValues.feature_type = tab;
          const q = new PaginationQuery({
            paging: { page: params.current ?? 1, pageSize: params.pageSize ?? 20 },
            formValues,
            orderBy: ['sort_order', 'id'],
          });
          const resp = await fetchListCategoryDefaultFeatures(q);
          return {
            data: resp.items ?? [],
            success: true,
            total: Number(resp.total ?? 0),
          };
        }}
      />

      <FeaturePicker
        open={pickerOpen}
        onCancel={() => setPickerOpen(false)}
        onConfirm={(feats) => {
          setPickerOpen(false);
          if (!category?.id || feats.length === 0) return;
          batchAdd({
            categoryId: category.id,
            items: feats.map((f) => ({ featureId: f.id! })),
          });
        }}
      />

      <Modal
        open={!!editing}
        title={t('editOverride')}
        onCancel={() => setEditing(null)}
        onOk={() => {
          if (!editing) return;
          doUpdate({
            id: editing.id!,
            values: { overrideSpec: editOverride ?? undefined },
          });
        }}
      >
        <OverrideSpecForm
          featureType={editing?.featureType}
          value={editOverride}
          onChange={setEditOverride}
        />
      </Modal>
    </Drawer>
  );
};

export default CategoryDefaultFeaturesDrawer;
