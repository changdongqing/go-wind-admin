import { Form, InputNumber, Input, Select, Button, Space, Switch, Divider } from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

import {
  accessModeOptions,
  dataTypeOptions,
  propertyCategoryOptions,
  type DataType,
} from '../constants';

interface PropertySpecFormProps {
  /** Form 字段名前缀（spec.property）/ Form field name prefix */
  namePath: (string | number)[];
}

/**
 * 属性 spec 表单 / Property spec form.
 *
 * 设计依据 / Design ref: docs/thingmodel/sheji/10-特征参数与spec设计.md §1
 *
 * 按 dataType 动态显示约束字段：
 *   - INT/FLOAT/DOUBLE → constraints.{min,max,step,defaultValue}
 *   - BOOL → boolLabels.{false,true}
 *   - ENUM → enumItems[] (Form.List)
 *   - TEXT → textMaxLength
 *   - STRUCT/ARRAY → 嵌套结构（本期前端仅提供 JSON 文本框，深嵌套留给后续优化）
 */
const PropertySpecForm: React.FC<PropertySpecFormProps> = ({ namePath }) => {
  const { t } = useTranslation('feature');

  return (
    <>
      <Form.Item
        label={t('dataType')}
        name={[...namePath, 'dataType']}
        rules={[{ required: true, message: t('dataTypePlaceholder') }]}
      >
        <Select options={dataTypeOptions(t)} placeholder={t('dataTypePlaceholder')} />
      </Form.Item>

      <Form.Item
        label={t('accessMode')}
        name={[...namePath, 'accessMode']}
        rules={[{ required: true }]}
      >
        <Select options={accessModeOptions(t)} placeholder={t('accessMode')} />
      </Form.Item>

      <Form.Item label={t('propertyCategory')} name={[...namePath, 'category']}>
        <Select
          options={propertyCategoryOptions(t)}
          placeholder={t('propertyCategoryPlaceholder')}
          allowClear
        />
      </Form.Item>

      <Form.Item label={t('isRated')} name={[...namePath, 'isRated']} valuePropName="checked">
        <Switch />
      </Form.Item>

      {/* unit：单位引用（unitCode + unitSymbol；unitId 由后端按 code 解析或前端用单位选择器赋值） */}
      <Divider plain style={{ margin: '4px 0' }}>
        {t('unit')}
      </Divider>
      <Space.Compact block>
        <Form.Item
          label={t('unit')}
          name={[...namePath, 'unit', 'unitCode']}
          style={{ flex: 1 }}
          tooltip="例 celsius；后端按 code 自动解析 unitId"
        >
          <Input placeholder="unitCode" />
        </Form.Item>
        <Form.Item
          label={t('unit') + ' (symbol)'}
          name={[...namePath, 'unit', 'unitSymbol']}
          style={{ flex: 1 }}
        >
          <Input placeholder="℃ / kW / s ..." />
        </Form.Item>
      </Space.Compact>

      <Divider plain style={{ margin: '4px 0' }}>
        {t('constraints')}
      </Divider>

      {/* 数值约束 */}
      <Form.Item
        noStyle
        shouldUpdate={(prev, cur) =>
          prev?.spec?.property?.dataType !== cur?.spec?.property?.dataType
        }
      >
        {({ getFieldValue }) => {
          const dt: DataType | undefined = getFieldValue([...namePath, 'dataType']);
          return (
            <>
              {(dt === 'INT' || dt === 'FLOAT' || dt === 'DOUBLE') && (
                <Space.Compact block>
                  <Form.Item
                    label={t('min')}
                    name={[...namePath, 'constraints', 'min']}
                    style={{ flex: 1 }}
                  >
                    <InputNumber style={{ width: '100%' }} placeholder={t('min')} />
                  </Form.Item>
                  <Form.Item
                    label={t('max')}
                    name={[...namePath, 'constraints', 'max']}
                    style={{ flex: 1 }}
                  >
                    <InputNumber style={{ width: '100%' }} placeholder={t('max')} />
                  </Form.Item>
                  <Form.Item
                    label={t('step')}
                    name={[...namePath, 'constraints', 'step']}
                    style={{ flex: 1 }}
                  >
                    <InputNumber style={{ width: '100%' }} placeholder={t('step')} />
                  </Form.Item>
                  <Form.Item
                    label={t('defaultValue')}
                    name={[...namePath, 'constraints', 'defaultValue']}
                    style={{ flex: 1 }}
                  >
                    <Input placeholder={t('defaultValue')} />
                  </Form.Item>
                </Space.Compact>
              )}

              {dt === 'BOOL' && (
                <Space.Compact block>
                  <Form.Item
                    label={t('boolFalseLabel')}
                    name={[...namePath, 'boolLabels', 'false']}
                    style={{ flex: 1 }}
                  >
                    <Input placeholder="关" />
                  </Form.Item>
                  <Form.Item
                    label={t('boolTrueLabel')}
                    name={[...namePath, 'boolLabels', 'true']}
                    style={{ flex: 1 }}
                  >
                    <Input placeholder="开" />
                  </Form.Item>
                </Space.Compact>
              )}

              {dt === 'TEXT' && (
                <Form.Item label={t('textMaxLength')} name={[...namePath, 'textMaxLength']}>
                  <InputNumber min={1} style={{ width: '100%' }} />
                </Form.Item>
              )}

              {dt === 'ENUM' && (
                <Form.Item label={t('enumItems')} required>
                  <Form.List name={[...namePath, 'enumItems']}>
                    {(fields, { add, remove }) => (
                      <>
                        {fields.map((f) => (
                          <Space.Compact key={f.key} block style={{ marginBottom: 4 }}>
                            <Form.Item
                              name={[f.name, 'value']}
                              rules={[{ required: true }]}
                              style={{ flex: 1, marginBottom: 0 }}
                            >
                              <InputNumber style={{ width: '100%' }} placeholder={t('enumValue')} />
                            </Form.Item>
                            <Form.Item
                              name={[f.name, 'label']}
                              rules={[{ required: true }]}
                              style={{ flex: 2, marginBottom: 0 }}
                            >
                              <Input placeholder={t('enumLabel')} />
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
              )}

              {dt === 'STRUCT' && (
                <Form.Item label={t('structFields')} required>
                  <Form.List name={[...namePath, 'structFields']}>
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
              )}

              {dt === 'ARRAY' && (
                <Space.Compact block>
                  <Form.Item
                    label={t('arraySize')}
                    name={[...namePath, 'arraySpec', 'size']}
                    style={{ flex: 1 }}
                  >
                    <InputNumber min={1} style={{ width: '100%' }} />
                  </Form.Item>
                  <Form.Item
                    label={t('elementType')}
                    name={[...namePath, 'arraySpec', 'element', 'dataType']}
                    style={{ flex: 1 }}
                    rules={[{ required: true }]}
                  >
                    <Select options={dataTypeOptions(t)} placeholder={t('dataType')} />
                  </Form.Item>
                </Space.Compact>
              )}
            </>
          );
        }}
      </Form.Item>
    </>
  );
};

export default PropertySpecForm;
