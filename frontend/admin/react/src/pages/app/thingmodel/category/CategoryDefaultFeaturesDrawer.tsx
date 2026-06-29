/**
 * 配置默认模型 Drawer（入口①）— 仅 level=4 细类调用。
 * Category default features drawer (entry ①).
 *
 * 内含 Tab(全部/属性/事件/服务/关系) + ProTable + 编辑 Spec(完整)Modal + 新增特征(FeaturePicker 多选)。
 *
 * 设计依据 / Design ref:
 *   - docs/thingmodel/sheji/模型管理/05-前端实现设计.md §1.1
 *   - docs/thingmodel/sheji/修改记录/CR-001-结构化约束下沉到模型层.md
 *
 * CR-001（2026-06-29）：
 *   - "编辑覆写"按钮改为"编辑约束 spec"，弹出 Modal 内挂载完整 FeatureSpecForm；
 *   - "overridden(yes|no)" 列改为冗余特化列 dataType/eventLevel/callMode/relationType 展示。
 */
import { useEffect, useRef, useState } from 'react';
import { Drawer, Button, Form, Space, Tabs, Popconfirm, App, Modal, Tag } from 'antd';
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
  thingmodelservicev1_FeatureSpec,
} from '@/api/generated/admin/service/v1';
import FeaturePicker from '../_shared/FeaturePicker';
import FeatureSpecForm from '../_shared/specs/FeatureSpecForm';

interface Props {
  open: boolean;
  category: thingmodelservicev1_Category | null;
  onClose: () => void;
}

type TabKey = 'ALL' | 'PROPERTY' | 'EVENT' | 'SERVICE' | 'RELATION';

// feature_type → Tag 颜色（与特征管理保持一致）
const featureTypeColor: Record<string, string> = {
  PROPERTY: 'blue',
  EVENT: 'orange',
  SERVICE: 'green',
  RELATION: 'purple',
};

export const CategoryDefaultFeaturesDrawer = ({ open, category, onClose }: Props) => {
  const { t } = useTranslation(['category-default-feature', 'common']);
  const actionRef = useRef<ActionType>(null);
  const qc = useQueryClient();
  const { message } = App.useApp();

  const [tab, setTab] = useState<TabKey>('ALL');
  const [pickerOpen, setPickerOpen] = useState(false);
  const [editing, setEditing] = useState<thingmodelservicev1_CategoryDefaultFeature | null>(null);
  const [specForm] = Form.useForm<{ spec: thingmodelservicev1_FeatureSpec | undefined }>();

  // 注：editing 变化时不在此处 setFieldsValue —— 改在 Modal afterOpenChange 中执行，
  // 等子表单（destroyOnHidden 重新挂载）完成 register 后再回填，避免值丢失。
  useEffect(() => {
    if (open) {
      actionRef.current?.reload();
    }
  }, [tab, open]);

  const { mutate: batchAdd } = useBatchAddCategoryDefaultFeatures({
    onSuccess: () => {
      message.success(t('addSuccess'));
      actionRef.current?.reload();
      qc.invalidateQueries({ queryKey: ['listCategoryDefaultFeatures'] });
    },
  });
  const { mutate: doDelete } = useDeleteCategoryDefaultFeature({
    onSuccess: () => {
      message.success(t('common:button.delete'));
      actionRef.current?.reload();
    },
  });
  const { mutate: doUpdate } = useUpdateCategoryDefaultFeature({
    onSuccess: () => {
      message.success(t('common:button.ok'));
      actionRef.current?.reload();
      setEditing(null);
    },
  });

  const renderTypeSummary = (row: thingmodelservicev1_CategoryDefaultFeature) => {
    if (!row.featureType) return '-';
    switch (row.featureType) {
      case 'PROPERTY':
        return [row.dataType, row.accessMode].filter(Boolean).join(' / ') || '-';
      case 'EVENT':
        return row.eventLevel || '-';
      case 'SERVICE':
        return row.callMode || '-';
      case 'RELATION':
        return row.relationType || '-';
      default:
        return '-';
    }
  };

  const columns: ProColumns<thingmodelservicev1_CategoryDefaultFeature>[] = [
    {
      title: t('featureCode'),
      dataIndex: 'featureCode',
      width: 140,
      render: (_, row) => row.featureCode || '-',
    },
    {
      title: t('featureIdentifier'),
      dataIndex: 'featureIdentifier',
      width: 160,
      render: (_, row) => row.featureIdentifier || '-',
    },
    {
      title: t('featureName'),
      dataIndex: 'featureName',
      render: (_, row) => row.featureName || row.displayName || '-',
    },
    {
      title: t('featureType'),
      dataIndex: 'featureType',
      width: 100,
      render: (_, row) =>
        row.featureType ? (
          <Tag color={featureTypeColor[row.featureType] ?? 'default'}>
            {t(`featureTypeMap.${row.featureType}`, row.featureType)}
          </Tag>
        ) : (
          '-'
        ),
    },
    {
      title: t('specSummary', { defaultValue: 'Spec 概要' }),
      key: 'specSummary',
      width: 140,
      render: (_, row) => renderTypeSummary(row),
    },
    {
      title: t('actionTitle'),
      key: 'op',
      width: 160,
      render: (_, row) => (
        <Space size="small">
          <Button size="small" type="link" onClick={() => setEditing(row)}>
            {t('editSpec', { defaultValue: '编辑 Spec' })}
          </Button>
          <Popconfirm
            title={t('deleteConfirm')}
            onConfirm={() => row.id && doDelete({ ids: [row.id] })}
          >
            <Button size="small" type="link" danger>
              {t('common:button.delete')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Drawer
      size={1024}
      open={open}
      onClose={onClose}
      title={category ? `${t('drawerTitle')} — ${category.name} (${category.code})` : ''}
      destroyOnHidden
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
        params={{ tab }}
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
        title={t('editSpec', { defaultValue: '编辑约束 Spec' })}
        width={760}
        onCancel={() => setEditing(null)}
        destroyOnHidden
        afterOpenChange={(open) => {
          // 在 Modal 完全打开（子组件已挂载）后再 setFieldsValue，
          // 避免 destroyOnHidden + 异步挂载导致子 Form.Item 还没 register 时丢值。
          if (open && editing) {
            specForm.resetFields();
            specForm.setFieldsValue({ spec: editing.spec });
            // eslint-disable-next-line no-console
            console.log('[CDF-DBG] open editing.id=', editing.id, ' editing.spec=', editing.spec);
          }
        }}
        onOk={async () => {
          if (!editing) return;
          const values = await specForm.validateFields();
          // eslint-disable-next-line no-console
          console.log('[CDF-DBG] submit values=', values);
          doUpdate({
            id: editing.id!,
            values: { spec: values.spec },
          });
        }}
      >
        <Form form={specForm} layout="vertical">
          <FeatureSpecForm featureType={editing?.featureType} />
        </Form>
      </Modal>
    </Drawer>
  );
};

export default CategoryDefaultFeaturesDrawer;
