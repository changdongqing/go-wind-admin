/**
 * 物模型-分类管理页面（极简版 v2.1）
 * Thing-model category management page (minimalist v2.1)
 *
 * 设计依据 / Design ref: docs/thingmodel/sheji/分类管理/05-前端实现设计.md
 *
 * 形态：单一 ProTable + Drawer。kind 是搜索表单里的一个普通过滤项，不用 Tabs/Splitter/Tree。
 * 默认按 (kind+code) 升序展示——code 自带层级前缀，多 kind 自然分块、单 kind 内层级清晰。
 */
import { useRef, useState } from 'react';
import { Button, Popconfirm, Tag, Space, App } from 'antd';
import { ProTable, type ActionType, type ProColumns } from '@ant-design/pro-components';
import { useTranslation } from 'react-i18next';
import { useQueryClient } from '@tanstack/react-query';

import ContentContainer from '@/layouts/components/PageContainer/ContentContainer';
import { useProTableScrollY } from '@/hooks/useProTableScrollY';
import { PaginationQuery } from '@/core';
import { fetchListCategories, useDeleteCategory } from '@/api/hooks/category';
import type { thingmodelservicev1_Category } from '@/api/generated/admin/service/v1';
import CategoryDrawer from './CategoryDrawer';
import CategoryDefaultFeaturesDrawer from './CategoryDefaultFeaturesDrawer';
import {
  kindValueEnum,
  levelValueEnum,
  levelColor,
  kindColor,
  enableValueEnum,
} from './constants';

