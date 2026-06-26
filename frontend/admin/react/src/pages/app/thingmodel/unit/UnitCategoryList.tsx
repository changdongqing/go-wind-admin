import { useRef, useState } from 'react';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { ProTable } from '@ant-design/pro-components';
import { Button, Popconfirm, Tag, App, theme } from 'antd';
import { EditOutlined, DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';

import { PaginationQuery } from '@/core';
import { TABLE } from '@/config/constants';
import { useProTableScrollY } from '@/hooks/useProTableScrollY';
import { fetchListUnitCategories, useDeleteUnitCategory } from '@/api/hooks/unit';
import { enableBoolOptions, getEnableColor, getEnableLabel } from './constants';
import UnitCategoryDrawer from './UnitCategoryDrawer';

interface UnitCategoryListProps {
  currentCategoryId: number | null;
  onCategorySelect: (id: number, name?: string) => void;
}

/**
 * 物理量分类列表（左侧） / Unit category list (left panel)
 *
 * 行为镜像 dict/DictTypeList：点击行高亮并通知父组件；删除走 Popconfirm。
 */
const UnitCategoryList: React.FC<UnitCategoryListProps> = ({
  currentCategoryId,
  onCategorySelect,
}) => {
  const { t } = useTranslation('unit-category');
  const { token } = theme.useToken();
  const actionRef = useRef<ActionType>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const tableScrollY = useProTableScrollY(containerRef);
  const queryClient = useQueryClient();
  const { message } = App.useApp();

  const [drawerOpen, setDrawerOpen] = useState(false);
  const [drawerMode, setDrawerMode] = useState<'create' | 'edit'>('create');
  const [editing, setEditing] = useState<any>(null);

  const deleteMutation = useDeleteUnitCategory({
    onSuccess: () => {
      message.success(t('deleteSuccess'));
      actionRef.current?.reload();
      queryClient.invalidateQueries({ queryKey: ['listUnitCategories'] });
    },
    onError: (error: Error) => message.error(error.message || t('deleteFailed')),
  });

  const columns: ProColumns<any>[] = [
    {
      title: t('name'),
      dataIndex: 'name',
      width: 130,
      fixed: 'left',
    },
    {
      title: t('code'),
      dataIndex: 'code',
      width: 140,
    },
    {
      title: t('baseUnitSymbol'),
      dataIndex: 'baseUnitSymbol',
      width: 110,
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
      width: 80,
      fixed: 'right',
      render: (_, record) => [
        <a
          key="edit"
          onClick={(e) => {
            e.stopPropagation();
            setEditing(record);
            setDrawerMode('edit');
            setDrawerOpen(true);
          }}
        >
          <EditOutlined />
        </a>,
        <Popconfirm
          key="delete"
          title={t('deleteConfirmTitle')}
          description={t('deleteConfirmDesc', { moduleName: t('moduleName') })}
          onConfirm={(e) => {
            e?.stopPropagation();
            record.id && deleteMutation.mutate({ ids: [record.id] });
          }}
          onCancel={(e) => e?.stopPropagation()}
          okText={t('common:button.ok')}
          cancelText={t('common:button.cancel')}
        >
          <a style={{ color: '#ff4d4f' }} onClick={(e) => e.stopPropagation()}>
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
        <ProTable<any>
          actionRef={actionRef}
          columns={columns}
          headerTitle={t('categoryList')}
          request={async (params) => {
            try {
              const query = new PaginationQuery({
                paging: {
                  page: params.current || 1,
                  pageSize: params.pageSize || TABLE.DEFAULT_PAGE_SIZE,
                },
                formValues: Object.fromEntries(
                  Object.entries(params).filter(
                    ([key]) => !['current', 'pageSize'].includes(key),
                  ),
                ),
              });
              const response = await fetchListUnitCategories(query);
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
            span: 24,
          }}
          pagination={{
            defaultPageSize: TABLE.DEFAULT_PAGE_SIZE,
            showSizeChanger: true,
            showQuickJumper: true,
          }}
          toolBarRender={() => [
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
            density: false,
            fullScreen: false,
            setting: false,
            reload: true,
          }}
          size="small"
          bordered
          cardBordered={false}
          scroll={{ y: tableScrollY, x: 500 }}
          rowClassName={(record) =>
            record.id === currentCategoryId ? 'unit-category-row-selected' : ''
          }
          onRow={(record) => {
            const isSelected = record.id === currentCategoryId;
            return {
              onClick: () => onCategorySelect(record.id, record.name),
              style: {
                cursor: 'pointer',
                // 通过 CSS 自定义属性把 antd token 颜色值注入到 <tr>，
                // Less 中用 var(--unit-selected-bg) 读取，确保主题切换时颜色正确
                '--unit-selected-bg': isSelected ? token.colorPrimaryBgHover : undefined,
                '--unit-selected-bg-hover': isSelected ? token.colorPrimaryActive : undefined,
              } as React.CSSProperties,
            };
          }}
        />
      </div>

      <UnitCategoryDrawer
        open={drawerOpen}
        mode={drawerMode}
        data={editing}
        onClose={() => {
          setDrawerOpen(false);
          setEditing(null);
        }}
        onSuccess={() => actionRef.current?.reload()}
      />
    </>
  );
};

export default UnitCategoryList;
