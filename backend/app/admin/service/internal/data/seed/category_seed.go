// Package seed — 物模型分类种子（v2.0：变长 code + kind 枚举 + 4 层）。
// Package seed — thing-model category seed (v2.0: variable-length code + kind enum + 4 levels).
//
// 设计依据 / Design ref: docs/thingmodel/sheji/分类管理/06-种子数据与实施计划.md
//
// 数据源 / Data source:
//   - seed/categories/system.md   （智能系统分类清单，30-36 段 7 大类）
//   - seed/categories/space.md    （空间分类清单，10 段 1 大类）
//   - seed/categories/facility.md （设备设施分类清单，20-26 段 7 大类）
//
// 解析与派生算法 / Parse & derive algorithm:
//   1. 按 "## NN 标题" 切段识别大类（level=1, code=NN，直接 2 位不补 0）
//   2. 解析表格每行得到 (code8, 大类名, 中类名, 小类名, 细类名)
//   3. 由 code8 截断派生：中类 = code8[:4]、小类 = code8[:6]、细类 = code8
//   4. 去重：用 (kind, code) 作为唯一键
//   5. 按 level 升序 → code 升序入库，保证父节点先于子节点
//   6. parent_id 通过 (kind, parent_code = self_code[:-2]) 在已写入缓存中查
//
// 幂等策略 / Idempotency:
//   - 按 (tenant_id=0, kind, code) upsert；重复执行不报错也不复制。
//   - 仅维护 tenant_id=0 的系统预置数据；租户自建（tenant_id>0）永不被覆盖。
package seed

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"go-wind-admin/app/admin/service/internal/data/ent"
	"go-wind-admin/app/admin/service/internal/data/ent/category"
	appViewer "go-wind-admin/pkg/entgo/viewer"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

//go:embed categories/system.md
var categorySystemMd string

//go:embed categories/space.md
var categorySpaceMd string

//go:embed categories/facility.md
var categoryFacilityMd string

// seedCategoryNode 单个分类节点的中间表示。
// seedCategoryNode is the intermediate representation of one category node.
type seedCategoryNode struct {
	Kind  thingmodelV1.CategoryKind
	Code  string // 变长 2/4/6/8 位
	Level uint8  // 1..4，= len(Code) / 2
	Name  string
}

// SeedThingmodelCategories 将三套国标清单（智能系统/空间/设备设施）写入数据库（幂等 upsert）。
// SeedThingmodelCategories upserts three GB-standard classification lists into the DB idempotently.
//
// 重要：内部把 ctx 包成 SystemViewerContext，否则 TenantPrivacy 策略会拒绝写入 tenant_id=0 的全局数据。
// IMPORTANT: wraps ctx with SystemViewerContext, otherwise TenantPrivacy rejects writes for tenant_id=0.
func SeedThingmodelCategories(ctx context.Context, client *ent.Client, logger *log.Helper) error {
	if client == nil {
		return fmt.Errorf("seed: ent client is nil")
	}
	if logger == nil {
		logger = log.NewHelper(log.With(log.GetLogger(), "module", "thingmodel-category-seed"))
	}

	// 系统视图绕过 TenantPrivacy 过滤 / System viewer bypasses TenantPrivacy
	ctx = appViewer.NewSystemViewerContext(ctx)

	// 1) 解析三份 markdown，合并去重
	seen := make(map[string]*seedCategoryNode, 800)
	addAll(seen, parseCategoryMarkdown(categorySystemMd, thingmodelV1.CategoryKind_SYSTEM))
	addAll(seen, parseCategoryMarkdown(categorySpaceMd, thingmodelV1.CategoryKind_SPACE))
	addAll(seen, parseCategoryMarkdown(categoryFacilityMd, thingmodelV1.CategoryKind_FACILITY))

	// 2) 按 (kind, level, code) 排序 → 保证父先于子
	nodes := sortCategoryNodes(seen)

	// 3) 逐条 upsert；维护 (kind|code) → id 缓存供子节点回填 parent_id
	idByKey := make(map[string]uint32, len(nodes))
	now := time.Now()
	const sysTenant uint32 = 0

	totalByKind := map[thingmodelV1.CategoryKind]int{}

	for _, n := range nodes {
		var parentID uint32
		if n.Level > 1 {
			parentCode := n.Code[:len(n.Code)-2]
			pkey := categoryKey(n.Kind, parentCode)
			parentID = idByKey[pkey]
			if parentID == 0 {
				return fmt.Errorf("seed category: parent not seeded yet: %s/%s", n.Kind, parentCode)
			}
		}

		id, err := upsertCategoryNode(ctx, client, n, parentID, sysTenant, now)
		if err != nil {
			logger.Errorf("seed category %s/%s/L%d failed: %v", n.Kind, n.Code, n.Level, err)
			return fmt.Errorf("upsert category %s/%s/L%d: %w", n.Kind, n.Code, n.Level, err)
		}

		idByKey[categoryKey(n.Kind, n.Code)] = id
		totalByKind[n.Kind]++
	}

	logger.Infof("[seed] thingmodel categories: SYSTEM=%d, SPACE=%d, FACILITY=%d, total=%d",
		totalByKind[thingmodelV1.CategoryKind_SYSTEM],
		totalByKind[thingmodelV1.CategoryKind_SPACE],
		totalByKind[thingmodelV1.CategoryKind_FACILITY],
		len(nodes),
	)
	return nil
}

