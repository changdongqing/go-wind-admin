import { Splitter } from 'antd';
import { useState } from 'react';

import ContentContainer from '@/layouts/components/PageContainer/ContentContainer';
import UnitCategoryList from './UnitCategoryList';
import UnitList from './UnitList';

import './index.less';

/**
 * 单位管理页面 / Unit management page.
 *
 * Splitter 左右分栏：左侧为物理量分类列表，右侧为该分类下的单位列表。
 * 镜像 dict 管理页（`pages/app/system/dict/index.tsx`）。
 */
const UnitManagement = () => {
  const [currentCategoryId, setCurrentCategoryId] = useState<number | null>(null);
  const [currentCategoryName, setCurrentCategoryName] = useState<string | undefined>();

  const handleCategorySelect = (id: number, name?: string) => {
    setCurrentCategoryId(id);
    setCurrentCategoryName(name);
  };

  return (
    <ContentContainer heightMode="fixed" padding="16px" bottomMargin={0}>
      <Splitter style={{ height: '100%', flex: 1, minHeight: 0 }}>
        <Splitter.Panel
          collapsible
          defaultSize="38%"
          min="25%"
          max="50%"
          style={{ display: 'flex', flexDirection: 'column' }}
        >
          <UnitCategoryList
            currentCategoryId={currentCategoryId}
            onCategorySelect={handleCategorySelect}
          />
        </Splitter.Panel>
        <Splitter.Panel style={{ display: 'flex', flexDirection: 'column' }}>
          <UnitList categoryId={currentCategoryId} categoryName={currentCategoryName} />
        </Splitter.Panel>
      </Splitter>
    </ContentContainer>
  );
};

export default UnitManagement;
