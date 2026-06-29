import { Form, Input, InputNumber, Select, Button, Space, Switch, Divider } from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

import { callModeOptions, dataTypeOptions } from '../../feature/constants';
import UnitSelect from '../../feature/components/UnitSelect';

interface ServiceSpecFormProps {
  namePath: (string | number)[];
}

/**
 * 服务 spec 表单 / Service spec form.
 *
 * 字段：callMode（必填）+ inputParams[] + outputParams[] + timeout + description。
 */
const ServiceSpecForm: React.FC<ServiceSpecFormProps> = ({ namePath }) => {
  const { t } = useTranslation('feature');
  const form = Form.useFormInstance();

  const paramListRow = (parentKey: 'inputParams' | 'outputParams', titleKey: string) => (
    <Form.Item label={t(titleKey)}>
      <Form.List name={[...namePath, parentKey]}>
        {(fields, { add, remove }) => (
          <>
            {fields.map((f) => (
              <Space.Compact key={f.key} block style={{ marginBottom: 4 }}>
                <Form.Item
                  name={[f.name, 'key']}
                  rules={[{ required: true }]}
                  style={{ flex: 1, marginBottom: 0 }}
                >
                  <Input placeholder={t('paramKey')} />
                </Form.Item>
                <Form.Item
                  name={[f.name, 'dataType']}
                  rules={[{ required: true }]}
                  style={{ flex: 1, marginBottom: 0 }}
                >
                  <Select options={dataTypeOptions(t)} placeholder={t('dataType')} />
                </Form.Item>
                <Form.Item
                  name={[f.name, 'unit', 'unitCode']}
                  style={{ flex: 1, marginBottom: 0 }}
                >
                  <UnitSelect
                    onUnitChange={(u) =>
                      form.setFieldValue(
                        [...namePath, parentKey, f.name, 'unit', 'unitSymbol'],
                        u?.symbol,
                      )
                    }
                  />
                </Form.Item>
                <Form.Item name={[f.name, 'unit', 'unitSymbol']} hidden>
                  <Input />
                </Form.Item>
                {parentKey === 'inputParams' && (
                  <Form.Item
                    name={[f.name, 'required']}
                    valuePropName="checked"
                    style={{ marginBottom: 0, paddingLeft: 8 }}
                  >
                    <Switch checkedChildren={t('required')} unCheckedChildren=" " size="small" />
                  </Form.Item>
                )}
                <Button
                  type="text"
                  danger
                  icon={<MinusCircleOutlined />}
                  onClick={() => remove(f.name)}
                />
              </Space.Compact>
            ))}
            <Button type="dashed" block icon={<PlusOutlined />} onClick={() => add()}>
              {t('addItem')}
            </Button>
          </>
        )}
      </Form.List>
    </Form.Item>
  );

  return (
    <>
      <Form.Item
        label={t('callMode')}
        name={[...namePath, 'callMode']}
        rules={[{ required: true }]}
      >
        <Select options={callModeOptions(t)} placeholder={t('callMode')} />
      </Form.Item>

      <Form.Item label={t('timeout')} name={[...namePath, 'timeout']}>
        <InputNumber min={1} placeholder={t('timeoutPlaceholder')} />
      </Form.Item>

      <Form.Item label={t('description')} name={[...namePath, 'description']}>
        <Input placeholder={t('descriptionPlaceholder')} />
      </Form.Item>

      <Divider plain style={{ margin: '4px 0' }} />
      {paramListRow('inputParams', 'inputParams')}
      <Divider plain style={{ margin: '4px 0' }} />
      {paramListRow('outputParams', 'outputParams')}
    </>
  );
};

export default ServiceSpecForm;
