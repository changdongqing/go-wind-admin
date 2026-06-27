/**
 * 物模型-分类 新建/编辑 抽屉
 * Thing-model category create/edit drawer
 *
 * 设计依据 / Design ref: docs/thingmodel/sheji/分类管理/05-前端实现设计.md §4.2
 *
 * 三种打开模式 / Three open modes:
 *   - 顶层新建 (mode=create, parent=null)  → 用户选 kind、level=1、code 为 2 位
 *   - 子分类新建 (mode=create, parent=row) → kind 继承自父、level=parent.level+1、code 为 parent.code + 2 位
 *   - 编辑 (mode=edit, data=row)           → kind/code/parent/level 全部只读，仅可改 name 等
 *
 * 不可变字段（kind/code/parent_id/level）在编辑模式下 readOnly；后端 Service 同时拦截。
 */
import { useEffect, useMemo, useRef, useState } from 'react';
import {
  DrawerForm,
  ProFormText,
  ProFormDigit,
  ProFormSelect,
  ProFormRadio,
  ProFormTextArea,
  type ProFormInstance,
} from '@ant-design/pro-components';
import { App } from 'antd';
import { useTranslation } from 'react-i18next';

import {
  useCreateCategory,
  useGetCategory,
  useUpdateCategory,
} from '@/api/hooks/category';
import type { thingmodelservicev1_Category } from '@/api/generated/admin/service/v1';

interface CategoryDrawerProps {
  open: boolean;
  mode: 'create' | 'edit';
  data?: thingmodelservicev1_Category | null;
  parent?: thingmodelservicev1_Category | null;
  onClose: (refresh?: boolean) => void;
}

