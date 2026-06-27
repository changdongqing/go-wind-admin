import { Form, InputNumber, Input, Select, Button, Space, Switch, Divider } from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

import {
  accessModeOptions,
  dataTypeOptions,
  propertyCategoryOptions,
  type DataType,
} from '../constants';
import UnitSelect from '../components/UnitSelect';

/** 仅这些数据类型在物模型语义上需要带物理单位 */
const UNIT_BEARING_TYPES: ReadonlySet<DataType> = new Set(['INT', 'FLOAT', 'DOUBLE']);

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
  const form = Form.useFormInstance();

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

      {/* unit：单位引用（unitCode 由下拉选择，unitSymbol 自动回填；unitId 由后端按 code 解析） */}
      <Form.Item
        noStyle
        shouldUpdate={(prev, cur) =>
          prev?.spec?.property?.dataType !== cur?.spec?.property?.dataType
        }
      >
        {({ getFieldValue }) => {
          const dt: DataType | undefined = getFieldValue([...namePath, 'dataType']);
          if (!dt || !UNIT_BEARING_TYPES.has(dt)) return null;
          return (
            <>
              <Divider plain style={{ margin: '4px 0' }}>
                {t('unit')}
              </Divider>
              <Form.Item
                label={t('unit')}
                name={[...namePath, 'unit', 'unitCode']}
                tooltip="按编码选择已登记的单位，名称/符号自动同步"
              >
                <UnitSelect
                  onUnitChange={(u) =>
                    form.setFieldValue([...namePath, 'unit', 'unitSymbol'], u?.symbol)
                  }
                />
              </Form.Item>
              {/* 隐藏字段：保存 symbol，提交时随 spec 一起回传后端 */}
              <Form.Item name={[...namePath, 'unit', 'unitSymbol']} hidden>
                <Input />
              </Form.Item>
            </>
          );
        }}
      </Form.Item>

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
                              name={[f.name, 'unit', 'unitCode']}
                              style={{ flex: 1, marginBottom: 0 }}
                            >
                              <UnitSelect
                                onUnitChange={(u) =>
                                  form.setFieldValue(
                                    [...namePath, 'structFields', f.name, 'unit', 'unitSymbol'],
                                    u?.symbol,
                                  )
                                }
                              />
                            </Form.Item>
                            <Form.Item name={[f.name, 'unit', 'unitSymbol']} hidden>
                              <Input />
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
