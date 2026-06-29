import { useRef, useState, useEffect } from 'react';
import type { ProFormInstance } from '@ant-design/pro-components';
import {
  DrawerForm,
  ProFormText,
  ProFormDigit,
  ProFormRadio,
  ProFormSelect,
  ProFormTextArea,
} from '@ant-design/pro-components';
import { Alert, App, Divider } from 'antd';
import { useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';

import { useCreateFeature, useUpdateFeature } from '@/api/hooks/feature';
import {
  enableBoolRadioOptions,
  featureTypeOptions,
  type FeatureType,
} from './constants';

interface FeatureDrawerProps {
  open: boolean;
  mode: 'create' | 'edit';
  /** 当前选中的特征类型；新建时作为默认 / 编辑时从 data 读取 */
  featureType: FeatureType;
  data?: any;
  onClose: () => void;
  onSuccess: () => void;
}

/**
 * 特征 新建/编辑 抽屉（CR-001 后：仅骨架字段）/ Feature create-edit drawer.
 *
 * CR-001（2026-06-29）：结构化约束 spec 整体下沉到模型层，本抽屉只承载特征骨架：
 *   - 公共字段：code / identifier / name / featureType / applicableScope / description
 *   - CR-001 新增：semanticTag / recommendedUnitCategoryId（推荐元信息，不参与约束计算）
 *   - 治理：sortOrder / isEnabled
 *
 * 详见 docs/thingmodel/sheji/修改记录/CR-001-结构化约束下沉到模型层.md。
 */
const FeatureDrawer: React.FC<FeatureDrawerProps> = ({
  open,
  mode,
  featureType,
  data,
  onClose,
  onSuccess,
}) => {
  const { t } = useTranslation('feature');
  const formRef = useRef<ProFormInstance>(null);
  const queryClient = useQueryClient();
  const { message } = App.useApp();

  const [confirmLoading, setConfirmLoading] = useState(false);

  // 编辑模式回填
  useEffect(() => {
    if (!open) return;
    const ft = (data?.featureType as FeatureType) || featureType;
    setTimeout(() => {
      if (mode === 'edit' && data) {
        formRef.current?.setFieldsValue({
          featureType: ft,
          code: data.code,
          identifier: data.identifier,
          name: data.name,
          nameEn: data.nameEn,
          description: data.description,
          applicableScope: data.applicableScope,
          semanticTag: data.semanticTag,
          recommendedUnitCategoryId: data.recommendedUnitCategoryId,
          sortOrder: data.sortOrder ?? 0,
          isEnabled: data.isEnabled ?? true,
        });
      } else {
        // 新建时初始化
        formRef.current?.resetFields();
        formRef.current?.setFieldsValue({
          featureType: ft,
          isEnabled: true,
          sortOrder: 0,
        });
      }
    }, 0);
  }, [open, mode, data, featureType]);

  const createMutation = useCreateFeature({
    onSuccess: () => {
      message.success(t('createSuccess'));
      queryClient.invalidateQueries({ queryKey: ['listFeatures'] });
      queryClient.invalidateQueries({ queryKey: ['listFeaturesByType'] });
      onSuccess();
      onClose();
    },
    onError: (error: Error) => message.error(error.message || t('createFailed')),
  });

  const updateMutation = useUpdateFeature({
    onSuccess: () => {
      message.success(t('updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['listFeatures'] });
      queryClient.invalidateQueries({ queryKey: ['listFeaturesByType'] });
      onSuccess();
      onClose();
    },
    onError: (error: Error) => message.error(error.message || t('updateFailed')),
  });

  const handleSubmit = async (values: Record<string, any>) => {
    try {
      setConfirmLoading(true);
      const ft = values.featureType as FeatureType;

      if (mode === 'create') {
        await createMutation.mutateAsync({
          data: {
            featureType: ft as any,
            code: values.code,
            identifier: values.identifier,
            name: values.name,
            nameEn: values.nameEn,
            description: values.description,
            applicableScope: values.applicableScope,
            semanticTag: values.semanticTag,
            recommendedUnitCategoryId: values.recommendedUnitCategoryId,
            sortOrder: values.sortOrder,
            isEnabled: values.isEnabled,
          },
        });
      } else {
        await updateMutation.mutateAsync({
          id: data?.id,
          values: {
            featureType: ft,
            identifier: values.identifier,
            name: values.name,
            nameEn: values.nameEn,
            description: values.description,
            applicableScope: values.applicableScope,
            semanticTag: values.semanticTag,
            recommendedUnitCategoryId: values.recommendedUnitCategoryId,
            sortOrder: values.sortOrder,
            isEnabled: values.isEnabled,
          },
        });
      }
      return true;
    } finally {
      setConfirmLoading(false);
    }
  };

  return (
    <DrawerForm
      title={mode === 'create' ? t('create') : t('edit')}
      width={680}
      open={open}
      formRef={formRef}
      onOpenChange={(o) => {
        if (!o) onClose();
      }}
      onFinish={handleSubmit}
      drawerProps={{ destroyOnHidden: true, maskClosable: false }}
      submitter={{
        searchConfig: { submitText: t('common:button.ok'), resetText: t('common:button.cancel') },
        submitButtonProps: { loading: confirmLoading },
      }}
      autoFocusFirstInput
    >
      <Alert
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
        message={t('skeletonOnlyNotice', {
          defaultValue:
            'CR-001 起，特征仅承载身份骨架；结构化约束（数据类型/范围/枚举项等）在分类默认模型与产品模型中按场景配置。',
        })}
      />

      <Divider orientation="left" plain>
        {t('commonSection')}
      </Divider>

      <ProFormSelect
        name="featureType"
        label={t('featureType')}
        options={featureTypeOptions(t)}
        rules={[{ required: true }]}
        disabled={mode === 'edit'}
      />

      <ProFormText
        name="code"
        label={t('code')}
        placeholder={t('codePlaceholder')}
        rules={[{ required: true, message: t('requiredCode') }]}
        readonly={mode === 'edit'}
      />
      <ProFormText
        name="identifier"
        label={t('identifier')}
        placeholder={t('identifierPlaceholder')}
        rules={[{ required: true, message: t('requiredIdentifier') }]}
      />
      <ProFormText
        name="name"
        label={t('name')}
        placeholder={t('namePlaceholder')}
        rules={[{ required: true, message: t('requiredName') }]}
      />
      <ProFormText
        name="nameEn"
        label={t('nameEn')}
        placeholder={t('nameEnPlaceholder')}
      />
      <ProFormTextArea
        name="description"
        label={t('description')}
        placeholder={t('descriptionPlaceholder')}
        fieldProps={{ rows: 2 }}
      />
      <ProFormText
        name="applicableScope"
        label={t('applicableScope')}
        placeholder={t('applicableScopePlaceholder')}
      />

      <Divider orientation="left" plain>
        {t('recommendationSection', { defaultValue: '推荐元信息（仅用于检索与提示）' })}
      </Divider>

      <ProFormText
        name="semanticTag"
        label={t('semanticTag', { defaultValue: '语义标签' })}
        placeholder={t('semanticTagPlaceholder', {
          defaultValue: '如 pressure / temperature / runMode（软性检索字段）',
        })}
      />
      <ProFormDigit
        name="recommendedUnitCategoryId"
        label={t('recommendedUnitCategoryId', { defaultValue: '推荐单位分类 ID' })}
        placeholder={t('recommendedUnitCategoryIdPlaceholder', {
          defaultValue: '指向 thingmodel_unit_categories.id，模型层 spec 编辑时按此预过滤单位',
        })}
        min={0}
        fieldProps={{ precision: 0 }}
      />

      <Divider orientation="left" plain>
        {t('governanceSection')}
      </Divider>

      <ProFormDigit
        name="sortOrder"
        label={t('sortOrder')}
        placeholder={t('sortOrderPlaceholder')}
        min={0}
        fieldProps={{ precision: 0 }}
      />
      <ProFormRadio.Group
        name="isEnabled"
        label={t('status')}
        options={enableBoolRadioOptions(t)}
      />
    </DrawerForm>
  );
};

export default FeatureDrawer;
