# 四、与 open-design 代理本地 CLI 方式的深度对比

> 配套：`README.md`（总览）、`01-ui-layout.md`、`02-cli-connection.md`、`03-implementation-plan.md`。
>
> 本文是 2026-06 评审产出，回答两个问题：**(1) 本方案与 [nexu-io/open-design](https://github.com/nexu-io/open-design) 代理本机 CLI 的方式是否一样？(2) 各有什么优缺点？** 并为 `02`/`03` 的 v2 修订提供溯源。

---

## 0. 结论先行

**核心机制——两者完全一样**：都是 `child_process.spawn(cli, args, {cwd, shell:false})` 拉起本机已装 Agent CLI → 逐行解析 stdout 流式协议 → 归一化成类型化事件 → 喂给 UI；都用 per-CLI adapter 屏蔽协议差异；都靠 PATH 探测 + env 注入凭证。这是 Cursor / Cline / Continue / open-design 共用的标准范式，本方案选它没有问题。

**但在架构部署、权限模型、降级路径、CLI 广度上差别很大**，且 open-design 的几个选择恰好暴露了本方案的风险点（最关键：交互式权限、无 BYOK 降级）。所以「一样」只在「spawn 这一招」上成立，往上两层不一样。

> **资料口径**：以 [open-design README 原文](https://github.com/nexu-io/open-design)（2026-06 抓取）为准。早期搜索摘要称其后端为 "Go + Chi + WebSocket + PostgreSQL/pgvector"，与 README 不符（实为 Node 24 + Express + SSE + better-sqlite3，daemon 源码是 TypeScript），已摒弃。

---

## 1. open-design 代理本机 CLI 的真实方式

形态：**Web 应用（Next.js 16）+ 本地 daemon（Node 24 + Express + better-sqlite3，唯一特权进程）+ 可选 Electron 外壳**。

```
浏览器 (Next.js 16)
   │  /api/*  +  /api/chat (SSE)
   ▼
本地 daemon (Node + Express + SQLite)        ← 唯一特权进程
   │  spawn(cli, [...], { cwd: .od/projects/<id>/ })
   ▼
claude · codex · gemini · opencode · cursor-agent · qwen · copilot · …  (16 个 CLI)
```

关键事实（均出自 README，可核验）：

1. **PATH 扫描 + 16 CLI 自适应**：daemon 启动扫 PATH，自动发现 Claude Code / Codex / Devin / Cursor Agent / Gemini / OpenCode / Qwen / Qoder / Copilot / Hermes / Kimi / Pi / Kiro / Kilo / Mistral Vibe / DeepSeek。每个 CLI 在 `apps/daemon/src/agents.ts` 的 `AGENT_DEFS` 一条记录，流式格式七类：`claude-stream-json` / `qoder-stream-json` / `copilot-stream-json` / `json-event-stream`(per-CLI parser) / `acp-json-rpc`(Agent Client Protocol) / `pi-rpc` / `plain`。**加一个 CLI = 加一行**。

2. **BYOK proxy = "same loop minus the spawn"**：没装任何 CLI 时走 `POST /api/proxy/{anthropic,openai,azure,google,ollama,senseaudio}/stream`，daemon 把上游 SSE 归一成同样的 `delta/end/error`。**spawn 与 BYOK 吐同一种 typed-event 流**，上层 Skills/Design Systems 完全不感知来源。带 SSRF 防护（放行 loopback 本地 LLM，拒绝内网/CGNAT/重定向）。

3. **权限模型 = 文件夹沙箱 + bypass（与本方案最大哲学差异）**：Claude Code `--permission-mode bypassPermissions`、OpenCode `--dangerously-skip-permissions`、Gemini `--yolo`、Devin `--permission-mode dangerous`、Copilot `--allow-all-tools` —— **全部不做交互式权限询问**。安全边界是 cwd 本身：`.od/projects/<id>/` 是一次性产出目录，agent 在其中自由 Read/Write/Bash/WebFetch，产出设计物，不动用户真实代码。

4. **daemon 是唯一特权进程**（借鉴 `multica-ai/multica`）：集中管 spawn、SQLite 持久化（projects/conversations/messages/tabs/templates）、BYOK proxy、文件工作区。Web 与 Desktop 两种外壳连同一个 daemon。

5. **反向 MCP**：open-design 还自带只读 stdio MCP server（`search_files`/`get_file`/`get_artifact`），让别的 repo 里的 agent 反过来读 open-design 项目文件。即它是**双向**的：既 spawn 驱动 CLI，又作为 MCP 工具被 CLI 调用。

6. **Windows argv 坑**：每个 adapter 处理 CreateProcess ~32KB argv 上限——长 prompt 走 stdin，stdin 也溢出再退回临时 prompt-file。

7. **真实 bug（[issue #3146](https://github.com/nexu-io/open-design/issues/3146)）**：daemon spawn 的 claude 跑完 API 调用却不退出，导致 stdout→parser 的泵（`claude.feed`）从未触发，UI 收不到任何事件。说明**事件发射被绑死进程生命周期**是真实坑。

---

## 2. 逐轴对比

| 维度 | 本方案（antdesignx） | open-design | 是否一样 |
|------|----------------------|-------------|---------|
| spawn 机制 | `child_process.spawn`，`shell:false` | 同 | ✅ 一样 |
| stdout 解析→归一化事件 | ClaudeCodeProvider NDJSON / OpenCodeProvider SSE | per-CLI parser，7 类流式格式 | ✅ 机制一样 |
| per-CLI 抽象 | 主进程 `AgentProvider` + 渲染层 `AgentAdapter` | `AGENT_DEFS` + per-CLI eventParser | ✅ 思路一样 |
| **谁来做 spawn** | Electron 主进程，**按会话** spawn | 长驻 **daemon**（唯一特权进程） | ❌ 不同 |
| **UI↔spawn 桥** | Electron IPC（contextBridge） | HTTP `/api` + SSE | ❌ 不同 |
| **形态耦合** | desktop=本地 spawn，browser=远程 agent（两轴耦合） | daemon 同时服务 web/desktop，spawn 与 BYOK 都在 daemon（解耦） | ❌ 不同 |
| **CLI 广度** | 2 个（Claude Code、OpenCode），显式 | 16 个，PATH 自适应 | ❌ 不同 |
| **无 CLI 时的降级** | 远程后端 agent（P5，另一套代码，需 Kratos 配合）→ **v2 新增本地 BYOK**（同管线换 HTTP） | 本地 BYOK proxy，同事件管线换 HTTP 传输 | ❌ 不同（v2 后趋同） |
| **权限模型** | 交互式权限卡 + stdin 回写（Cursor/Cline 式）→ **v2 改为 P0 待验 + 兜底** | 文件夹沙箱 + bypassPermissions | ❌ 相反 |
| **工作目录** | 用户真实项目目录（dialog 选） | 一次性沙箱 `.od/projects/<id>/` | ❌ 不同 |
| **MCP 角色** | 无 | 额外提供只读 stdio MCP server（反向） | ❌ 不同 |
| **持久化** | useAgentStore + CLI 自带 `--resume` | SQLite + 文件工作区 | ❌ 不同 |
| **产品定位** | 真实代码库上的编码助手（Cursor/Cline 形态） | 设计物生成（Claude Design 形态） | ❌ 不同 |

---

## 3. 两种方式各自的优缺点

### 3.1 本方案（Electron 主进程 IPC + 交互式权限）

**优点**
- **最小新架构**：复用现有 `registerXxxIpc` 模式，零新增原生依赖（Node `child_process` + Electron `safeStorage`/`dialog`），P0 go/no-go 关卡设得对。
- **与 GoWind 深度贴合**：偏好面板、主题、i18n、路由权限码、Sider/Header 一致，UX 不割裂。
- **Kratos 后端保持干净**：agent 是纯桌面侧子系统，不污染业务后端。
- **安全模型文档化扎实**：`02-cli-connection.md` §8 八条（禁 shell、命令/参数白名单、cwd 沙箱、env 凭证、stderr 脱敏），比 open-design 在 spawn 路径上写得更细。

**缺点 / 风险**
- 🔴 **交互式权限是最大未验证假设**：`--permission-mode` 无「headless emit 请求等应答」语义，`-p` 遇需权限工具更可能直接拒绝。open-design 全线 bypass 是有力反面证据。→ v2 已提为 P0 go/no-go，见 `02` §4.1。
- 🟠 **(原)没有 BYOK 降级**：用户没装 CLI 功能即死。→ v2 已补「本地 BYOK SSE proxy」第三传输，见 `03` §1。
- 🟠 **形态耦合**：原把 desktop↔本地 CLI 与 browser↔远程 agent 绑死。→ v2 BYOK 传输部分解耦。
- 🟡 **2 个 CLI、无自适应**：加第 3 个要手写 Provider；open-design 证明 16 个也不比 2 个难多少。
- 🟡 **Windows argv**：原未提 CreateProcess 32KB 上限。→ v2 已补，见 `02` §4.3。
- 🟡 **事件泵绑生命周期**：原未防进程假死。→ v2 已补看门狗 + 强杀，见 `02` §7。

### 3.2 open-design（本地 daemon + 沙箱 bypass + BYOK）

**优点**
- **解耦最彻底**：spawn 与 BYOK 都在 daemon，吐同一事件流；web/desktop 同源；「传输」与「形态」正交。
- **16 CLI 自适应 + 一行加新 CLI**：广度与扩展性领先。
- **BYOK proxy 让「无 CLI」也能跑**：同管线换 HTTP，体验不降级。
- **沙箱 + bypass 的权限模型简单可靠**：不动真实代码、不与 CLI 交互权限流搏斗。
- **daemon 集中生命周期**：唯一特权进程，多客户端、长任务、headless/Docker 都能复用。
- **反向 MCP**：双向能力，agent 可反查 OD 文件。

**缺点**
- 🔴 **bypass 模型不适用于「改真实代码」**：一旦想在用户真实 GoWind 仓库 Read/Write/Bash，`bypassPermissions` 即裸奔——其安全前提「cwd 是一次性沙箱」换到真实代码库就塌。
- 🟠 **daemon = 攻击面**：监听端口的本地 web 服务 + HTTP proxy，SSRF/同源/凭证转发都得守（README 篇幅很大的 `OD_BIND_HOST`/`OD_ALLOWED_ORIGINS`/SSRF guard 说明这块很重）。
- 🟠 **复杂度高**：Node daemon + Express + SQLite + 16 parser + BYOK proxy + sidecar IPC + Electron 外壳，体量远超「主进程加一个 IPC」。对已有 Kratos 后端的项目，再起 Node daemon 是额外负担。
- 🟠 **进程生命周期坑已暴露**：[#3146](https://github.com/nexu-io/open-design/issues/3146) daemon-spawn 不退出→事件断流。
- 🟡 **SQLite 单写者**：dev-server 与 desktop 不能同时写同一 `.od/`，多实例靠 namespace 隔离。

---

## 4. 对本方案的 4 条修订（溯源）

| # | 修订 | 落盘位置 | 对标 open-design |
|---|------|---------|-----------------|
| 1 | 交互式权限标为「P0 未验证假设」+ 兜底（沙箱+bypass / 预批准白名单） | `02` §4.1、§8.6；`01` §4；`03` §1 P0 必验项、§2.4、§4 风险、§6 验收 #5 | 全线 `bypassPermissions` 回避交互权限 |
| 2 | 新增「本地 BYOK SSE proxy」第三传输，解耦传输与形态 | `03` §1、§3.1 `byok-proxy.ts`、§3.2 `LocalByokAdapter` | 同 daemon 既 spawn 又 BYOK，「same loop minus the spawn」 |
| 3 | Windows argv 长度防护（长 prompt 走 stdin / 临时文件） | `02` §4.3；`03` §2.4 | 每 adapter 的 `ENAMETOOLONG` 回退 |
| 4 | 事件泵与进程生命周期解耦（持续 pump + 看门狗 + SIGTERM 超时强杀 + Windows 进程树） | `02` §7 | issue #3146 的真实教训 |

---

## 5. 一句话总结

**机制一样（spawn + 解析 + 归一化），架构与取舍不同**：本方案是「Electron 壳内、按会话 spawn、面向真实代码库、交互式权限」的轻量编码助手；open-design 是「独立 daemon、沙箱 bypass、面向设计物产出、16 CLI + BYOK」的重型设计工坊。本方案不必照搬 open-design 的重型 daemon，但它的 4 个教训（权限、BYOK、argv、事件泵）已被吸收进 v2 修订。
