/**
 * 产品详情页 / Product detail page.
 *
 * - 顶部：基本信息卡片 + 发布/取消发布/删除按钮
 * - 主体：Tab(属性/事件/服务/关系) 列表 + "+ 添加特征" 下拉（拉取默认 / 加全局 / 加本地）
 * - PUBLISHED 产品：添加/删除特征置灰，仅可改 override
 *
 * 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/05-前端实现设计.md §1.3
 */
import { useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Button,
  Card,
  Descriptions,
  Tabs,
  Space,
  Popconfirm,
  App,
  Modal,
  Tag,
  Dropdown,
} from 'antd';
import { ProTable, type ActionType, type ProColumns } from '@ant-design/pro-components';
import { useTranslation } from 'react-i18next';

import ContentContainer from '@/layouts/components/PageContainer/ContentContainer';
import {
  useGetProduct,
  usePublishProduct,
  useUnpublishProduct,
  useDeleteProduct,
} from '@/api/hooks/product';
import {
  fetchListProductFeatures,
  useDeleteProductFeature,
  usePullFromDefault,
} from '@/api/hooks/product-feature';
import { fetchListCategoryDefaultFeatures } from '@/api/hooks/category-default-feature';
import { PaginationQuery } from '@/core/transport/rest';
import type { thingmodelservicev1_ProductFeature } from '@/api/generated/admin/service/v1';
import FeatureSourceTag from '../_shared/FeatureSourceTag';
import ProductFeatureDrawer from './ProductFeatureDrawer';

type Tab = 'PROPERTY' | 'EVENT' | 'SERVICE' | 'RELATION';
type DrawerMode = 'edit' | 'create-local' | 'create-global';

