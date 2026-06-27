import { Splitter } from 'antd';
import { useState } from 'react';

import ContentContainer from '@/layouts/components/PageContainer/ContentContainer';
import FeatureTypeTree from './FeatureTypeTree';
import FeatureList from './FeatureList';
import type { FeatureType } from './constants';

/**
 * 特征管理页面 / Feature management page.
 *
 * Splitter 左右分栏：左侧为特征类型树（4 个固定节点），右侧为该类型下的特征列表与表单。
 * 镜像 unit/index.tsx 范式。
 * Mirrors unit/index.tsx.
 */
const FeatureManagement = () => {
  const [currentType, setCurrentType] = useState<FeatureType>('PROPERTY');

  return (
    <ContentContainer heightMode="fixed" padding="16px" bottomMargin={0}>
      <Splitter style={{ height: '100%', flex: 1, minHeight: 0 }}>
        <Splitter.Panel
          collapsible
          defaultSize="22%"
          min="18%"
          max="35%"
          style={{ display: 'flex', flexDirection: 'column' }}
        >
          <FeatureTypeTree currentType={currentType} onTypeSelect={setCurrentType} />
        </Splitter.Panel>
        <Splitter.Panel style={{ display: 'flex', flexDirection: 'column' }}>
          <FeatureList featureType={currentType} />
        </Splitter.Panel>
      </Splitter>
    </ContentContainer>
  );
};

export default FeatureManagement;
