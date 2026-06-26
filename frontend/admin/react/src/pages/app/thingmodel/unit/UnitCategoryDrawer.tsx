import { useRef, useState, useEffect } from 'react';
import type { ProFormInstance } from '@ant-design/pro-components';
import {
  DrawerForm,
  ProFormText,
  ProFormDigit,
  ProFormRadio,
  ProFormTextArea,
} from '@ant-design/pro-components';
import { App } from 'antd';
import { useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';

import { useCreateUnitCategory, useUpdateUnitCategory } from '@/api/hooks/unit';
import { enableBoolRadioOptions } from './constants';

interface UnitCategoryDrawerProps {
  open: boolean;
  mode: 'create' | 'edit';
  data?: any;
  onClose: () => void;
  onSuccess: () => void;
}

/**
 * 物理量分类 新建/编辑 抽屉 / UnitCategory create-edit drawer
 *
 * `code` 字段在后端 Ent schema 为 Immutable，编辑模式下置为只读。
 * The `code` field is Immutable in Ent schema; readOnly in edit mode.
 */
const UnitCategoryDrawer: React.FC<UnitCategoryDrawerProps> = ({
  open,
  mode,
  data,
  onClose,
  onSuccess,
}) => {
  const { t } = useTranslation('unit-category');
  const formRef = useRef<ProFormInstance>(null);
  const queryClient = useQueryClient();
  const { message } = App.useApp();

  const [confirmLoading, setConfirmLoading] = useState(false);

  // 编辑模式下设置表单值（destroyOnClose 时需延迟赋值）
  useEffect(() => {
    if (open && mode === 'edit' && data) {
      setTimeout(() => {
        formRef.current?.setFieldsValue({
          code: data.code || '',
          name: data.name || '',
          nameEn: data.nameEn || '',
          quantity: data.quantity || '',
          baseUnitSymbol: data.baseUnitSymbol || '',
          icon: data.icon || '',
          description: data.description || '',
          sortOrder: data.sortOrder ?? 0,
          isEnabled: data.isEnabled ?? true,
        });
      }, 0);
    }
  }, [open, mode, data]);

  const createMutation = useCreateUnitCategory({
    onSuccess: () => {
      message.success(t('createSuccess'));
      queryClient.invalidateQueries({ queryKey: ['listUnitCategories'] });
      onSuccess();
      onClose();
    },
    onError: (error: Error) => message.error(error.message || t('createFailed')),
  });

  const updateMutation = useUpdateUnitCategory({
    onSuccess: () => {
      message.success(t('updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['listUnitCategories'] });
      onSuccess();
      onClose();
    },
    onError: (error: Error) => message.error(error.message || t('updateFailed')),
  });

  const handleSubmit = async (values: Record<string, any>) => {
    try {
      setConfirmLoading(true);
      if (mode === 'edit' && data?.id) {
        // 编辑时不更新 code（Immutable）
        const { code: _ignore, ...rest } = values;
        await updateMutation.mutateAsync({ id: data.id, values: rest });
      } else {
        await createMutation.mutateAsync({ data: values });
      }
      return true;
    } catch {
      return false;
    } finally {
      setConfirmLoading(false);
    }
  };

  return (
    <DrawerForm
      formRef={formRef}
      title={mode === 'create' ? t('create') : t('edit')}
      open={open}
      onOpenChange={(visible) => {
        if (!visible) {
          formRef.current?.resetFields();
          onClose();
        }
      }}
      initialValues={{
        sortOrder: 0,
        isEnabled: true,
      }}
      onFinish={handleSubmit}
      submitter={{
        searchConfig: {
          submitText: t('common:button.submit'),
          resetText: t('common:button.cancel'),
        },
        submitButtonProps: {
          loading:
            confirmLoading || createMutation.isPending || updateMutation.isPending,
        },
        resetButtonProps: { onClick: onClose },
      }}
      drawerProps={{ destroyOnClose: true, onClose, placement: 'left', size: 600 }}
    >
      <ProFormText
        name="code"
        label={t('code')}
        placeholder={t('codePlaceholder')}
        rules={[{ required: true, message: t('requiredCode') }]}
        readonly={mode === 'edit'}
        fieldProps={{ allowClear: true }}
      />

      <ProFormText
        name="name"
        label={t('name')}
        placeholder={t('namePlaceholder')}
        rules={[{ required: true, message: t('requiredName') }]}
        fieldProps={{ allowClear: true }}
      />

      <ProFormText
        name="nameEn"
        label={t('nameEn')}
        placeholder={t('nameEnPlaceholder')}
        fieldProps={{ allowClear: true }}
      />

      <ProFormText
        name="quantity"
        label={t('quantity')}
        placeholder={t('quantityPlaceholder')}
        fieldProps={{ allowClear: true }}
      />

      <ProFormText
        name="baseUnitSymbol"
        label={t('baseUnitSymbol')}
        placeholder={t('baseUnitSymbolPlaceholder')}
        fieldProps={{ allowClear: true }}
      />

      <ProFormText
        name="icon"
        label={t('icon')}
        placeholder={t('iconPlaceholder')}
        fieldProps={{ allowClear: true }}
      />

      <ProFormTextArea
        name="description"
        label={t('description')}
        placeholder={t('descriptionPlaceholder')}
        fieldProps={{ rows: 2 }}
      />

      <ProFormDigit
        name="sortOrder"
        label={t('sortOrder')}
        placeholder={t('sortOrderPlaceholder')}
        fieldProps={{ precision: 0, min: 0 }}
      />

      <ProFormRadio.Group
        name="isEnabled"
        label={t('status')}
        rules={[{ required: true, message: t('requiredStatus') }]}
        options={enableBoolRadioOptions(t)}
        fieldProps={{ optionType: 'button', buttonStyle: 'solid' }}
      />
    </DrawerForm>
  );
};

export default UnitCategoryDrawer;
