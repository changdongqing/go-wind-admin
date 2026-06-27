/**
 * 新增产品两步向导 / Two-step product creation wizard.
 *
 * Step 1: 基本信息（含 level=4 分类级联选择，限制 kind=FACILITY）
 * Step 2: 拉取默认模型（默认全选，可反选 + SKIP/REPLACE 冲突策略）
 *
 * 提交流程：CreateProduct → 拿到 id → PullFromDefault。
 *
 * 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/05-前端实现设计.md §1.4
 */
import { useEffect, useState } from 'react';
import {
  Modal,
  Steps,
  Form,
  Input,
  Checkbox,
  Radio,
  App,
  Spin,
  Cascader,
  Button as AntButton,
} from 'antd';
import { useTranslation } from 'react-i18next';
import { useCreateProduct, fetchGetProduct } from '@/api/hooks/product';
import { usePullFromDefault } from '@/api/hooks/product-feature';
import { fetchListCategoryDefaultFeatures } from '@/api/hooks/category-default-feature';
import { fetchListCategories } from '@/api/hooks/category';
import { PaginationQuery } from '@/core/transport/rest';
import type { thingmodelservicev1_CategoryDefaultFeature } from '@/api/generated/admin/service/v1';

interface Props {
  open: boolean;
  onClose: (reload?: boolean) => void;
}

interface CascadeNode {
  value: number;
  label: string;
  children?: CascadeNode[];
}

/**
 * 按 code 前缀构建 FACILITY 分类树（level 1→2→3→4）。
 * Build FACILITY category tree by parent_id.
 */
function buildCascade(
  items: Array<{
    id?: number;
    code?: string;
    level?: number;
    name?: string;
    parentId?: number;
  }>,
): CascadeNode[] {
  const byId = new Map<number, CascadeNode>();
  items.forEach((it) => {
    if (it.id == null) return;
    byId.set(it.id, { value: it.id, label: `${it.code ?? ''} ${it.name ?? ''}` });
  });
  const roots: CascadeNode[] = [];
  items.forEach((it) => {
    if (it.id == null) return;
    const n = byId.get(it.id)!;
    const parent = it.parentId ? byId.get(it.parentId) : undefined;
    if (parent) {
      (parent.children ??= []).push(n);
    } else {
      roots.push(n);
    }
  });
  return roots;
}

