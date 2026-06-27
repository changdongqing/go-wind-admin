/**
 * 临时占位 / Temporary placeholder.
 * 真正的实现由 Task 9 完成。
 */
import { Modal } from 'antd';

interface Props {
  open: boolean;
  onClose: (reload?: boolean) => void;
}

const CreateProductWizard = ({ open, onClose }: Props) => (
  <Modal open={open} title="新增产品" onCancel={() => onClose(false)} footer={null}>
    placeholder
  </Modal>
);

export default CreateProductWizard;
