/**
 * 临时占位 / Temporary placeholder.
 * 真正的实现由 Task 11 完成。
 */
import { Drawer } from 'antd';

interface Props {
  open: boolean;
  productId: number;
  feature: any;
  mode: 'edit' | 'create-local' | 'create-global';
  readonly?: 'partial' | false;
  onClose: (reload?: boolean) => void;
}

const ProductFeatureDrawer = ({ open, onClose }: Props) => (
  <Drawer open={open} onClose={() => onClose(false)} title="编辑特征 (placeholder)">
    placeholder
  </Drawer>
);

export default ProductFeatureDrawer;