const ProductDetailPage = () => {
  const { id } = useParams();
  const productId = Number(id);
  const { t } = useTranslation(['product', 'product-feature', 'common']);
  const navigate = useNavigate();
  const actionRef = useRef<ActionType>(null);
  const { message } = App.useApp();
  const [tab, setTab] = useState<Tab>('PROPERTY');
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [drawerMode, setDrawerMode] = useState<DrawerMode>('edit');
  const [editing, setEditing] = useState<thingmodelservicev1_ProductFeature | null>(null);

  const { data: product, refetch: refetchProduct } = useGetProduct({ id: productId });
  const { mutate: doPublish } = usePublishProduct();
  const { mutate: doUnpublish } = useUnpublishProduct();
  const { mutate: doDeleteProduct } = useDeleteProduct();
  const { mutate: doDeletePF } = useDeleteProductFeature({
    onSuccess: () => {
      message.success(t('common:deleteSuccess'));
      actionRef.current?.reload();
    },
    onError: (err) => message.error(err.message),
  });
  const { mutate: doPull } = usePullFromDefault({
    onSuccess: () => {
      message.success(t('product:pullSuccess'));
      actionRef.current?.reload();
    },
    onError: (err) => message.error(err.message),
  });

  const published = product?.status === 'PUBLISHED';

  const handlePublish = () => {
    Modal.confirm({
      title: t('product:confirmPublish'),
      content: t('product:publishHint'),
      onOk: () =>
        new Promise<void>((resolve, reject) => {
          doPublish(
            { id: productId },
            {
              onSuccess: () => {
                message.success(t('product:publishSuccess'));
                refetchProduct();
                resolve();
              },
              onError: (err) => {
                message.error(err.message);
                reject(err);
              },
            },
          );
        }),
    });
  };
  const handleUnpublish = () =>
    doUnpublish(
      { id: productId },
      {
        onSuccess: () => {
          message.success(t('product:unpublishSuccess'));
          refetchProduct();
        },
      },
    );
  const handleDelete = () =>
    doDeleteProduct(
      { ids: [productId] },
      {
        onSuccess: () => {
          message.success(t('common:deleteSuccess'));
          navigate('/thingmodel/product');
        },
      },
    );

  const handlePullMore = async () => {
    if (!product?.categoryId) return;
    const q = new PaginationQuery({
      paging: { page: 1, pageSize: 500 },
      formValues: { category_id: product.categoryId },
      orderBy: ['sort_order', 'id'],
    });
    const resp = await fetchListCategoryDefaultFeatures(q);
    const ids = (resp.items ?? []).map((x) => x.id!).filter(Boolean);
    if (ids.length === 0) {
      message.info(t('product:noDefaultModel'));
      return;
    }
    doPull({ productId, defaultFeatureIds: ids, onConflict: 'SKIP' });
  };

  const columns: ProColumns<thingmodelservicev1_ProductFeature>[] = [
    {
      title: t('product-feature:source'),
      dataIndex: 'source',
      width: 80,
      render: (_, row) => <FeatureSourceTag source={row.source as any} />,
    },
    { title: t('product-feature:code'), dataIndex: 'code', width: 160, copyable: true },
    { title: t('product-feature:identifier'), dataIndex: 'identifier', width: 160 },
    { title: t('product-feature:name'), dataIndex: 'name' },
    {
      title: t('product-feature:dataType'),
      dataIndex: 'dataType',
      width: 100,
      render: (v: any) => (v ? <Tag>{v}</Tag> : null),
    },
    { title: t('product-feature:accessMode'), dataIndex: 'accessMode', width: 90 },
    {
      title: t('product-feature:specSummary', { defaultValue: 'Spec 概要' }),
      width: 140,
      render: (_, row) => {
        // CR-001：用冗余特化列做摘要展示（spec 完整 JSON 太长不适合列表）
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
      },
    },
    {
      title: t('common:table.action'),
      width: 160,
      render: (_, row) => (
        <Space size="small">
          <Button
            size="small"
            type="link"
            onClick={() => {
              setEditing(row);
              setDrawerMode('edit');
              setDrawerOpen(true);
            }}
          >
            {t('common:edit')}
          </Button>
          {!published && (
            <Popconfirm
              title={t('common:deleteConfirm')}
              onConfirm={() => row.id && doDeletePF({ ids: [row.id] })}
            >
              <Button size="small" type="link" danger>
                {t('common:delete')}
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  if (!product) return null;

  return (
    <ContentContainer heightMode="fixed" padding="16px" bottomMargin={0}>
      <Space style={{ marginBottom: 12 }}>
        <Button onClick={() => navigate('/thingmodel/product')}>← {t('common:back')}</Button>
        <span style={{ fontSize: 18, fontWeight: 600 }}>{product.name}</span>
        {published ? (
          <Tag color="green">{t('product:statusMap.PUBLISHED')}</Tag>
        ) : (
          <Tag>{t('product:statusMap.DRAFT')}</Tag>
        )}
      </Space>
      <Space style={{ float: 'right' }}>
        {!published && (
          <Button type="primary" onClick={handlePublish}>
            {t('product:publish')}
          </Button>
        )}
        {published && <Button onClick={handleUnpublish}>{t('product:unpublish')}</Button>}
        <Popconfirm title={t('common:deleteConfirm')} onConfirm={handleDelete}>
          <Button danger>{t('common:delete')}</Button>
        </Popconfirm>
      </Space>
      <Card style={{ marginBottom: 12 }}>
        <Descriptions column={2} size="small">
          <Descriptions.Item label={t('product:code')}>{product.code}</Descriptions.Item>
          <Descriptions.Item label={t('product:category')}>
            {product.categoryName} ({product.categoryCode})
          </Descriptions.Item>
          <Descriptions.Item label={t('product:manufacturer')}>
            {product.manufacturer || '-'}
          </Descriptions.Item>
          <Descriptions.Item label={t('product:modelNo')}>
            {product.modelNo || '-'}
          </Descriptions.Item>
          <Descriptions.Item label={t('product:description')} span={2}>
            {product.description || '-'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Tabs
        activeKey={tab}
        onChange={(k) => setTab(k as Tab)}
        items={(['PROPERTY', 'EVENT', 'SERVICE', 'RELATION'] as Tab[]).map((tp) => ({
          key: tp,
          label: t(`product-feature:type.${tp}`),
        }))}
      />

      <ProTable<thingmodelservicev1_ProductFeature>
        actionRef={actionRef}
        rowKey="id"
        search={false}
        columns={columns}
        scroll={{ x: 1280 }}
        toolBarRender={() => [
          <Dropdown
            key="add"
            menu={{
              items: [
                { key: 'pull', label: t('product:addFrom.pull'), onClick: handlePullMore },
                {
                  key: 'global',
                  label: t('product:addFrom.global'),
                  onClick: () => {
                    setEditing(null);
                    setDrawerMode('create-global');
                    setDrawerOpen(true);
                  },
                },
                {
                  key: 'local',
                  label: t('product:addFrom.local'),
                  onClick: () => {
                    setEditing(null);
                    setDrawerMode('create-local');
                    setDrawerOpen(true);
                  },
                },
              ],
            }}
          >
            <Button type="primary" disabled={published}>
              + {t('product:addFeature')}
            </Button>
          </Dropdown>,
        ]}
        request={async (params) => {
          const q = new PaginationQuery({
            paging: { page: params.current ?? 1, pageSize: params.pageSize ?? 50 },
            formValues: {
              product_id: productId,
              feature_type: tab,
            },
            orderBy: ['sort_order', 'id'],
          });
          const resp = await fetchListProductFeatures(q);
          return {
            data: resp.items ?? [],
            success: true,
            total: Number(resp.total ?? 0),
          };
        }}
      />

      <ProductFeatureDrawer
        open={drawerOpen}
        productId={productId}
        feature={editing}
        mode={drawerMode}
        readonly={published && drawerMode === 'edit' ? 'partial' : false}
        onClose={(reload) => {
          setDrawerOpen(false);
          if (reload) actionRef.current?.reload();
        }}
      />
    </ContentContainer>
  );
};

export default ProductDetailPage;
