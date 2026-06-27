package seed

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/xuri/excelize/v2"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// feature_seed_template_test.go: 从 AllFeatureSeeds() 生成特征导入模板与完整种子 Excel。
//
// 用途 / Purpose:
//   - docs/thingmodel/sheji/14-特征种子导入模板.xlsx        —— 仅表头 + 2 行示例（日常新增用）
//   - docs/thingmodel/sheji/14-特征种子数据-完整.xlsx        —— 255 条完整种子（保底恢复用）
//
// 运行 / Run:
//   go test ./app/admin/service/internal/data/seed/ -run TestGenerateFeatureImportTemplate -v
//
// 列布局（与后端 ImportFeatureRow proto、前端 ImportFeaturesModal 解析器一一对应）：
//   featureType | code | identifier | name | nameEn | description | applicableScope | sortOrder | specJson
//   其中 specJson 是 spec map 的 JSON 字符串（与种子同构），导入时由后端 buildFeatureSpecProto 还原。

// 导入列顺序常量（单一事实来源，后端/前端/Excel 共用）。
var featureImportColumns = []string{
	"featureType", "code", "identifier", "name", "nameEn",
	"description", "applicableScope", "sortOrder", "specJson",
}

// seedRowToImportRow 把一条 SeedFeature 转成 Excel 行（[]any，顺序同 featureImportColumns）。
func seedRowToImportRow(f SeedFeature) ([]any, error) {
	specJSON := ""
	if f.Spec != nil {
		b, err := json.Marshal(f.Spec)
		if err != nil {
			return nil, fmt.Errorf("marshal spec for %s: %w", f.Code, err)
		}
		specJSON = string(b)
	}
	return []any{
		f.FeatureType.String(),
		f.Code,
		f.Identifier,
		f.Name,
		f.NameEn,
		f.Description,
		f.ApplicableScope,
		f.SortOrder,
		specJSON,
	}, nil
}

// writeFeatureExcel 构造一个 .xlsx：表头 + 给定数据行，保存到 outPath。
func writeFeatureExcel(outPath string, rows [][]any) error {
	f := excelize.NewFile()
	defer f.Close()

	sheet := f.GetSheetName(0)
	// 表头
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E8F4FF"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "center", WrapText: true},
	})
	for c, col := range featureImportColumns {
		cell, _ := excelize.CoordinatesToCellName(c+1, 1)
		if err := f.SetCellValue(sheet, cell, col); err != nil {
			return err
		}
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
	}
	// 数据行
	for r, row := range rows {
		for c, val := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			if err := f.SetCellValue(sheet, cell, val); err != nil {
				return err
			}
		}
	}
	// 列宽
	widths := map[string]float64{
		"A": 12, "B": 18, "C": 22, "D": 18, "E": 22,
		"F": 40, "G": 22, "H": 10, "I": 80,
	}
	for col, w := range widths {
		_ = f.SetColWidth(sheet, col, col, w)
	}
	// 冻结首行（自动筛选非必需，省略以避免不同 excelize 版本签名差异）
	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})
	return f.SaveAs(outPath)
}

// docsThingmodelShejiDir 返回 docs/thingmodel/sheji 的绝对路径（基于本测试文件位置回溯）。
func docsThingmodelShejiDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test file path")
	}
	// 本文件在 .../backend/app/admin/service/internal/data/seed/
	// docs 在 .../go-wind-admin/docs/thingmodel/sheji
	// 回溯到 go-wind-admin 根：backend/app/admin/service/internal/data/seed -> 上 7 层到根
	// 但跨平台用 filepath 连接更稳：seed -> internal/data -> ... -> backend -> root
	dir := filepath.Dir(file)
	root := dir
	for i := 0; i < 7; i++ {
		root = filepath.Dir(root)
	}
	target := filepath.Join(root, "docs", "thingmodel", "sheji")
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("docs dir not found at %s (from %s): %v", target, dir, err)
	}
	return target
}

// TestGenerateFeatureImportTemplate 生成两张 Excel：完整种子 + 空模板（2 行示例）。
// 该测试通过写文件产生交付物，默认在 CI 之外手动运行；用 -run 显式触发，避免误报。
func TestGenerateFeatureImportTemplate(t *testing.T) {
	all := AllFeatureSeeds()
	// 按 sortOrder 稳定排序，保证 Excel 可读性（与录入顺序一致）
	sort.SliceStable(all, func(i, j int) bool { return all[i].SortOrder < all[j].SortOrder })

	// ===== 1) 完整种子 Excel =====
	fullRows := make([][]any, 0, len(all))
	for _, f := range all {
		row, err := seedRowToImportRow(f)
		if err != nil {
			t.Fatal(err)
		}
		fullRows = append(fullRows, row)
	}
	dir := docsThingmodelShejiDir(t)
	fullPath := filepath.Join(dir, "14-特征种子数据-完整.xlsx")
	if err := writeFeatureExcel(fullPath, fullRows); err != nil {
		t.Fatalf("write full seed xlsx: %v", err)
	}
	t.Logf("✅ 完整种子 Excel: %s (%d 行)", fullPath, len(fullRows))

	// ===== 2) 空模板（表头 + 2 行示例：property + event）=====
	example := []SeedFeature{
		{ftProperty, "P-HVAC-9999", "exampleTemperature", "示例温度", "Example Temp",
			"这是一个示例属性行，导入前请删除或修改", "空调末端", 9999,
			map[string]any{
				"dataType": "DOUBLE", "accessMode": "R", "category": "measurement",
				"unit":       map[string]any{"unitCode": "celsius", "unitSymbol": "℃"},
				"constraints": map[string]any{"min": -40, "max": 80},
			}},
		{ftEvent, "E-HVAC-9999", "exampleAlarm", "示例告警", "Example Alarm",
			"这是一个示例事件行，导入前请删除或修改", "冷机", 9999,
			map[string]any{"level": "ALERT", "outputParams": []map[string]any{
				{"key": "value", "dataType": "DOUBLE"},
				{"key": "limit", "dataType": "DOUBLE"},
			}}},
	}
	tplRows := make([][]any, 0, len(example))
	for _, f := range example {
		row, err := seedRowToImportRow(f)
		if err != nil {
			t.Fatal(err)
		}
		tplRows = append(tplRows, row)
	}
	tplPath := filepath.Join(dir, "14-特征种子导入模板.xlsx")
	if err := writeFeatureExcel(tplPath, tplRows); err != nil {
		t.Fatalf("write template xlsx: %v", err)
	}
	t.Logf("✅ 空模板 Excel: %s (%d 行示例)", tplPath, len(tplRows))

	// ===== 3) 校验：生成的完整 Excel 行数 == AllFeatureSeeds 数量 =====
	if len(fullRows) != len(all) {
		t.Fatalf("row count mismatch: excel=%d, seeds=%d", len(fullRows), len(all))
	}
	// 按类型统计便于日志确认
	cnt := map[thingmodelV1.FeatureType]int{}
	for _, f := range all {
		cnt[f.FeatureType]++
	}
	t.Logf("统计: property=%d, event=%d, service=%d, relation=%d",
		cnt[ftProperty], cnt[ftEvent], cnt[ftService], cnt[ftRelation])
}
