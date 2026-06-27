import { useRef, useState, useEffect } from 'react';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { ProTable } from '@ant-design/pro-components';
import { Button, Popconfirm, Tag, App } from 'antd';
import { EditOutlined, DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';

import { PaginationQuery } from '@/core';
import { TABLE } from '@/config/constants';
import { useProTableScrollY } from '@/hooks/useProTableScrollY';
import { fetchListFeatures, useDeleteFeature } from '@/api/hooks/feature';
import {
  accessModeColor,
  callModeColor,
  dataTypeOptions,
  enableBoolOptions,
  eventLevelColor,
  featureTypeColor,
  getEnableColor,
  getEnableLabel,
  type FeatureType,
} from './constants';
import FeatureDrawer from './FeatureDrawer';

interface FeatureListProps {
  featureType: FeatureType;
}

/**
 * 特征列表（右侧）/ Feature list (right panel).
 *
 * 列随 `featureType` 动态变化（property 显示数据类型/访问模式；event 显示级别；service 显示调用模式；relation 显示关系类型/source/target）。
 * Columns vary by feature_type.
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
  ];

  // 按 featureType 动态列 / Type-specific columns
  const typeColumns: ProColumns<any>[] = (() => {
    switch (featureType) {
      case 'PROPERTY':
        return [
          {
            title: t('dataType'),
            dataIndex: 'dataType',
            width: 100,
            valueType: 'select',
            fieldProps: { options: dataTypeOptions(t) },
            render: (_, r) => (r.dataType ? <Tag>{t(`dataTypeMap.${r.dataType}`)}</Tag> : null),
          },
          {
            title: t('accessMode'),
            dataIndex: 'accessMode',
            width: 100,
            hideInSearch: true,
            render: (_, r) =>
              r.accessMode ? (
                <Tag color={accessModeColor[r.accessMode as 'R' | 'RW'] ?? 'default'}>
                  {t(`accessModeMap.${r.accessMode}`)}
                </Tag>
              ) : null,
          },
        ];
      case 'EVENT':
        return [
          {
            title: t('eventLevel'),
            dataIndex: 'eventLevel',
            width: 100,
            hideInSearch: true,
            render: (_, r) =>
              r.eventLevel ? (
                <Tag color={eventLevelColor[r.eventLevel as keyof typeof eventLevelColor] ?? 'default'}>
                  {t(`eventLevelMap.${r.eventLevel}`)}
                </Tag>
              ) : null,
          },
        ];
      case 'SERVICE':
        return [
          {
            title: t('callMode'),
            dataIndex: 'callMode',
            width: 100,
            hideInSearch: true,
            render: (_, r) =>
              r.callMode ? (
                <Tag color={callModeColor[r.callMode as keyof typeof callModeColor] ?? 'default'}>
                  {t(`callModeMap.${r.callMode}`)}
                </Tag>
              ) : null,
          },
        ];
      case 'RELATION':
        return [
          {
            title: t('relationType'),
            dataIndex: 'relationType',
            width: 130,
            hideInSearch: true,
            render: (_, r) =>
              r.relationType ? (
                <Tag>{t(`relationTypeMap.${r.relationType}`, { defaultValue: r.relationType })}</Tag>
              ) : null,
          },
          {
            title: `${t('source')} → ${t('target')}`,
            width: 220,
            hideInSearch: true,
            render: (_, r) => {
              const src = r.spec?.relation?.source;
              const tgt = r.spec?.relation?.target;
              const srcLabel = src?.identifier || src?.code || src?.id;
              const tgtLabel = tgt?.identifier || tgt?.code || tgt?.id;
              if (!srcLabel && !tgtLabel) return null;
              return (
                <span style={{ fontSize: 12, color: '#666' }}>
                  {srcLabel ?? '-'} → {tgtLabel ?? '-'}
                </span>
              );
            },
          },
        ];
      default:
        return [];
    }
  })();

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

  const columns = [...baseColumns, ...typeColumns, ...tailColumns];

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
    </>
  );
};

export default FeatureList;
