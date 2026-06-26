import { useRef, useState, useEffect } from 'react';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { ProTable } from '@ant-design/pro-components';
import { Button, Popconfirm, Tag, App, Empty, Drawer } from 'antd';
import {
  EditOutlined,
  DeleteOutlined,
  PlusOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';

import { PaginationQuery } from '@/core';
import { TABLE } from '@/config/constants';
import { useProTableScrollY } from '@/hooks/useProTableScrollY';
import { fetchListUnits, useDeleteUnit } from '@/api/hooks/unit';
import {
  baseUnitTagColor,
  conversionTypeColor,
  conversionTypeOptions,
  enableBoolOptions,
  getConversionTypeLabel,
  getEnableColor,
  getEnableLabel,
} from './constants';
import UnitDrawer from './UnitDrawer';
import UnitConverter from './UnitConverter';

interface UnitListProps {
  categoryId: number | null;
  categoryName?: string;
}

/**
 * 单位列表（右侧） / Unit list (right panel).
 *
 * 当 categoryId 为空时显示 Empty 提示；否则按选中分类过滤列表，并支持：
 *   - 新建 / 编辑 / 删除（删除走 Popconfirm）
 *   - 打开换算器（顶部按钮 或 行内"换算"操作，传入该单位 id 预填源单位）
 */
const UnitList: React.FC<UnitListProps> = ({ categoryId, categoryName }) => {
  const { t } = useTranslation('unit');
  const actionRef = useRef<ActionType>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const tableScrollY = useProTableScrollY(containerRef);
  const queryClient = useQueryClient();
  const { message } = App.useApp();

  const [drawerOpen, setDrawerOpen] = useState(false);
  const [drawerMode, setDrawerMode] = useState<'create' | 'edit'>('create');
  const [editing, setEditing] = useState<any>(null);

  const [converterOpen, setConverterOpen] = useState(false);
  const [converterSourceId, setConverterSourceId] = useState<number | undefined>();

  // 当 categoryId 切换时刷新表
  useEffect(() => {
    if (categoryId) {
      actionRef.current?.reload();
    }
  }, [categoryId]);

  const deleteMutation = useDeleteUnit({
    onSuccess: () => {
      message.success(t('deleteSuccess'));
      actionRef.current?.reload();
      queryClient.invalidateQueries({ queryKey: ['listUnits'] });
      queryClient.invalidateQueries({ queryKey: ['listUnitsByCategory'] });
    },
    onError: (error: Error) => message.error(error.message || t('deleteFailed')),
  });

  const columns: ProColumns<any>[] = [
    {
      title: t('symbol'),
      dataIndex: 'symbol',
      width: 100,
      fixed: 'left',
    },
    {
      title: t('name'),
      dataIndex: 'name',
      width: 140,
    },
    {
      title: t('code'),
      dataIndex: 'code',
      width: 140,
    },
    {
      title: t('isBase'),
      dataIndex: 'isBase',
      width: 80,
      hideInSearch: true,
      render: (_, record) =>
        record.isBase ? <Tag color={baseUnitTagColor}>{t('baseBadge')}</Tag> : null,
    },
    {
      title: t('conversionType'),
      dataIndex: 'conversionType',
      width: 110,
      valueType: 'select',
      fieldProps: { options: conversionTypeOptions(t) },
      render: (_, record) => (
        <Tag color={conversionTypeColor[record.conversionType as keyof typeof conversionTypeColor] ?? 'default'}>
          {getConversionTypeLabel(t, record.conversionType)}
        </Tag>
      ),
    },
    {
      title: `${t('factor')} / ${t('offset')}`,
      width: 160,
      hideInSearch: true,
      render: (_, record) => (
        <span>
          x·{record.factor ?? 1}
          {record.offset && record.offset !== 0 ? `+${record.offset}` : ''}
        </span>
      ),
    },
    {
      title: t('precision'),
      dataIndex: 'precision',
      width: 80,
      hideInSearch: true,
    },
    {
      title: t('status'),
      dataIndex: 'isEnabled',
      width: 90,
      valueType: 'select',
      fieldProps: { options: enableBoolOptions(t) },
      render: (_, record) => (
        <Tag color={getEnableColor(record.isEnabled)}>
          {getEnableLabel(t, record.isEnabled)}
        </Tag>
      ),
    },
    {
      title: t('action'),
      valueType: 'option',
      width: 130,
      fixed: 'right',
      render: (_, record) => [
        <a
          key="edit"
          onClick={() => {
            setEditing(record);
            setDrawerMode('edit');
            setDrawerOpen(true);
          }}
        >
          <EditOutlined />
        </a>,
        <a
          key="convert"
          onClick={() => {
            setConverterSourceId(record.id);
            setConverterOpen(true);
          }}
          title={t('openConverter')}
        >
          <ThunderboltOutlined />
        </a>,
        <Popconfirm
          key="delete"
          title={t('deleteConfirmTitle')}
          description={t('deleteConfirmDesc', { moduleName: t('moduleName') })}
          onConfirm={() => record.id && deleteMutation.mutate({ ids: [record.id] })}
          okText={t('common:button.ok')}
          cancelText={t('common:button.cancel')}
        >
          <a style={{ color: '#ff4d4f' }}>
            <DeleteOutlined />
          </a>
        </Popconfirm>,
      ],
    },
  ];

  return (
    <>
      <div
        ref={containerRef}
        className="page-container-content"
        style={{ padding: '0 8px', height: '100%' }}
      >
        {categoryId ? (
          <ProTable<any>
            actionRef={actionRef}
            columns={columns}
            headerTitle={categoryName ? `${categoryName} · ${t('unitList')}` : t('unitList')}
            params={{ categoryId }}
            request={async (params) => {
              try {
                const query = new PaginationQuery({
                  paging: {
                    page: params.current || 1,
                    pageSize: params.pageSize || TABLE.DEFAULT_PAGE_SIZE,
                  },
                  formValues: {
                    ...Object.fromEntries(
                      Object.entries(params).filter(
                        ([key]) => !['current', 'pageSize', 'categoryId'].includes(key),
                      ),
                    ),
                    category_id: categoryId,
                  },
                });
                const response = await fetchListUnits(query);
                return {
                  data: response.items || [],
                  total: response.total || 0,
                  success: true,
                };
              } catch (error: any) {
                message.error(error.message || t('fetchFailed'));
                return { data: [], total: 0, success: false };
              }
            }}
            rowKey="id"
            search={{
              labelWidth: 'auto',
              defaultCollapsed: false,
            }}
            pagination={{
              defaultPageSize: TABLE.DEFAULT_PAGE_SIZE,
              showSizeChanger: true,
              showQuickJumper: true,
            }}
            toolBarRender={() => [
              <Button
                key="converter"
                icon={<ThunderboltOutlined />}
                size="small"
                onClick={() => {
                  setConverterSourceId(undefined);
                  setConverterOpen(true);
                }}
              >
                {t('openConverter')}
              </Button>,
              <Button
                key="create"
                type="primary"
                icon={<PlusOutlined />}
                size="small"
                onClick={() => {
                  setEditing(null);
                  setDrawerMode('create');
                  setDrawerOpen(true);
                }}
              >
                {t('create')}
              </Button>,
            ]}
            options={{
              density: true,
              fullScreen: true,
              setting: true,
              reload: true,
            }}
            size="small"
            bordered
            cardBordered={false}
            scroll={{ y: tableScrollY, x: 900 }}
          />
        ) : (
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%',
            }}
          >
            <Empty description={t('selectCategoryFirst')} />
          </div>
        )}
      </div>

      {categoryId && (
        <UnitDrawer
          open={drawerOpen}
          mode={drawerMode}
          data={editing}
          categoryId={categoryId}
          categoryName={categoryName}
          onClose={() => {
            setDrawerOpen(false);
            setEditing(null);
          }}
          onSuccess={() => actionRef.current?.reload()}
        />
      )}

      <Drawer
        title={t('converter.title')}
        open={converterOpen}
        onClose={() => setConverterOpen(false)}
        width={720}
        destroyOnClose
      >
        <UnitConverter
          defaultCategoryId={categoryId}
          defaultSourceUnitId={converterSourceId}
        />
      </Drawer>
    </>
  );
};

export default UnitList;
