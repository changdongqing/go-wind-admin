import { useState } from 'react';
import { Button, Card, Form, Input, Select, Space, Table, Tag, message, Popconfirm } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useListTenants, useDeleteTenant } from '@/api/hooks/tenant';
import type { identityservicev1_Tenant } from '@/api/generated/admin/service/v1';
import { PaginationQuery } from '@/core';
import { EditOutlined, DeleteOutlined } from '@ant-design/icons';

/**
 * 租户列表页面
 */
const TenantList = () => {
  const [form] = Form.useForm();
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 10,
  });

  // 获取租户列表
  const listTenantsMutation = useListTenants({
    onSuccess: (data) => {
      console.log('租户列表:', data);
    },
    onError: () => {
      message.error('获取租户列表失败');
    },
  });

  // 删除租户
  const deleteTenantMutation = useDeleteTenant({
    onSuccess: () => {
      message.success('删除成功');
      // 重新加载列表
      handleSearch();
    },
    onError: () => {
      message.error('删除失败');
    },
  });

  // 搜索
  const handleSearch = () => {
    const values = form.getFieldsValue();
    const query = new PaginationQuery({
      page: 1,
      pageSize: pagination.pageSize,
      ...values,
    });
    listTenantsMutation.mutate(query);
  };

  // 重置
  const handleReset = () => {
    form.resetFields();
    setPagination({ page: 1, pageSize: 10 });
    handleSearch();
  };

  // 分页变化
  const handleTableChange = (newPagination: any) => {
    setPagination({
      page: newPagination.current,
      pageSize: newPagination.pageSize,
    });
    const values = form.getFieldsValue();
    const query = new PaginationQuery({
      page: newPagination.current,
      pageSize: newPagination.pageSize,
      ...values,
    });
    listTenantsMutation.mutate(query);
  };

  // 删除租户
  const handleDelete = (id: number) => {
    deleteTenantMutation.mutate({ id });
  };

  // 表格列定义
  const columns: ColumnsType<identityservicev1_Tenant> = [
    {
      title: '序号',
      dataIndex: 'id',
      width: 80,
      render: (_: any, __: any, index: number) => {
        return (pagination.page - 1) * pagination.pageSize + index + 1;
      },
    },
    {
      title: '租户名称',
      dataIndex: 'name',
      width: 150,
    },
    {
      title: '租户编码',
      dataIndex: 'code',
      width: 150,
    },
    {
      title: '管理员用户名',
      dataIndex: 'adminUsername',
      width: 150,
    },
    {
      title: '租户类型',
      dataIndex: 'tenantType',
      width: 100,
      render: (type: number) => {
        const typeMap: Record<number, { text: string; color: string }> = {
          0: { text: '免费', color: 'default' },
          1: { text: '付费', color: 'success' },
        };
        const config = typeMap[type] || { text: '未知', color: 'default' };
        return <Tag color={config.color}>{config.text}</Tag>;
      },
    },
    {
      title: '审核状态',
      dataIndex: 'auditStatus',
      width: 100,
      render: (status: number) => {
        const statusMap: Record<number, { text: string; color: string }> = {
          0: { text: '待审核', color: 'warning' },
          1: { text: '已通过', color: 'success' },
          2: { text: '已拒绝', color: 'error' },
        };
        const config = statusMap[status] || { text: '未知', color: 'default' };
        return <Tag color={config.color}>{config.text}</Tag>;
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (status: number) => {
        const statusMap: Record<number, { text: string; color: string }> = {
          0: { text: '禁用', color: 'error' },
          1: { text: '启用', color: 'success' },
        };
        const config = statusMap[status] || { text: '未知', color: 'default' };
        return <Tag color={config.color}>{config.text}</Tag>;
      },
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      width: 180,
      render: (time: string) => {
        return time ? new Date(time).toLocaleString('zh-CN') : '-';
      },
    },
    {
      title: '备注',
      dataIndex: 'remark',
      ellipsis: true,
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      fixed: 'right',
      render: (_: any, record: identityservicev1_Tenant) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => message.info(`编辑租户: ${record.name}`)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除"
            description={`确定要删除租户 "${record.name}" 吗？`}
            onConfirm={() => handleDelete(record.id || 0)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card>
      {/* 搜索表单 */}
      <Form form={form} layout="inline" style={{ marginBottom: 16 }}>
        <Form.Item name="name" label="租户名称">
          <Input placeholder="请输入" allowClear style={{ width: 200 }} />
        </Form.Item>
        <Form.Item name="code" label="租户编码">
          <Input placeholder="请输入" allowClear style={{ width: 200 }} />
        </Form.Item>
        <Form.Item name="tenantType" label="租户类型">
          <Select placeholder="请选择" allowClear style={{ width: 150 }}>
            <Select.Option value={0}>免费</Select.Option>
            <Select.Option value={1}>付费</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item name="auditStatus" label="审核状态">
          <Select placeholder="请选择" allowClear style={{ width: 150 }}>
            <Select.Option value={0}>待审核</Select.Option>
            <Select.Option value={1}>已通过</Select.Option>
            <Select.Option value={2}>已拒绝</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item name="status" label="状态">
          <Select placeholder="请选择" allowClear style={{ width: 150 }}>
            <Select.Option value={0}>禁用</Select.Option>
            <Select.Option value={1}>启用</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item>
          <Space>
            <Button onClick={handleReset}>重置</Button>
            <Button type="primary" onClick={handleSearch} loading={listTenantsMutation.isPending}>
              查询
            </Button>
          </Space>
        </Form.Item>
      </Form>

      {/* 数据表格 */}
      <Table
        columns={columns}
        dataSource={listTenantsMutation.data?.items || []}
        rowKey="id"
        loading={listTenantsMutation.isPending}
        pagination={{
          current: pagination.page,
          pageSize: pagination.pageSize,
          total: listTenantsMutation.data?.total || 0,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
        onChange={handleTableChange}
        scroll={{ x: 1200 }}
      />
    </Card>
  );
};

export default TenantList;
