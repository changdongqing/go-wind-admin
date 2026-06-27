import { Form, Input, InputNumber, Select, Button, Space, Switch } from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

import { dataTypeOptions, eventLevelOptions } from '../constants';

interface EventSpecFormProps {
  namePath: (string | number)[];
}

/**
 * 事件 spec 表单 / Event spec form.
 *
 * 字段：level（必填）+ outputParams[] + triggerCondition + severity。
 */
const EventSpecForm: React.FC<EventSpecFormProps> = ({ namePath }) => {
  const { t } = useTranslation('feature');

  return (
    <>
      <Form.Item
        label={t('eventLevel')}
        name={[...namePath, 'level']}
        rules={[{ required: true }]}
      >
        <Select options={eventLevelOptions(t)} placeholder={t('eventLevel')} />
      </Form.Item>

      <Form.Item label={t('triggerCondition')} name={[...namePath, 'triggerCondition']}>
        <Input placeholder={t('triggerConditionPlaceholder')} />
      </Form.Item>

      <Form.Item label={t('severity')} name={[...namePath, 'severity']}>
        <InputNumber min={1} max={5} placeholder={t('severityPlaceholder')} />
      </Form.Item>

      <Form.Item label={t('outputParams')}>
        <Form.List name={[...namePath, 'outputParams']}>
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
                    name={[f.name, 'unit', 'unitSymbol']}
                    style={{ flex: 1, marginBottom: 0 }}
                  >
                    <Input placeholder="unitSymbol" />
                  </Form.Item>
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
    </>
  );
};

export default EventSpecForm;
