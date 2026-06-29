import { Form, Input, InputNumber, Select, Switch, Divider, Space } from 'antd';
import { useTranslation } from 'react-i18next';

import {
  cardinalityOptions,
  entityKindOptions,
  relationTypeOptions,
} from '../../feature/constants';

interface RelationSpecFormProps {
  namePath: (string | number)[];
}

/**
 * 关系 spec 表单 / Relation spec form.
 *
 * 字段：relationType + source + target + cardinality + directional + properties(map)
 * source/target.kind 支持 feature（指向同表）/ external（外部实体，本期弱校验）。
 */
const RelationSpecForm: React.FC<RelationSpecFormProps> = ({ namePath }) => {
  const { t } = useTranslation('feature');

  const renderEntityRef = (childPath: 'source' | 'target', title: string) => (
    <>
      <Divider plain style={{ margin: '6px 0' }}>
        {title}
      </Divider>

      <Form.Item
        label={t('entityKind')}
        name={[...namePath, childPath, 'kind']}
        rules={[{ required: true }]}
      >
        <Select options={entityKindOptions(t)} placeholder={t('entityKind')} />
      </Form.Item>

      <Form.Item
        noStyle
        shouldUpdate={(prev, cur) => {
          const get = (o: any) => o?.spec?.relation?.[childPath]?.kind;
          return get(prev) !== get(cur);
        }}
      >
        {({ getFieldValue }) => {
          const kind = getFieldValue([...namePath, childPath, 'kind']);
          if (kind === 'feature') {
            return (
              <Space.Compact block>
                <Form.Item
                  label="id"
                  name={[...namePath, childPath, 'id']}
                  style={{ flex: 1 }}
                  tooltip="同表 feature.id；与 code/identifier 至少给一个"
                >
                  <InputNumber min={1} style={{ width: '100%' }} placeholder="feature id" />
                </Form.Item>
                <Form.Item
                  label="code"
                  name={[...namePath, childPath, 'code']}
                  style={{ flex: 1 }}
                >
                  <Input placeholder="P-RUN-0001" />
                </Form.Item>
                <Form.Item
                  label="identifier"
                  name={[...namePath, childPath, 'identifier']}
                  style={{ flex: 1 }}
                >
                  <Input placeholder="powerSwitch" />
                </Form.Item>
              </Space.Compact>
            );
          }
          if (kind === 'external') {
            return (
              <Space.Compact block>
                <Form.Item
                  label={t('entityKind') + ' type'}
                  name={[...namePath, childPath, 'type']}
                  style={{ flex: 1 }}
                >
                  <Input placeholder={t('externalTypePlaceholder')} />
                </Form.Item>
                <Form.Item
                  label="code"
                  name={[...namePath, childPath, 'code']}
                  style={{ flex: 1 }}
                >
                  <Input placeholder={t('externalCodePlaceholder')} />
                </Form.Item>
              </Space.Compact>
            );
          }
          return null;
        }}
      </Form.Item>
    </>
  );

  return (
    <>
      <Form.Item
        label={t('relationType')}
        name={[...namePath, 'relationType']}
        rules={[{ required: true }]}
      >
        <Select
          options={relationTypeOptions(t)}
          placeholder={t('relationType')}
          showSearch
          optionFilterProp="label"
        />
      </Form.Item>

      <Form.Item label={t('cardinality')} name={[...namePath, 'cardinality']}>
        <Select options={cardinalityOptions(t)} placeholder={t('cardinality')} allowClear />
      </Form.Item>

      <Form.Item
        label={t('directional')}
        name={[...namePath, 'directional']}
        valuePropName="checked"
      >
        <Switch />
      </Form.Item>

      {renderEntityRef('source', t('source'))}
      {renderEntityRef('target', t('target'))}
    </>
  );
};

export default RelationSpecForm;