const CategoryManagement = () => {
  const { t } = useTranslation('category');
  const actionRef = useRef<ActionType>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const tableScrollY = useProTableScrollY(containerRef);
  const queryClient = useQueryClient();
  const { message } = App.useApp();

  const [drawerOpen, setDrawerOpen] = useState(false);
  const [drawerMode, setDrawerMode] = useState<'create' | 'edit'>('create');
  const [editing, setEditing] = useState<thingmodelservicev1_Category | null>(null);
  const [parentNode, setParentNode] = useState<thingmodelservicev1_Category | null>(null);

  // 分类默认模型 Drawer 状态（模型管理入口①，仅 level=4 行触发）
  // Category default features drawer state (model management entry ①, triggered on level=4 rows)
  const [cdfOpen, setCdfOpen] = useState(false);
  const [cdfCategory, setCdfCategory] = useState<thingmodelservicev1_Category | null>(null);
  const openCdfDrawer = (row: thingmodelservicev1_Category) => {
    setCdfCategory(row);
    setCdfOpen(true);
  };

  const { mutate: del } = useDeleteCategory({
    onSuccess: () => {
      message.success(t('deleteSuccess'));
      actionRef.current?.reload();
      queryClient.invalidateQueries({ queryKey: ['listCategories'] });
    },
    onError: (err: Error) => message.error(err.message || t('deleteFailed')),
  });

  const openCreateTop = () => {
    setDrawerMode('create');
    setEditing(null);
    setParentNode(null);
    setDrawerOpen(true);
  };

  const openCreateChild = (row: thingmodelservicev1_Category) => {
    setDrawerMode('create');
    setEditing(null);
    setParentNode(row);
    setDrawerOpen(true);
  };

  const openEdit = (row: thingmodelservicev1_Category) => {
    setDrawerMode('edit');
    setEditing(row);
    setParentNode(null);
    setDrawerOpen(true);
  };

  const handleDrawerClose = (refresh?: boolean) => {
    setDrawerOpen(false);
    setEditing(null);
    setParentNode(null);
    if (refresh) {
      actionRef.current?.reload();
      queryClient.invalidateQueries({ queryKey: ['listCategories'] });
    }
  };

  const columns: ProColumns<thingmodelservicev1_Category>[] = [
    // 种类（搜索表单第 1 项）
    {
      title: t('fields.kind'),
      dataIndex: 'kind',
      width: 110,
      valueType: 'select',
      valueEnum: kindValueEnum(t),
      render: (_, row) =>
        row.kind ? (
          <Tag color={kindColor[row.kind] ?? 'default'}>{t(`kind.${row.kind}`)}</Tag>
        ) : (
          '-'
        ),
    },

    // 编码（带缩进，直观表达层级）
    {
      title: t('fields.code'),
      dataIndex: 'code',
      width: 200,
      fixed: 'left',
      // ProTable 默认会把搜索值塞进 formValues；前端把它作为 code___starts_with 前缀过滤
      render: (_, row) => {
        const lvl = row.level ?? 1;
        const indent = '\u00A0\u00A0\u00A0\u00A0'.repeat(lvl - 1);
        return (
          <span style={{ fontFamily: 'monospace' }}>
            {indent}
            {row.code}
          </span>
        );
      },
    },

    // 名称
    {
      title: t('fields.name'),
      dataIndex: 'name',
      width: 240,
      ellipsis: true,
    },

    // 英文名
    {
      title: t('fields.nameEn'),
      dataIndex: 'nameEn',
      width: 200,
      hideInSearch: true,
      ellipsis: true,
    },

    // 层级
    {
      title: t('fields.level'),
      dataIndex: 'level',
      width: 90,
      valueType: 'select',
      valueEnum: levelValueEnum(t),
      render: (_, row) => (
        <Tag color={levelColor[row.level ?? 1] ?? 'default'}>{t(`level.${row.level ?? 1}`)}</Tag>
      ),
    },

    // 状态
    {
      title: t('fields.isEnabled'),
      dataIndex: 'isEnabled',
      width: 90,
      valueType: 'select',
      valueEnum: enableValueEnum(t),
    },

    {
      title: t('fields.referenceCount'),
      dataIndex: 'referenceCount',
      width: 100,
      hideInSearch: true,
    },
    {
      title: t('fields.sortOrder'),
      dataIndex: 'sortOrder',
      width: 90,
      hideInSearch: true,
    },

    {
      title: t('common:table.action', '操作'),
      key: 'action',
      width: 320,
      fixed: 'right',
      hideInSearch: true,
      render: (_, row) => (
        <Space>
          {(row.level ?? 1) < 4 && (
            <Button type="link" size="small" onClick={() => openCreateChild(row)}>
              {t('actions.createChild')}
            </Button>
          )}
          {/* 仅 level=4 细类显示"配置默认模型"（模型管理入口①）*/}
          {(row.level ?? 1) === 4 && (
            <Button type="link" size="small" onClick={() => openCdfDrawer(row)}>
              {t('actions.configDefaultModel', '配置默认模型')}
            </Button>
          )}
          <Button type="link" size="small" onClick={() => openEdit(row)}>
            {t('actions.edit')}
          </Button>
          <Popconfirm
            title={t('tips.deleteConfirm')}
            onConfirm={() => row.id && del({ ids: [row.id] })}
          >
            <Button type="link" size="small" danger>
              {t('actions.delete')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <ContentContainer heightMode="fixed" padding="16px" bottomMargin={0}>
      <div ref={containerRef} style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
        <ProTable<thingmodelservicev1_Category>
          actionRef={actionRef}
          columns={columns}
          rowKey="id"
          search={{ labelWidth: 'auto', defaultCollapsed: false }}
          scroll={{ y: tableScrollY, x: 1280 }}
          request={async (params) => {
            // params 中含 kind / level / code / name / isEnabled 等搜索表单值 + current/pageSize
            const { current, pageSize, kind, level, code, name, isEnabled } = params as Record<
              string,
              any
            >;

            const formValues: Record<string, unknown> = {};
            if (kind) formValues.kind = kind;
            if (level !== undefined && level !== null && level !== '') formValues.level = Number(level);
            if (code) formValues.code___starts_with = code;
            if (name) formValues.name___icontains = name;
            if (isEnabled !== undefined && isEnabled !== '') {
              formValues.is_enabled = isEnabled === true || isEnabled === 'true';
            }

            const query = new PaginationQuery({
              paging: { page: current, pageSize },
              formValues,
              // 默认按 kind+code 升序——code 自带层级前缀，天然 = 深度优先遍历
              orderBy: ['kind', 'code'],
            });

            const res = await fetchListCategories(query);
            return {
              data: res.items ?? [],
              total: Number(res.total ?? 0),
              success: true,
            };
          }}
          toolBarRender={() => [
            <Button key="add" type="primary" onClick={openCreateTop}>
              {t('actions.create')}
            </Button>,
          ]}
        />
      </div>

      <CategoryDrawer
        open={drawerOpen}
        mode={drawerMode}
        data={editing}
        parent={parentNode}
        onClose={handleDrawerClose}
      />

      <CategoryDefaultFeaturesDrawer
        open={cdfOpen}
        category={cdfCategory}
        onClose={() => setCdfOpen(false)}
      />
    </ContentContainer>
  );
};

export default CategoryManagement;