const CategoryDrawer: React.FC<CategoryDrawerProps> = ({
  open,
  mode,
  data,
  parent,
  onClose,
}) => {
  const { t } = useTranslation('category');
  const formRef = useRef<ProFormInstance>(null);
  const { message } = App.useApp();

  const isEdit = mode === 'edit';

  // 本次抽屉的 level（编辑：来自 data；新建顶层：1；新建子分类：parent.level+1）
  const level = useMemo<number>(() => {
    if (isEdit) return data?.level ?? 1;
    if (parent?.level) return (parent.level ?? 0) + 1;
    return 1;
  }, [isEdit, data?.level, parent?.level]);

  const expectCodeLen = level * 2;

  // 编辑模式下若该分类不是顶层，拉取父分类用于展示 "code · name"
  // Edit mode: fetch parent category to display "code · name" instead of "#id"
  const editParentId = isEdit ? data?.parentId : undefined;
  const { data: fetchedParent } = useGetCategory(
    { id: editParentId ?? 0 },
    {
      enabled: !!editParentId,
      staleTime: 30_000,
    },
  );

  // 当前用于显示的父分类（新建子分类→props.parent；编辑→拉取到的 fetchedParent；顶层→null）
  const activeParent: thingmodelservicev1_Category | null = isEdit
    ? fetchedParent ?? null
    : parent ?? null;

  // 父分类只读展示文本
  const parentDisplay = useMemo(() => {
    if (isEdit) {
      if (!editParentId) return t('tips.topLevel'); // 顶层（编辑大类时）
      if (fetchedParent) return `${fetchedParent.code ?? ''} · ${fetchedParent.name ?? ''}`;
      return `#${editParentId}`; // fallback：父节点还未拉到
    }
    return parent ? `${parent.code ?? ''} · ${parent.name ?? ''}` : t('tips.topLevel');
  }, [isEdit, editParentId, fetchedParent, parent, t]);

  // 层级展示文本（"大类/中类/小类/细类"）
  const levelDisplay = useMemo(() => t(`level.${level}`), [level, t]);

  const [submitting, setSubmitting] = useState(false);

  const createMutation = useCreateCategory({
    onSuccess: () => {
      message.success(t('createSuccess'));
      onClose(true);
    },
    onError: (err: Error) => message.error(err.message || t('createFailed')),
  });

  const updateMutation = useUpdateCategory({
    onSuccess: () => {
      message.success(t('updateSuccess'));
      onClose(true);
    },
    onError: (err: Error) => message.error(err.message || t('updateFailed')),
  });

  // 进入抽屉后异步回填字段（destroyOnClose 时需要延迟一帧）
  useEffect(() => {
    if (!open) return;
    setTimeout(() => {
      if (isEdit && data) {
        formRef.current?.setFieldsValue({
          kind: data.kind,
          code: data.code,
          levelDisplay,
          parentDisplay,
          name: data.name,
          nameEn: data.nameEn,
          icon: data.icon,
          description: data.description,
          sortOrder: data.sortOrder ?? 0,
          isEnabled: data.isEnabled ?? true,
        });
      } else {
        // 新建：顶层（parent=null）或子分类（parent=row）
        formRef.current?.setFieldsValue({
          kind: parent?.kind, // 子分类继承父 kind；顶层为空待用户选
          code: '',            // 始终留空，由用户输入；placeholder 提示前缀
          levelDisplay,
          parentDisplay,
          name: undefined,
          nameEn: undefined,
          icon: undefined,
          description: undefined,
          sortOrder: 0,
          isEnabled: true,
        });
      }
    }, 0);
  }, [open, isEdit, data, parent, levelDisplay, parentDisplay]);

  const handleSubmit = async (values: Record<string, any>) => {
    try {
      setSubmitting(true);
      if (isEdit && data?.id) {
        // 编辑：不可变字段（kind/code/parentId/level）一律不提交
        const editable = {
          name: values.name,
          nameEn: values.nameEn,
          icon: values.icon,
          description: values.description,
          sortOrder: values.sortOrder,
          isEnabled: values.isEnabled,
        };
        await updateMutation.mutateAsync({ id: data.id, values: editable });
      } else {
        const payload: thingmodelservicev1_Category = {
          kind: values.kind,
          code: values.code,
          level,
          parentId: activeParent?.id,
          name: values.name,
          nameEn: values.nameEn,
          icon: values.icon,
          description: values.description,
          sortOrder: values.sortOrder,
          isEnabled: values.isEnabled,
        };
        await createMutation.mutateAsync({ data: payload });
      }
      return true;
    } catch {
      return false;
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <DrawerForm
      formRef={formRef}
      title={isEdit ? t('actions.edit') : t('actions.create')}
      open={open}
      onOpenChange={(visible) => {
        if (!visible) {
          formRef.current?.resetFields();
          onClose(false);
        }
      }}
      onFinish={handleSubmit}
      submitter={{
        submitButtonProps: {
          loading: submitting || createMutation.isPending || updateMutation.isPending,
        },
        resetButtonProps: { onClick: () => onClose(false) },
      }}
      drawerProps={{ destroyOnClose: true, placement: 'right', size: 600 }}
    >
      <ProFormSelect
        name="kind"
        label={t('fields.kind')}
        rules={[{ required: true, message: t('tips.requiredKind') }]}
        options={[
          { value: 'SYSTEM', label: t('kind.SYSTEM') },
          { value: 'SPACE', label: t('kind.SPACE') },
          { value: 'FACILITY', label: t('kind.FACILITY') },
        ]}
        // 子分类继承父 kind、编辑不可改 → 只读
        readonly={isEdit || !!parent}
      />

      <ProFormText
        name="parentDisplay"
        label={t('fields.parent')}
        readonly
      />

      {/* 层级用只读文本展示翻译后的名称（大类/中类/小类/细类），而不是数字 1/2/3/4 */}
      <ProFormText
        name="levelDisplay"
        label={t('fields.level')}
        readonly
      />

      <ProFormText
        name="code"
        label={t('fields.code')}
        readonly={isEdit}
        placeholder={
          isEdit
            ? undefined
            : activeParent?.code
              ? `${activeParent.code}__（${t('tips.codeLength', { need: expectCodeLen })}）`
              : `${'_'.repeat(expectCodeLen)} （${t('tips.codeLength', { need: expectCodeLen })}）`
        }
        rules={[
          { required: true, message: t('tips.requiredCode') },
          { pattern: /^\d+$/, message: t('tips.codeFormat') },
          () => ({
            validator(_, v) {
              if (isEdit) return Promise.resolve();
              if (!v) return Promise.resolve();
              if (v.length !== expectCodeLen) {
                return Promise.reject(
                  new Error(t('tips.codeLength', { need: expectCodeLen })),
                );
              }
              const parentCode = activeParent?.code ?? '';
              if (parentCode) {
                if (!v.startsWith(parentCode) || v.length !== parentCode.length + 2) {
                  return Promise.reject(
                    new Error(t('tips.codePrefix', { prefix: parentCode })),
                  );
                }
              }
              return Promise.resolve();
            },
          }),
        ]}
      />

      <ProFormText
        name="name"
        label={t('fields.name')}
        rules={[{ required: true, message: t('tips.requiredName') }]}
        fieldProps={{ allowClear: true }}
      />

      <ProFormText name="nameEn" label={t('fields.nameEn')} fieldProps={{ allowClear: true }} />

      <ProFormText name="icon" label={t('fields.icon')} fieldProps={{ allowClear: true }} />

      <ProFormTextArea
        name="description"
        label={t('fields.description')}
        fieldProps={{ rows: 2 }}
      />

      <ProFormDigit
        name="sortOrder"
        label={t('fields.sortOrder')}
        fieldProps={{ precision: 0, min: 0 }}
      />

      <ProFormRadio.Group
        name="isEnabled"
        label={t('fields.isEnabled')}
        options={[
          { value: true, label: t('tags.enabled') },
          { value: false, label: t('tags.disabled') },
        ]}
        fieldProps={{ optionType: 'button', buttonStyle: 'solid' }}
      />
    </DrawerForm>
  );
};

export default CategoryDrawer;
