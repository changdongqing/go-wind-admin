/**
 * 产品管理列表 / Product list (entry ②).
 *
 * ProTable + 搜索（name/code/manufacturer/status）+ "+ 新增产品" 按钮（打开向导）。
 * 点击行进入 ProductDetailPage（路由 :id）。
 *
 * 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/05-前端实现设计.md §1.2
 */
import { useRef, useState } from 'react';
import { Button, Space, Popconfirm, App } from 'antd';
import { ProTable, type ActionType, type ProColumns } from '@ant-design/pro-components';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import ContentContainer from '@/layouts/components/PageContainer/ContentContainer';
import { useProTableScrollY } from '@/hooks/useProTableScrollY';
import { PaginationQuery } from '@/core/transport/rest';
import { fetchListProducts, useDeleteProduct } from '@/api/hooks/product';
import type { thingmodelservicev1_Product } from '@/api/generated/admin/service/v1';
import CreateProductWizard from './CreateProductWizard';
import { statusValueEnum, enabledValueEnum } from './constants';

const ProductManagement = () => {
  const { t } = useTranslation(['product', 'common']);
  const navigate = useNavigate();
  const actionRef = useRef<ActionType>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const tableScrollY = useProTableScrollY(containerRef);
  const { message } = App.useApp();
  const [wizardOpen, setWizardOpen] = useState(false);

  const { mutate: del } = useDeleteProduct({
    onSuccess: () => {
      message.success(t('common:deleteSuccess'));
      actionRef.current?.reload();
    },
    onError: (err: Error) => message.error(err.message),
  });

  const columns: ProColumns<thingmodelservicev1_Product>[] = [
    {
      title: t('code'),
      dataIndex: 'code',
      width: 180,
      copyable: true,
    },
    {
      title: t('name'),
      dataIndex: 'name',
    },
    {
      title: t('category'),
      dataIndex: 'categoryName',
      width: 220,
      hideInSearch: true,
      render: (_, row) => (
        <span>
          {row.categoryName} ({row.categoryCode})
        </span>
      ),
    },
    {
      title: t('manufacturer'),
      dataIndex: 'manufacturer',
      width: 120,
    },
    {
      title: t('modelNo'),
      dataIndex: 'modelNo',
      width: 140,
      hideInSearch: true,
    },
    {
      title: t('status'),
      dataIndex: 'status',
      width: 100,
      valueEnum: statusValueEnum,
    },
    {
      title: t('isEnabled'),
      dataIndex: 'isEnabled',
      width: 90,
      hideInSearch: true,
      valueEnum: enabledValueEnum,
    },
    {
      title: t('common:table.action'),
      key: 'op',
      width: 180,
      fixed: 'right',
      hideInSearch: true,
      render: (_, row) => (
        <Space size="small">
          <Button
            size="small"
            type="link"
            onClick={() => navigate(`/thingmodel/product/${row.id}`)}
          >
            {t('common:edit')}
          </Button>
          <Popconfirm
            title={t('common:deleteConfirm')}
            onConfirm={() => row.id && del({ ids: [row.id] })}
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
    <ContentContainer heightMode="fixed" padding="16px" bottomMargin={0}>
      <div ref={containerRef} style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
        <ProTable<thingmodelservicev1_Product>
          actionRef={actionRef}
          rowKey="id"
          columns={columns}
          scroll={{ y: tableScrollY, x: 1280 }}
          search={{ labelWidth: 'auto', defaultCollapsed: false }}
          request={async (params) => {
            const { current, pageSize, name, code, manufacturer, status } = params as Record<
              string,
              any
            >;

            const formValues: Record<string, unknown> = {};
            if (name) formValues.name___icontains = name;
            if (code) formValues.code___icontains = code;
            if (manufacturer) formValues.manufacturer___icontains = manufacturer;
            if (status) formValues.status = status;

            const q = new PaginationQuery({
              paging: { page: current, pageSize },
              formValues,
              orderBy: ['-created_at'],
            });
            const resp = await fetchListProducts(q);
            return {
              data: resp.items ?? [],
              total: Number(resp.total ?? 0),
              success: true,
            };
          }}
          toolBarRender={() => [
            <Button key="add" type="primary" onClick={() => setWizardOpen(true)}>
              {t('createProduct')}
            </Button>,
          ]}
        />
      </div>

      <CreateProductWizard
        open={wizardOpen}
        onClose={(reload) => {
          setWizardOpen(false);
          if (reload) actionRef.current?.reload();
        }}
      />
    </ContentContainer>
  );
};

export default ProductManagement;