const CreateProductWizard = ({ open, onClose }: Props) => {
  const { t } = useTranslation(['product', 'common']);
  const { message } = App.useApp();
  const [step, setStep] = useState(0);
  const [form] = Form.useForm();
  const [productId, setProductId] = useState<number | null>(null);
  const [defaults, setDefaults] = useState<thingmodelservicev1_CategoryDefaultFeature[]>([]);
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [onConflict, setOnConflict] = useState<'SKIP' | 'REPLACE'>('SKIP');
  const [loadingDefaults, setLoadingDefaults] = useState(false);
  const [categoryOptions, setCategoryOptions] = useState<CascadeNode[]>([]);

  const { mutate: doCreate, isPending: creating } = useCreateProduct();
  const { mutate: doPull, isPending: pulling } = usePullFromDefault();

  // 进入向导时拉一次完整 FACILITY 分类树
  useEffect(() => {
    if (!open) return;
    (async () => {
      const q = new PaginationQuery({
        paging: { page: 1, pageSize: 1000 },
        formValues: { kind: 'FACILITY' },
        orderBy: ['code'],
      });
      try {
        const resp = await fetchListCategories(q);
        setCategoryOptions(buildCascade(resp.items ?? []));
      } catch (err) {
        // ignore: 可能权限/网络问题
        // eslint-disable-next-line no-console
        console.warn('[CreateProductWizard] load categories failed', err);
      }
    })();
  }, [open]);

  const handleNext = async () => {
    try {
      const v = await form.validateFields();
      doCreate(
        {
          data: {
            code: v.code,
            name: v.name,
            nameEn: v.nameEn,
            categoryId: Number(v.categoryId[v.categoryId.length - 1]),
            manufacturer: v.manufacturer,
            modelNo: v.modelNo,
            description: v.description,
          },
        },
        {
          onSuccess: async () => {
            // 按 code 拿新产品 id
            const p = await fetchGetProduct({ code: v.code });
            setProductId(p.id ?? null);
            // 拉该分类的默认条目
            setLoadingDefaults(true);
            const cdfQ = new PaginationQuery({
              paging: { page: 1, pageSize: 500 },
              formValues: {
                category_id: Number(v.categoryId[v.categoryId.length - 1]),
              },
              orderBy: ['sort_order', 'id'],
            });
            const cdfResp = await fetchListCategoryDefaultFeatures(cdfQ);
            const items = cdfResp.items ?? [];
            setDefaults(items);
            setSelectedIds(items.map((x) => x.id!).filter(Boolean));
            setLoadingDefaults(false);
            setStep(1);
          },
          onError: (err) => message.error(err.message),
        },
      );
    } catch {
      /* form validation */
    }
  };

  const handlePullAndFinish = () => {
    if (!productId) return;
    doPull(
      {
        productId,
        defaultFeatureIds: selectedIds,
        onConflict,
      },
      {
        onSuccess: () => {
          message.success(t('createSuccess'));
          reset();
          onClose(true);
        },
        onError: (err) => message.error(err.message),
      },
    );
  };

  const handleSkipPull = () => {
    message.success(t('createSuccess'));
    reset();
    onClose(true);
  };

  const reset = () => {
    setStep(0);
    setProductId(null);
    setDefaults([]);
    setSelectedIds([]);
    form.resetFields();
  };

  const handleCancel = () => {
    reset();
    onClose(false);
  };

  return (
    <Modal
      open={open}
      title={t('createProduct')}
      width={760}
      onCancel={handleCancel}
      footer={null}
      destroyOnClose
    >
      <Steps
        current={step}
        items={[{ title: t('wizard.step1') }, { title: t('wizard.step2') }]}
      />
      <div style={{ marginTop: 24 }}>
        {step === 0 && (
          <Form form={form} layout="vertical">
            <Form.Item
              label={t('category')}
              name="categoryId"
              rules={[{ required: true, message: t('categoryRequired') }]}
            >
              <Cascader
                options={categoryOptions}
                placeholder={t('categoryPlaceholder')}
                showSearch
                changeOnSelect
                displayRender={(labels) => labels.join(' / ')}
              />
            </Form.Item>
            <Form.Item label={t('code')} name="code" rules={[{ required: true }]}>
              <Input placeholder="GREE-LSBLG320" />
            </Form.Item>
            <Form.Item label={t('name')} name="name" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item label={t('nameEn')} name="nameEn">
              <Input />
            </Form.Item>
            <Form.Item label={t('manufacturer')} name="manufacturer">
              <Input />
            </Form.Item>
            <Form.Item label={t('modelNo')} name="modelNo">
              <Input />
            </Form.Item>
            <Form.Item label={t('description')} name="description">
              <Input.TextArea rows={2} />
            </Form.Item>
            <div style={{ textAlign: 'right' }}>
              <AntButton type="primary" onClick={handleNext} loading={creating}>
                {t('common:next')}
              </AntButton>
            </div>
          </Form>
        )}
        {step === 1 && (
          <Spin spinning={loadingDefaults}>
            <p>{t('wizard.pullDescription')}</p>
            <Checkbox.Group
              value={selectedIds}
              onChange={(v) => setSelectedIds(v as number[])}
              style={{
                display: 'flex',
                flexDirection: 'column',
                gap: 8,
                maxHeight: 320,
                overflowY: 'auto',
              }}
            >
              {defaults.map((d) => (
                <Checkbox key={d.id} value={d.id}>
                  <strong>{d.featureCode}</strong> {d.featureName}
                  {d.overrideSpec && (
                    <span style={{ marginLeft: 8, color: '#1677ff' }}>{t('hasOverride')}</span>
                  )}
                </Checkbox>
              ))}
            </Checkbox.Group>
            <div style={{ marginTop: 16 }}>
              <span>{t('wizard.onConflict')}：</span>
              <Radio.Group value={onConflict} onChange={(e) => setOnConflict(e.target.value)}>
                <Radio value="SKIP">{t('wizard.skip')}</Radio>
                <Radio value="REPLACE">{t('wizard.replace')}</Radio>
              </Radio.Group>
            </div>
            <div style={{ textAlign: 'right', marginTop: 16 }}>
              <AntButton onClick={handleSkipPull} disabled={pulling}>
                {t('wizard.skipPull')}
              </AntButton>
              <AntButton
                type="primary"
                style={{ marginLeft: 8 }}
                onClick={handlePullAndFinish}
                loading={pulling}
              >
                {t('wizard.createAndPull')}
              </AntButton>
            </div>
          </Spin>
        )}
      </div>
    </Modal>
  );
};

export default CreateProductWizard;
