import { useRef, useState, useEffect } from 'react';
import type { ProFormInstance } from '@ant-design/pro-components';
import {
  DrawerForm,
  ProFormDependency,
  ProFormText,
  ProFormDigit,
  ProFormRadio,
  ProFormSelect,
  ProFormSwitch,
  ProFormTextArea,
} from '@ant-design/pro-components';
import { App, Form } from 'antd';
import { useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';

import { useCreateUnit, useUpdateUnit } from '@/api/hooks/unit';
import {
  conversionTypeOptions,
  enableBoolRadioOptions,
} from './constants';

interface UnitDrawerProps {
  open: boolean;
  mode: 'create' | 'edit';
  data?: any;
  /** 当前选中的物理量分类 id（新建时必传）。 */
  categoryId: number;
  /** 当前分类名（仅用于展示，避免再查一次）。 */
  categoryName?: string;
  onClose: () => void;
  onSuccess: () => void;
}

/**
 * 单位 新建/编辑 抽屉 / Unit create-edit drawer.
 *
 * 前端即时校验与后端 validateUnit 保持一致（设计文档 03 §4）：
 *   - isBase=true   → factor=1, offset=0，且 factor/offset 输入框禁用
 *   - LINEAR        → offset 强制 0（输入禁用）
 *   - AFFINE        → factor、offset 均可编辑
 *   - LOGARITHMIC / CONDITIONAL / NONE → 系数无意义，但允许填写以保持透明（后端不参与计算）
 *
 * `code` 字段 Immutable，编辑模式只读。
 */
const UnitDrawer: React.FC<UnitDrawerProps> = ({
  open,
  mode,
  data,
  categoryId,
  categoryName,
  onClose,
  onSuccess,
}) => {
  const { t } = useTranslation('unit');
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
          symbol: data.symbol || '',
          name: data.name || '',
          nameEn: data.nameEn || '',
          isBase: data.isBase ?? false,
          conversionType: data.conversionType || 'LINEAR',
          factor: data.factor ?? 1,
          offset: data.offset ?? 0,
          formulaExpr: data.formulaExpr || '',
          precision: data.precision ?? 2,
          isSiUnit: data.isSiUnit ?? false,
          isLegalUnit: data.isLegalUnit ?? false,
          sortOrder: data.sortOrder ?? 0,
          isEnabled: data.isEnabled ?? true,
        });
      }, 0);
    }
  }, [open, mode, data]);

  const createMutation = useCreateUnit({
    onSuccess: () => {
      message.success(t('createSuccess'));
      queryClient.invalidateQueries({ queryKey: ['listUnits'] });
      queryClient.invalidateQueries({ queryKey: ['listUnitsByCategory'] });
      onSuccess();
      onClose();
    },
    onError: (error: Error) => message.error(error.message || t('createFailed')),
  });

  const updateMutation = useUpdateUnit({
    onSuccess: () => {
      message.success(t('updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['listUnits'] });
      queryClient.invalidateQueries({ queryKey: ['listUnitsByCategory'] });
      onSuccess();
      onClose();
    },
    onError: (error: Error) => message.error(error.message || t('updateFailed')),
  });

  const handleSubmit = async (values: Record<string, any>) => {
    try {
      setConfirmLoading(true);
      // 即时校验：基准 / 线性 一致性
      const isBase = !!values.isBase;
      const ct = values.conversionType as string | undefined;
      const factor = Number(values.factor ?? 1);
      const offset = Number(values.offset ?? 0);
      if (isBase && (factor !== 1 || offset !== 0)) {
        message.error(t('validation.baseFactorMustBeOne'));
        return false;
      }
      if (ct === 'LINEAR' && offset !== 0) {
        message.error(t('validation.linearOffsetMustBeZero'));
        return false;
      }
      if (factor === 0) {
        message.error(t('validation.factorMustNotBeZero'));
        return false;
      }

      const payload = {
        ...values,
        categoryId,
      };

      if (mode === 'edit' && data?.id) {
        const { code: _ignore, ...rest } = payload;
        await updateMutation.mutateAsync({ id: data.id, values: rest });
      } else {
        await createMutation.mutateAsync({ data: payload });
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
        isBase: false,
        conversionType: 'LINEAR',
        factor: 1,
        offset: 0,
        precision: 2,
        isSiUnit: false,
        isLegalUnit: false,
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
      drawerProps={{ destroyOnClose: true, onClose, size: 600 }}
    >
      {/* 所属分类（只读展示） / Belongs-to category (readonly) */}
      <Form.Item label={t('categoryId')}>
        <span>
          {categoryName ? `${categoryName} (id=${categoryId})` : `id=${categoryId}`}
        </span>
      </Form.Item>

      <ProFormText
        name="code"
        label={t('code')}
        placeholder={t('codePlaceholder')}
        rules={[{ required: true, message: t('requiredCode') }]}
        readonly={mode === 'edit'}
        fieldProps={{ allowClear: true }}
      />

      <ProFormText
        name="symbol"
        label={t('symbol')}
        placeholder={t('symbolPlaceholder')}
        rules={[{ required: true, message: t('requiredSymbol') }]}
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

      <ProFormSwitch name="isBase" label={t('isBase')} />

      <ProFormSelect
        name="conversionType"
        label={t('conversionType')}
        options={conversionTypeOptions(t)}
        rules={[{ required: true }]}
        allowClear={false}
      />

      <ProFormDependency name={['isBase', 'conversionType']}>
        {({ isBase, conversionType }) => {
          const baseLock = !!isBase;
          const offsetLock = baseLock || conversionType === 'LINEAR';
          return (
            <>
              <ProFormDigit
                name="factor"
                label={t('factor')}
                placeholder={t('factorPlaceholder')}
                disabled={baseLock}
                fieldProps={{ stringMode: false, step: 0.001 }}
              />
              <ProFormDigit
                name="offset"
                label={t('offset')}
                placeholder={t('offsetPlaceholder')}
                disabled={offsetLock}
                fieldProps={{ stringMode: false, step: 0.001 }}
              />
            </>
          );
        }}
      </ProFormDependency>

      <ProFormTextArea
        name="formulaExpr"
        label={t('formulaExpr')}
        placeholder={t('formulaExprPlaceholder')}
        fieldProps={{ rows: 2 }}
      />

      <ProFormDigit
        name="precision"
        label={t('precision')}
        placeholder={t('precisionPlaceholder')}
        fieldProps={{ precision: 0, min: 0, max: 12 }}
      />

      <ProFormSwitch name="isSiUnit" label={t('isSiUnit')} />
      <ProFormSwitch name="isLegalUnit" label={t('isLegalUnit')} />

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

export default UnitDrawer;