// upsertCategoryNode 按 (tenant_id, kind, code) upsert 单个分类节点，返回其 id。
// upsertCategoryNode upserts one category by (tenant_id, kind, code) and returns its id.
//
// 关键点 / Key points:
//   - OnConflictColumns 用 (tenant_id, kind, code) 三元组——v2 唯一键无需 level。
//   - 不维护 reference_count（保留运行时累计值，由未来 thing_property / *_instance 模块写入）。
func upsertCategoryNode(
	ctx context.Context,
	client *ent.Client,
	n *seedCategoryNode,
	parentID uint32,
	sysTenant uint32,
	now time.Time,
) (uint32, error) {
	// 把 proto kind 转成 ent kind 字符串
	entKind := category.Kind(n.Kind.String())

	builder := client.Category.Create().
		SetTenantID(sysTenant).
		SetKind(entKind).
		SetCode(n.Code).
		SetLevel(n.Level).
		SetName(n.Name).
		SetIsEnabled(true).
		SetSortOrder(0).
		SetCreatedAt(now)

	if parentID != 0 {
		builder.SetParentID(parentID)
	}

	if err := builder.
		OnConflictColumns(category.FieldTenantID, category.FieldKind, category.FieldCode).
		Update(func(up *ent.CategoryUpsert) {
			// 重要：不覆盖 reference_count，避免清空运行时累计值。
			// Do NOT overwrite reference_count to preserve runtime counters.
			up.UpdateName().
				UpdateLevel().
				UpdateParentID().
				SetUpdatedAt(now)
		}).
		Exec(ctx); err != nil {
		return 0, err
	}

	// 取回 id（OnConflict 之后必须二次查询）
	cat, err := client.Category.Query().
		Where(
			category.TenantIDEQ(sysTenant),
			category.KindEQ(entKind),
			category.CodeEQ(n.Code),
		).
		Only(ctx)
	if err != nil {
		return 0, fmt.Errorf("query category after upsert: %w", err)
	}
	return cat.ID, nil
}

// ===== 解析与排序 / Parsing & sorting =====

var (
	categoryH2Regex  = regexp.MustCompile(`^##\s+(\d{2})\s+(.+?)\s*$`)
	categoryRowRegex = regexp.MustCompile(
		`^\|\s*(\d{8})\s*\|\s*([^|]+?)\s*\|\s*([^|]+?)\s*\|\s*([^|]+?)\s*\|\s*([^|]+?)\s*\|`)
)

// parseCategoryMarkdown 解析一份分类清单 markdown，返回去重后的节点列表。
// parseCategoryMarkdown parses one classification markdown into a deduplicated node list.
//
// 解析规则 / Parsing rules:
//   - H2 标题 "## NN 大类名" → 直接生成 level=1 节点（code=NN，2 位）。
//   - 表行 "| 8 位 code | 大类 | 中类 | 小类 | 细类 |" → 由 code8 派生 4 个节点。
//   - 同 (kind, code) 名字以**首次出现**为准（容错：清单里大类列与 H2 标题略有差异时以 H2 为准）。
//   - "## 分类说明" 等非数字 H2 自动忽略（正则要求 H2 后跟 2 位数字）。
func parseCategoryMarkdown(raw string, kind thingmodelV1.CategoryKind) []*seedCategoryNode {
	// (kind, code) → name；用 map 去重
	seen := make(map[string]string, 256)
	addOnce := func(code, name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		if _, ok := seen[code]; !ok {
			seen[code] = name
		}
	}

	var currentBigCode string
	var currentBigName string

	for _, rawLine := range strings.Split(raw, "\n") {
		line := strings.TrimRight(rawLine, "\r")

		// 1) 大类 H2 标题
		if m := categoryH2Regex.FindStringSubmatch(line); m != nil {
			currentBigCode, currentBigName = m[1], m[2]
			addOnce(currentBigCode, currentBigName) // 大类（2 位）
			continue
		}

		// 2) 表行：8 位 code + 大类/中类/小类/细类
		m := categoryRowRegex.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		code8 := m[1]
		bigFromRow := m[2]
		midName := m[3]
		smallName := m[4]
		leafName := m[5]

		bigCode := code8[:2]
		midCode := code8[:4]
		smallCode := code8[:6]

		// 容错：表格行的大类列可能与 H2 标题略有差异；若 H2 未读到，则以表格行为准
		if _, exists := seen[bigCode]; !exists {
			if currentBigName != "" && bigCode == currentBigCode {
				addOnce(bigCode, currentBigName)
			} else {
				addOnce(bigCode, bigFromRow)
			}
		}

		addOnce(midCode, midName)
		addOnce(smallCode, smallName)
		addOnce(code8, leafName)
	}

	out := make([]*seedCategoryNode, 0, len(seen))
	for code, name := range seen {
		out = append(out, &seedCategoryNode{
			Kind:  kind,
			Code:  code,
			Level: uint8(len(code) / 2),
			Name:  name,
		})
	}
	return out
}

// addAll 把切片节点合入 (kind|code) → node 全局 map（首次出现保留）。
// addAll merges nodes into the global (kind|code) map; first-wins on duplicates.
func addAll(into map[string]*seedCategoryNode, nodes []*seedCategoryNode) {
	for _, n := range nodes {
		k := categoryKey(n.Kind, n.Code)
		if _, ok := into[k]; !ok {
			into[k] = n
		}
	}
}

// sortCategoryNodes 按 (kind, level, code) 升序排序——保证父节点先于子节点入库。
// sortCategoryNodes orders nodes so parents precede children.
func sortCategoryNodes(m map[string]*seedCategoryNode) []*seedCategoryNode {
	out := make([]*seedCategoryNode, 0, len(m))
	for _, n := range m {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		if out[i].Level != out[j].Level {
			return out[i].Level < out[j].Level
		}
		return out[i].Code < out[j].Code
	})
	return out
}

// categoryKey 生成 (kind|code) 复合键。
// categoryKey builds the composite key "<kind>|<code>".
func categoryKey(kind thingmodelV1.CategoryKind, code string) string {
	return kind.String() + "|" + code
}
