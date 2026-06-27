import { useState } from 'react';
import { Modal, Upload, Button, Alert, Space, Typography, Tag, App } from 'antd';
import { InboxOutlined, DownloadOutlined } from '@ant-design/icons';
import type { UploadProps } from 'antd';
import ExcelJS from 'exceljs';
import { useTranslation } from 'react-i18next';

import { useImportFeatures } from '@/api/hooks/feature';
import type { thingmodelservicev1_ImportFeatureRow } from '@/api/generated/admin/service/v1';

const { Dragger } = Upload;
const { Text, Paragraph } = Typography;

// 导入列顺序（与后端 ImportFeatureRow、Excel 表头、Go 种子生成器一致）。
// Import column order (matches backend ImportFeatureRow, Excel header, and Go seed generator).
const IMPORT_COLUMNS = [
  'featureType',
  'code',
  'identifier',
  'name',
  'nameEn',
  'description',
  'applicableScope',
  'sortOrder',
  'specJson',
] as const;

// 公共模板/种子文件路径（构建期从 docs 拷贝到 public/templates）。
const TEMPLATE_URL = '/templates/feature-import-template.xlsx';
const FULL_SEED_URL = '/templates/feature-seed-full.xlsx';

interface ImportFeaturesModalProps {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

/**
 * 导入特征弹窗（保底方案）/ Import features modal (fallback).
 *
 * 流程：选 .xlsx → exceljs 客户端解析 → 调 ImportFeatures（按 code 幂等 upsert）
 *   → 展示 succeeded/failed 与失败明细。
 *
 * Excel 列：featureType | code | identifier | name | nameEn | description |
 *           applicableScope | sortOrder | specJson(spec 的 JSON 字符串)
 */
const ImportFeaturesModal: React.FC<ImportFeaturesModalProps> = ({
  open,
  onClose,
  onSuccess,
}) => {
  const { t } = useTranslation('feature');
  const { message } = App.useApp();

  const [rows, setRows] = useState<thingmodelservicev1_ImportFeatureRow[]>([]);
  const [fileName, setFileName] = useState<string>('');
  const [parseError, setParseError] = useState<string>('');
  const [result, setResult] = useState<{
    succeeded?: number;
    failed?: number;
    total?: number;
    errors?: string[];
  } | null>(null);

  const importMutation = useImportFeatures({
    onSuccess: (resp) => {
      setResult(resp);
      const failed = resp.failed ?? 0;
      if (failed === 0) {
        message.success(
          t('importSuccess', {
            succeeded: resp.succeeded ?? 0,
            failed,
            total: resp.total ?? 0,
          }),
        );
        onSuccess();
      } else {
        message.warning(
          t('importSuccess', {
            succeeded: resp.succeeded ?? 0,
            failed,
            total: resp.total ?? 0,
          }),
        );
        onSuccess(); // 即使部分失败，也刷新列表（成功的已落库）
      }
    },
    onError: (error: Error) => {
      message.error(t('importFailed', { error: error.message }));
    },
  });

  // 重置内部状态（每次打开/关闭时）
  const reset = () => {
    setRows([]);
    setFileName('');
    setParseError('');
    setResult(null);
  };

  // exceljs 解析 .xlsx：表头=IMPORT_COLUMNS，逐行组装 ImportFeatureRow。
  const parseWorkbook = async (file: File): Promise<thingmodelservicev1_ImportFeatureRow[]> => {
    const buf = await file.arrayBuffer();
    const wb = new ExcelJS.Workbook();
    await wb.xlsx.load(buf);
    const ws = wb.getWorksheet(1);
    if (!ws) return [];

    // 读表头，建立 列号 → 字段名 映射（兼容列顺序微调）
    const headerMap: Record<number, string> = {};
    ws.getRow(1).eachCell((cell, colNumber) => {
      const h = String(cell.value ?? '').trim();
      if (IMPORT_COLUMNS.includes(h as (typeof IMPORT_COLUMNS)[number])) {
        headerMap[colNumber] = h;
      }
    });

    const out: thingmodelservicev1_ImportFeatureRow[] = [];
    for (let r = 2; r <= ws.rowCount; r++) {
      const row = ws.getRow(r);
      // 跳过完全空行（以 code 列是否为空为准）
      const rowData: Partial<thingmodelservicev1_ImportFeatureRow> = {};
      let hasAny = false;
      Object.entries(headerMap).forEach(([colStr, field]) => {
        const cell = row.getCell(Number(colStr));
        const val = cell.value;
        // 处理 exceljs 的富文本/超链接对象
        let strVal: string | undefined;
        if (val == null) strVal = undefined;
        else if (typeof val === 'object') {
          // { richText: [...] } / { text, hyperlink }
          const anyVal = val as any;
          strVal = anyVal.text ?? (Array.isArray(anyVal.richText) ? anyVal.richText.map((rt: any) => rt.text).join('') : String(val));
        } else {
          strVal = String(val);
        }
        if (field === 'sortOrder') {
          (rowData as any)[field] = val == null ? undefined : Number(val);
        } else {
          (rowData as any)[field] = strVal?.trim() || undefined;
        }
        if (strVal && strVal.trim()) hasAny = true;
      });
      if (!hasAny) continue;
      // code 必填，缺失则跳过（后端也会拒）
      if (!rowData.code) continue;
      out.push(rowData as thingmodelservicev1_ImportFeatureRow);
    }
    return out;
  };

  const draggerProps: UploadProps = {
    name: 'file',
    multiple: false,
    accept: '.xlsx,.xls',
    showUploadList: false,
    beforeUpload: (file) => {
      reset();
      setFileName(file.name);
      parseWorkbook(file)
        .then((parsed) => {
          if (parsed.length === 0) {
            setParseError(t('importNoData'));
            return;
          }
          setRows(parsed);
        })
        .catch((err: Error) => {
          setParseError(t('importReadFailed', { error: err.message }));
        });
      return false; // 阻止 antd 自动上传
    },
  };

  const handleImport = () => {
    if (rows.length === 0) return;
    importMutation.mutate(rows);
  };

  const handleClose = () => {
    reset();
    onClose();
  };

  const downloading = (url: string) => {
    // 简单的 a 标签下载（public 静态资源）
    const a = document.createElement('a');
    a.href = url;
    a.download = '';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
  };

  return (
    <Modal
      title={t('importTitle')}
      open={open}
      onCancel={handleClose}
      width={640}
      footer={[
        <Button key="cancel" onClick={handleClose}>
          {t('importConfirmRetry')}
        </Button>,
        <Button
          key="import"
          type="primary"
          loading={importMutation.isPending}
          disabled={rows.length === 0 || !!parseError}
          onClick={handleImport}
        >
          {t('importStart')}
        </Button>,
      ]}
    >
      <Space direction="vertical" size="middle" style={{ width: '100%' }}>
        <Paragraph type="secondary" style={{ marginBottom: 0 }}>
          {t('importDesc')}
        </Paragraph>

        <Space>
          <Button size="small" icon={<DownloadOutlined />} onClick={() => downloading(TEMPLATE_URL)}>
            {t('importDownloadTemplate')}
          </Button>
          <Button size="small" icon={<DownloadOutlined />} onClick={() => downloading(FULL_SEED_URL)}>
            {t('importDownloadFullSeed')}
          </Button>
        </Space>

        <Dragger {...draggerProps} style={{ padding: 8 }}>
          <p className="ant-upload-drag-icon">
            <InboxOutlined />
          </p>
          <p className="ant-upload-text">{t('importDragHint')}</p>
        </Dragger>

        {fileName && !parseError && rows.length === 0 && (
          <Text type="secondary">{t('importSelected', { name: fileName })}</Text>
        )}
        {fileName && !parseError && rows.length > 0 && (
          <Alert
            type="info"
            showIcon
            message={t('importParsed', { count: rows.length })}
            description={
              <Text type="secondary">
                {t('importSelected', { name: fileName })}
              </Text>
            }
          />
        )}
        {parseError && <Alert type="error" showIcon message={parseError} />}

        {result && (
          <>
            <Alert
              type={(result.failed ?? 0) > 0 ? 'warning' : 'success'}
              showIcon
              message={t('importSuccess', {
                succeeded: result.succeeded ?? 0,
                failed: result.failed ?? 0,
                total: result.total ?? 0,
              })}
            />
            {(result.failed ?? 0) > 0 && (result.errors?.length ?? 0) > 0 && (
              <div>
                <Text strong>{t('importErrorListTitle')}</Text>
                <div style={{ marginTop: 8, maxHeight: 200, overflow: 'auto' }}>
                  {(result.errors ?? []).map((e, i) => (
                    <Tag key={i} color="red" style={{ margin: 2, whiteSpace: 'normal' }}>
                      {e}
                    </Tag>
                  ))}
                </div>
              </div>
            )}
          </>
        )}
      </Space>
    </Modal>
  );
};

export default ImportFeaturesModal;
