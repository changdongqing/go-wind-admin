import { useTranslation } from 'react-i18next';
import { Tag, Tooltip } from 'antd';

import { FEATURE_TYPES, featureTypeColor, type FeatureType } from './constants';

interface FeatureTypeTreeProps {
  currentType: FeatureType;
  onTypeSelect: (type: FeatureType) => void;
}

/**
 * 特征类型树（左侧）/ Feature type tree (left panel).
 *
 * 四个固定节点：属性 / 事件 / 服务 / 关系。
 * 选中节点高亮 + 左侧蓝条（镜像单位管理的选中样式）。
 */
const FeatureTypeTree: React.FC<FeatureTypeTreeProps> = ({ currentType, onTypeSelect }) => {
  const { t } = useTranslation('feature');

  return (
    <div
      className="page-container-content"
      style={{
        padding: '12px 8px',
        display: 'flex',
        flexDirection: 'column',
        gap: 8,
        overflow: 'auto',
      }}
    >
      <div style={{ padding: '4px 12px', fontWeight: 600, color: '#666' }}>
        {t('featureType')}
      </div>

      {FEATURE_TYPES.map((type) => {
        const selected = currentType === type;
        return (
          <div
            key={type}
            onClick={() => onTypeSelect(type)}
            style={{
              cursor: 'pointer',
              padding: '10px 12px',
              borderRadius: 6,
              border: selected ? '1px solid #1677ff' : '1px solid transparent',
              background: selected ? 'rgba(22,119,255,0.06)' : 'transparent',
              borderLeft: selected ? '4px solid #1677ff' : '4px solid transparent',
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              transition: 'all 0.18s',
            }}
            onMouseEnter={(e) => {
              if (!selected) {
                (e.currentTarget as HTMLDivElement).style.background = 'rgba(0,0,0,0.03)';
              }
            }}
            onMouseLeave={(e) => {
              if (!selected) {
                (e.currentTarget as HTMLDivElement).style.background = 'transparent';
              }
            }}
          >
            <Tag color={featureTypeColor[type]} style={{ marginRight: 0 }}>
              {t(`featureTypeMap.${type}`)}
            </Tag>
            <Tooltip title={t(`featureTypeDesc.${type}`)} placement="right">
              <span style={{ color: '#999', fontSize: 12 }}>
                {t(`featureTypeDesc.${type}`)}
              </span>
            </Tooltip>
          </div>
        );
      })}
    </div>
  );
};

export default FeatureTypeTree;
