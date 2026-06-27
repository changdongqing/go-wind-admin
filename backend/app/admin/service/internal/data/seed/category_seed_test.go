package seed

import (
	"testing"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// TestParseCategoryMarkdown_System 验证智能系统清单解析正确性。
// 预期：7 个大类（30-36）、每个细类 8 位、L1+L2+L3+L4 节点数与文档一致。
func TestParseCategoryMarkdown_System(t *testing.T) {
	nodes := parseCategoryMarkdown(categorySystemMd, thingmodelV1.CategoryKind_SYSTEM)

	if len(nodes) == 0 {
		t.Fatalf("system markdown should parse non-zero nodes")
	}

	// 按 level 统计
	byLevel := map[uint8]int{}
	codes := map[string]string{} // code → name
	for _, n := range nodes {
		byLevel[n.Level]++
		codes[n.Code] = n.Name
		if uint32(len(n.Code)) != uint32(n.Level)*2 {
			t.Errorf("node %s level=%d 长度应为 %d，实际 %d", n.Code, n.Level, n.Level*2, len(n.Code))
		}
	}

	// 7 个大类：30/31/32/33/34/35/36
	wantBigCodes := []string{"30", "31", "32", "33", "34", "35", "36"}
	for _, bc := range wantBigCodes {
		if _, ok := codes[bc]; !ok {
			t.Errorf("missing big category code=%s", bc)
		}
	}
	if byLevel[1] != 7 {
		t.Errorf("level=1 应为 7，实际 %d", byLevel[1])
	}

	// 抽查链路：30 / 3001 / 300101 / 30010100
	pairs := map[string]string{
		"30":       "暖通空调系统",
		"3001":     "冷热源系统",
		"300101":   "制冷系统",
		"30010100": "压缩式制冷系统",
	}
	for code, wantName := range pairs {
		got, ok := codes[code]
		if !ok {
			t.Errorf("missing code=%s", code)
			continue
		}
		if got != wantName {
			t.Errorf("code=%s name mismatch: want=%q got=%q", code, wantName, got)
		}
	}

	t.Logf("SYSTEM nodes total=%d, byLevel=%v", len(nodes), byLevel)
}

// TestParseCategoryMarkdown_Space 验证空间清单解析正确性（仅 1 大类）。
func TestParseCategoryMarkdown_Space(t *testing.T) {
	nodes := parseCategoryMarkdown(categorySpaceMd, thingmodelV1.CategoryKind_SPACE)
	if len(nodes) == 0 {
		t.Fatalf("space markdown should parse non-zero nodes")
	}

	byLevel := map[uint8]int{}
	codes := map[string]string{}
	for _, n := range nodes {
		byLevel[n.Level]++
		codes[n.Code] = n.Name
	}
	if byLevel[1] != 1 {
		t.Errorf("SPACE level=1 应为 1（仅 10 段），实际 %d", byLevel[1])
	}
	if codes["10"] != "建筑空间" {
		t.Errorf("code=10 name 应为'建筑空间', got=%q", codes["10"])
	}
	if codes["10010100"] != "办公空间" {
		t.Errorf("code=10010100 name 应为'办公空间', got=%q", codes["10010100"])
	}
	t.Logf("SPACE nodes total=%d, byLevel=%v", len(nodes), byLevel)
}

// TestParseCategoryMarkdown_Facility 验证设备设施清单解析正确性（7 大类 20-26）。
func TestParseCategoryMarkdown_Facility(t *testing.T) {
	nodes := parseCategoryMarkdown(categoryFacilityMd, thingmodelV1.CategoryKind_FACILITY)
	if len(nodes) == 0 {
		t.Fatalf("facility markdown should parse non-zero nodes")
	}

	byLevel := map[uint8]int{}
	codes := map[string]string{}
	for _, n := range nodes {
		byLevel[n.Level]++
		codes[n.Code] = n.Name
	}
	if byLevel[1] != 7 {
		t.Errorf("FACILITY level=1 应为 7（20-26），实际 %d", byLevel[1])
	}
	wantBigCodes := []string{"20", "21", "22", "23", "24", "25", "26"}
	for _, bc := range wantBigCodes {
		if _, ok := codes[bc]; !ok {
			t.Errorf("missing facility big category code=%s", bc)
		}
	}
	if codes["20010100"] != "电动压缩式冷水机组" {
		t.Errorf("code=20010100 name 应为'电动压缩式冷水机组', got=%q", codes["20010100"])
	}
	t.Logf("FACILITY nodes total=%d, byLevel=%v", len(nodes), byLevel)
}

// TestSortCategoryNodes_ParentBeforeChild 验证排序结果父先于子。
func TestSortCategoryNodes_ParentBeforeChild(t *testing.T) {
	seen := map[string]*seedCategoryNode{}
	addAll(seen, parseCategoryMarkdown(categorySystemMd, thingmodelV1.CategoryKind_SYSTEM))
	addAll(seen, parseCategoryMarkdown(categorySpaceMd, thingmodelV1.CategoryKind_SPACE))
	addAll(seen, parseCategoryMarkdown(categoryFacilityMd, thingmodelV1.CategoryKind_FACILITY))

	nodes := sortCategoryNodes(seen)

	// (kind|code) → seen index
	idx := map[string]int{}
	for i, n := range nodes {
		idx[categoryKey(n.Kind, n.Code)] = i
	}
	for i, n := range nodes {
		if n.Level <= 1 {
			continue
		}
		parentCode := n.Code[:len(n.Code)-2]
		pk := categoryKey(n.Kind, parentCode)
		pi, ok := idx[pk]
		if !ok {
			t.Errorf("node %s/%s missing parent %s in sorted list", n.Kind, n.Code, parentCode)
			continue
		}
		if pi >= i {
			t.Errorf("parent %s/%s (i=%d) should precede child %s/%s (i=%d)",
				n.Kind, parentCode, pi, n.Kind, n.Code, i)
		}
	}

	t.Logf("combined total nodes=%d", len(nodes))
}
