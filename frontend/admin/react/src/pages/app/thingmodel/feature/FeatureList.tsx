import { useRef, useState, useEffect } from 'react';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { ProTable } from '@ant-design/pro-components';
import { Button, Popconfirm, Tag, App } from 'antd';
import { EditOutlined, DeleteOutlined, PlusOutlined, ImportOutlined } from '@ant-design/icons';
import { useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';

import { PaginationQuery } from '@/core';
import { TABLE } from '@/config/constants';
import { useProTableScrollY } from '@/hooks/useProTableScrollY';
import { fetchListFeatures, useDeleteFeature } from '@/api/hooks/feature';
import {
  enableBoolOptions,
  featureTypeColor,
  getEnableColor,
  getEnableLabel,
  type FeatureType,
} from './constants';
import FeatureDrawer from './FeatureDrawer';
import ImportFeaturesModal from './ImportFeaturesModal';

interface FeatureListProps {
  featureType: FeatureType;
}

/**
 * 特征列表（右侧）/ Feature list (right panel).
 *
 * CR-001（2026-06-29）：thing_features 不再含 spec / 5 个特化列，本列表瘦身：
 *   - 删除按 featureType 切换的 dataType / accessMode / eventLevel / callMode / relationType / source/target 列；
 *   - 新增 semanticTag 列，便于按语义检索；
 *   - applicableScope 与启用状态列保留。
 */
const FeatureList: React.FC<FeatureListProps> = ({ featureType }) => {
  const { t } = useTranslation('feature');
  const actionRef = useRef<ActionType>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const tableScrollY = useProTableScrollY(containerRef);
  const queryClient = useQueryClient();
  const { message } = App.useApp();

  const [drawerOpen, setDrawerOpen] = useState(false);
  const [drawerMode, setDrawerMode] = useState<'create' | 'edit'>('create');
  const [editing, setEditing] = useState<any>(null);
  const [importOpen, setImportOpen] = useState(false);

  useEffect(() => {
    actionRef.current?.reload();
  }, [featureType]);

  const deleteMutation = useDeleteFeature({
    onSuccess: () => {
      message.success(t('deleteSuccess'));
      actionRef.current?.reload();
      queryClient.invalidateQueries({ queryKey: ['listFeatures'] });
      queryClient.invalidateQueries({ queryKey: ['listFeaturesByType'] });
    },
    onError: (error: Error) => {
      // 后端 FEATURE_IN_USE_CANNOT_DELETE 返回时给特化提示
      const msg = error.message || t('deleteFailed');
      if (msg.includes('referenced by relation')) {
        message.warning(t('referencedCannotDelete'));
      } else {
        message.error(msg);
      }
    },
  });

  // 公共列 / Common columns
  const baseColumns: ProColumns<any>[] = [
    { title: t('code'), dataIndex: 'code', width: 140, fixed: 'left' },
    { title: t('name'), dataIndex: 'name', width: 140 },
    { title: t('identifier'), dataIndex: 'identifier', width: 160 },
    {
      title: t('semanticTag', { defaultValue: '语义标签' }),
      dataIndex: 'semanticTag',
      width: 130,
    },
  ];

  const tailColumns: ProColumns<any>[] = [
    { title: t('applicableScope'), dataIndex: 'applicableScope', width: 140 },
    {
      title: t('status'),
      dataIndex: 'isEnabled',
      width: 90,
      valueType: 'select',
      fieldProps: { options: enableBoolOptions(t) },
      render: (_, r) => (
        <Tag color={getEnableColor(r.isEnabled)}>{getEnableLabel(t, r.isEnabled)}</Tag>
      ),
    },
    {
      title: t('action'),
      valueType: 'option',
      width: 110,
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
        <Popconfirm
          key="delete"
          title={t('deleteConfirmTitle')}
          description={t('deleteConfirmDesc')}
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

  const columns = [...baseColumns, ...tailColumns];

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
          headerTitle={
            <span>
              <Tag color={featureTypeColor[featureType]}>{t(`featureTypeMap.${featureType}`)}</Tag>
              {t('featureList')}
            </span>
          }
          params={{ featureType }}
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
                      ([key]) => !['current', 'pageSize', 'featureType'].includes(key),
                    ),
                  ),
                  feature_type: featureType,
                },
              });
              const response = await fetchListFeatures(query);
              return {
                data: response.items || [],
                total: response.total ? Number(response.total) : 0,
                success: true,
              };
            } catch (error: any) {
              message.error(error.message || t('fetchFailed'));
              return { data: [], total: 0, success: false };
            }
          }}
          rowKey="id"
          search={{ labelWidth: 'auto', defaultCollapsed: false }}
          pagination={{
            defaultPageSize: TABLE.DEFAULT_PAGE_SIZE,
            showSizeChanger: true,
            showQuickJumper: true,
          }}
          toolBarRender={() => [
            <Button
              key="import"
              icon={<ImportOutlined />}
              size="small"
              onClick={() => setImportOpen(true)}
            >
              {t('import')}
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
          options={{ density: true, fullScreen: true, setting: true, reload: true }}
          size="small"
          bordered
          cardBordered={false}
          scroll={{ y: tableScrollY, x: 1100 }}
        />
      </div>

      <FeatureDrawer
        open={drawerOpen}
        mode={drawerMode}
        featureType={featureType}
        data={editing}
        onClose={() => {
          setDrawerOpen(false);
          setEditing(null);
        }}
        onSuccess={() => actionRef.current?.reload()}
      />

      <ImportFeaturesModal
        open={importOpen}
        onClose={() => setImportOpen(false)}
        onSuccess={() => {
          actionRef.current?.reload();
          queryClient.invalidateQueries({ queryKey: ['listFeatures'] });
          queryClient.invalidateQueries({ queryKey: ['listFeaturesByType'] });
        }}
      />
    </>
  );
};

export default FeatureList;
