# 二、连接 CLI 方案

> 配套：`README.md`（总览）、`01-ui-layout.md`（界面）、`03-implementation-plan.md`（实施）。
>
> 本章是整个集成的技术核心：**如何在 Electron 桌面端把本地 Agent CLI 拉起来、把流式输出忠实回推到 React UI，并在渲染层用一套抽象屏蔽来源差异。**

---

## 1. 总体链路与分层

```
渲染进程                    主进程                      本机
─────────────────────────────────────────────────────────────────
window.desktop.agent            registerAgentIpc
   .start(cfg)        ──▶            │
                                  spawn CLI ────▶  claude / opencode
   .send(msg)         ──▶  写 stdin  │
                                  解析 stdout (NDJSON / SSE)
   .onEvent(cb)       ◀──  webContents.send('agent:event', 归一化事件)
   .answer(perm)      ──▶  写 stdin（权限响应）
   .stop()            ──▶  kill 进程  │
```

**三层职责**（与 README 架构图对应）：

| 层 | 位置 | 职责 | 屏蔽什么 |
|----|------|------|---------|
| `AgentAdapter`（渲染侧） | React | 对 UI 暴露统一异步流接口 | 来源：桌面 IPC / 远程 SSE |
| `registerAgentIpc` + `AgentProvider`（主进程侧） | Electron Main | spawn、协议解析、生命周期、安全 | 具体 CLI 协议（NDJSON vs SSE） |
| 归一化事件协议 | 跨层契约 | 统一事件 schema | CLI 私有字段 |

UI 与 `useXChat` **只消费归一化事件**，永远不直接触碰 CLI 私有协议。

---

## 2. 归一化事件协议（跨层契约）

主进程把不同 CLI 的输出归一成下列事件，经 `webContents.send('agent:event', { sessionId, ... })` 回推。渲染层 `AgentAdapter.onEvent` 透传，`useXChat`/`XStream` 消费。

```ts
/** 归一化事件类型 / Normalized agent event */
type AgentEvent =
  // 会话生命周期
  | { type: 'session.ready';       sessionId: string; provider: 'claude' | 'opencode'; cwd: string }
  | { type: 'session.end';         sessionId: string; reason: 'completed' | 'aborted' | 'error'; error?: string }
  // 文本流（增量）
  | { type: 'text.delta';          sessionId: string; messageId: string; delta: string }
  | { type: 'text.done';           sessionId: string; messageId: string }
  // 工具调用
  | { type: 'tool.start';          sessionId: string; toolId: string; name: string; input: unknown }
  | { type: 'tool.result';         sessionId: string; toolId: string; status: 'success' | 'error'; output: unknown }
  // 权限请求（交互式）
  | { type: 'permission.request';  sessionId: string; requestId: string; tool: string; summary: string }
  // 错误 / 日志
  | { type: 'error';               sessionId: string; message: string; recoverable: boolean }
  | { type: 'log';                 sessionId: string; level: 'info' | 'warn'; message: string };
```

- `messageId` / `toolId` / `requestId` 由主进程分配，UI 据此聚合增量（同一 `messageId` 的多次 `text.delta` 拼成一条气泡）。
- 该 schema **对 CLI 无关** —— Claude Code 和 OpenCode 都映射到它（见 §4、§5）。未来加新 CLI 只写主进程 Provider，不动 UI。

---

## 3. 渲染层：AgentAdapter 抽象

```ts
/** 渲染侧统一适配器接口 / Unified renderer-side adapter */
interface AgentAdapter {
  /** 启动会话：选 provider + 工作目录 + 权限模式 */
  startSession(cfg: { provider: AgentProvider; cwd: string; permissionMode?: PermissionMode }): Promise<string>;
  /** 发送一条用户消息（触发流式回复） */
  sendMessage(sessionId: string, text: string, attachments?: Attachment[]): void;
  /** 订阅归一化事件流（XStream 可直接消费此可迭代/订阅源） */
  onEvent(sessionId: string, handler: (e: AgentEvent) => void): () => void;  // 返回取消订阅
  /** 回答权限请求（交互式） */
  answerPermission(sessionId: string, requestId: string, decision: 'allow-once' | 'allow-session' | 'deny'): void;
  /** 中断当前生成 / 杀进程 */
  stop(sessionId: string): Promise<void>;
  /** 自检：本地是否安装了某 CLI（用于 UI 灰显） */
  probe(provider: AgentProvider): Promise<{ installed: boolean; version?: string }>;
}
```

两个实现（同一接口）：

| 实现 | 链路 | 形态 |
|------|------|------|
| `DesktopAgentAdapter` | 调 `window.desktop.agent.*`（preload 暴露） | 桌面端，**本期主交付** |
| `RemoteAgentAdapter` | 调后端 SSE `/api/agent/stream` | B/S 降级，P5 |

**工厂选择**：`const adapter = window.desktop?.isDesktop ? new DesktopAgentAdapter() : new RemoteAgentAdapter();`

### 3.1 接入 useXChat / XStream（@ant-design/x-sdk）

`useXChat` 的 `onRequest` 回调里调用 adapter，并把 `adapter.onEvent` 转成 `XStream` 可消费的异步迭代源（AsyncIterable），`useXChat` 自动把 `text.delta` 泵进当前气泡、把 `tool.*` 渲染为 `ThoughtChain`。

```ts
const [agent] = useXChat({
  requester: async (message, { messages, abortController }) => {
    const sid = await ensureSession();
    const stream = toAsyncIterable(adapter.onEvent, sid, abortController); // 包装为 AsyncIterable
    adapter.sendMessage(sid, message.content);
    return stream; // XStream/useXChat 逐 chunk 消费
  },
});
```

> 关键：渲染层不感知 `claude` / `opencode`，只感知归一化事件 —— 这是「换 CLI 不动 UI」的保证。

---

## 4. Claude Code 接入（默认 Provider）

### 4.1 CLI 调用方式

Claude Code 的 [headless 模式](https://docs.claude.com/en/docs/claude-code/headless) 以 NDJSON（每行一个 JSON）流式输出事件：

- **一次性提问**（无状态）：`claude -p "<prompt>" --output-format stream-json --verbose`
- **交互式多轮 + 权限**（推荐用于本场景）：
  ```
  claude -p --output-format stream-json --input-format stream-json --verbose \
         [--resume <session-id>] [--allowedTools <list>] [--permission-mode <mode>]
  ```
  - `--input-format stream-json` **必须**搭配 `--output-format stream-json`（不可与 text 输出混用）。
  - 通过 **stdin** 写入后续用户消息与权限响应；通过 **stdout** 读取 NDJSON。

> 调研来源：Claude Code headless 文档与社区实践（`--output-format stream-json`、`--print`、`--input-format stream-json`、`--sdk-url` 等 flag）。落地时以 `claude --help` 实测版本为准。

> 🔄 **v2 评审修订（对比 open-design）—— 交互式权限是最大未验证假设** ⚠️🔴
>
> 本方案 `01-ui-layout.md` §4 设计了「权限请求卡片 + `answerPermission` 写 stdin」的交互流，前提是 Claude Code 的 headless `stream-json` 模式会**发出可解析的权限请求事件并阻塞等 stdin 应答**。但 Claude Code 的 `--permission-mode` 枚举仅 `default` / `acceptEdits` / `bypassPermissions` / `plan`，**没有「headless 下 emit 权限请求事件并等 stdin 应答」的语义**；`-p` 非交互模式下遇到需权限的工具，更可能是**直接拒绝**（返回 `tool_result` 拒绝，模型自行决定下一步），而非弹卡等待。open-design 全线给 Claude Code 传 `--permission-mode bypassPermissions`、给 OpenCode 传 `--dangerously-skip-permissions`、Gemini `--yolo`、Copilot `--allow-all-tools` —— **正是集体回避了这个难点**，是有力的反面证据。
>
> **P0 必须实测证伪**（见 `03-implementation-plan.md` §2.4）。若证伪，二选一兜底：
> - **(a) 沙箱 + bypass（open-design 路线）**：agent 只在用户选定目录的**工作副本 / 临时目录**里跑 + 预批准/bypass，产出 patch → UI 展示 diff → 用户审查后「应用」回真实代码。真实代码不被 agent 直接写。
> - **(b) 预批准白名单 + 仅高危开闸**：spawn 时用 `--allowedTools` 固定只读/低危工具白名单 + `--permission-mode acceptEdits`；真正高危（`Bash` 任意命令、目录外写）以「会话级开关」在设置里显式开闸，而非「逐次 stdin 应答」。
>
> 当前 `01-ui-layout.md` §4 的交互卡 UX 应标注为「**依赖 P0 验证结果**」，未确认前不作为主链路承诺。

### 4.2 NDJSON 事件 → 归一化事件映射

| Claude Code NDJSON `type` | 归一化事件 |
|---------------------------|-----------|
| `init` | `session.ready`（取 session id / 模型信息） |
| `assistant` 文本增量 | `text.delta` / `text.done` |
| `assistant` 内 `tool_use` | `tool.start` |
| `user`（含 `tool_result`） | `tool.result` |
| 权限请求（交互流） | `permission.request` |
| `result` | `session.end`（reason: completed） |
| `api_retry` | `log`（warn） |
| 解析失败 / stderr | `error` |

### 4.3 进程模型

- `spawn('claude', args, { cwd, env, shell: false })` —— **必须 `shell:false`**，参数以数组传入，杜绝 shell 注入。
- 🔄 **v2 评审修订（对比 open-design）—— Windows argv 长度防护**：GoWind 桌面端 Windows 优先（NSIS 已验证），必须处理 `CreateProcess` ~32KB argv 上限。规则：**长 prompt 一律走 stdin**（`--input-format stream-json` 或 stdin 管道），`-p "<短 prompt>"` 仅用于一次性短提问；prompt 超过阈值（保守取 8KB，留足 flag/路径余量）即不再拼进 argv，改写 stdin；stdin 也溢出再退回**临时 prompt-file**（写 `%TEMP%` 后传路径）。open-design README 明确 `Windows ENAMETOOLONG fallbacks (stdin / prompt-file) on every adapter`，本方案须同样覆盖 ClaudeCodeProvider 与 OpenCodeProvider。
- **逐行读 stdout**（NDJSON 天然按行）：`child.stdout` pipe 到行缓冲，每完整一行 `JSON.parse` → 映射 → `webContents.send`。
- **stdin 写**：用户消息 / 权限响应序列化成 stream-json 输入帧写入 `child.stdin`。
- **stderr**：捕获为 `log`(warn) 或 `error`，不直接展示给用户原始堆栈（避免泄露路径/凭证）。

### 4.4 会话恢复

Claude Code 支持 `--resume <session-id>`。主进程把会话 id 持久化（随 `useAgentStore` 存本地），重启后可 `startSession({ resume })` 接续，UI 的 `Conversations` 直接显示历史。

---

## 5. OpenCode 接入（次选 Provider）

OpenCode 提供 [`opencode serve`](https://opencode.ai/docs/server/)：headless HTTP 服务器 + OpenAPI + **SSE 事件流**（session messages）。相比 Claude Code 的「每次 spawn 一个进程」，OpenCode 更适合「常驻一个 server，多次收发」。

### 5.1 接入方式

1. 主进程探测：若本机无运行中的 `opencode serve`，则 `spawn('opencode', ['serve', '--port', '<n>'], { shell: false })` 拉起一个（或连用户已启动的实例）。
2. 通过 HTTP REST 创建 session、发送 message。
3. 订阅 SSE 流拿事件，归一化映射：

| OpenCode SSE 事件 | 归一化事件 |
|-------------------|-----------|
| session 创建 | `session.ready` |
| assistant 文本片段 | `text.delta` / `text.done` |
| tool 开始 / 结束 | `tool.start` / `tool.result` |
| error | `error` |

4. 优雅停机：退出时 kill 拉起的 server（若是本方案启动的），不动用户自有实例。

> OpenCode 的 REST/SDK 细节以 [`opencode.ai/docs/server`](https://opencode.ai/docs/server/) 实测为准（社区反馈部分 SDK 文档滞后，落地时直连 HTTP + SSE 最稳）。

---

## 6. 主进程：AgentProvider 抽象

```ts
/** 主进程侧 Provider 接口（每个 CLI 一个实现）/ Main-process provider */
interface AgentProvider {
  readonly id: 'claude' | 'opencode';
  probe(): Promise<{ installed: boolean; version?: string }>;     // which/版本探测
  start(session: SessionContext): Promise<ChildProcess | Handle>; // 拉起/连接
  sendInput(handle, frame: unknown): void;                        // 写 stdin / POST
  normalize(rawEvent: unknown, session): AgentEvent | AgentEvent[]; // 协议归一化
  stop(handle): Promise<void>;                                    // 终止
}
```

- `ClaudeCodeProvider`：spawn + NDJSON 解析 + stdin 写入。
- `OpenCodeProvider`：serve 拉起 + HTTP + SSE。
- `registerAgentIpc`：`ipcMain.handle('agent:start'|'agent:send'|'agent:answer'|'agent:stop'|'agent:probe')`，内部按 `provider` 字段分发到对应 `AgentProvider`，输出统一经 `event.sender.send('agent:event', ...)` 回推。

> 落地文件：`frontend/admin/desktop/electron/main/ipc/agent.ts`（新增），并在 `electron/main/ipc/index.ts` 的 `registerAllIpc()` 注册 —— 与现有 `registerDesktopIpc`/`registerFileIpc`/`registerPrintIpc` 完全同构。

---

## 7. 进程生命周期与异常

| 场景 | 处理 |
|------|------|
| 用户中断 | `agent:stop` → `child.kill('SIGTERM')`（Claude）/ 停 SSE（OpenCode）→ `session.end(reason:'aborted')` |
| 任务正常结束 | 收到 `result`/SSE end → `session.end(reason:'completed')` |
| 进程崩溃 / 非零退出 | `session.end(reason:'error', error)` + stderr 摘要 |
| 超时 | 单次生成可配超时，超时自动 abort（配置项，默认关） |
| 窗口关闭 / 应用退出 | `app.before-quit` 时遍历活跃 session 全部 kill，**杜绝孤儿 CLI 进程** |
| stdin 管道断裂 | 捕获 `EPIPE`，转 `error` 事件，提示重连 |
| 探测失败 | UI 入口灰显 + tooltip「未检测到 claude，请先安装」 |

**孤儿进程防护**是桌面端最易被忽视的坑：必须在 `app.on('before-quit')` 与窗口关闭钩子里统一清理本方案拉起的子进程，并记录到 `docs/caozuo/desktop-optional-config.md` 的能力清单。

> 🔄 **v2 评审修订（吸取 open-design [issue #3146](https://github.com/nexu-io/open-design/issues/3146)）—— 事件泵与进程生命周期解耦** ⚠️
>
> open-design 出过真实 bug：daemon spawn 的 claude **跑完 API 调用却不退出**，导致 stdout→parser 的泵（`claude.feed`）从未触发，UI 收不到任何事件。教训：
> - **stdout 持续 pump**：行缓冲器「每读到完整一行」就立即 `JSON.parse`→`send`，**不要把事件发射绑在「进程退出 / 收到 `result`」上**；`result` 是「逻辑完成」信号，不是「开始吐事件」的前提。
> - **心跳 / 看门狗**：spawn 后启动 N 秒（如 30s）静默看门狗，超时无 stdout 且无 `result` → 主动 `SIGTERM` + `session.end(reason:'error')`，避免进程假死挂住 UI。
> - **SIGTERM 超时强杀**：`kill('SIGTERM')` 后设 3–5s 二次计时，未退出则 `kill('SIGKILL')`；启动时把每个子进程句柄登记进活跃表，`before-quit` 遍历强杀兜底。
> - **Windows 无 SIGTERM**：`kill()` 在 Windows 等价强制终止，需用「Job Object / `taskkill /T /PID`」关联子进程树，确保 claude 拉起的孙进程（如 Bash 工具起的进程）一并清理。

---

## 8. 安全模型（重点）

在 GUI 里 spawn 本地 Agent CLI，等于给「网页」开了执行本地代码的口子，安全边界必须清晰：

1. **禁止 shell**：所有 spawn 用 `shell: false` + 参数数组，禁止 `child_process.exec` / `execSync`。用户输入绝不拼进命令字符串。
2. **命令白名单**：可执行名固定枚举（`claude` / `opencode`），路径优先走 `which` 解析或用户在设置里显式配置的可执行路径，不接受任意可执行文件。
3. **参数白名单**：传给 CLI 的 flag 固定集合（见 §4.1 / §5.1），运行时校验，拒绝未知 flag。
4. **工作目录沙箱**：`cwd` 来自用户经**文件对话框**选择的项目目录（复用桌面端 `dialog` 能力），不取渲染进程传入的任意字符串直接当 cwd。
5. **凭证不落命令行**：API key / token 经 `env` 注入子进程，**绝不**进 `args`（进程列表 `ps` 可见命令行）。桌面端用 Electron `safeStorage`（OS keychain）存取，与 `desktop-client-plan.md` 第九章安全基线一致。
6. **权限最小化**：写操作不静默放行。⚠️ 交互式「逐次询问」依赖 P0 验证（见 §4.1 v2 评审修订）；未验证前默认走「预批准只读白名单 + 高危会话级开闸」；`bypassPermissions` **只允许在用户显式选定的一次性沙箱目录**内启用，**绝不**用于真实代码库。
7. **stderr 脱敏**：展示前过滤疑似凭证/绝对家目录路径，原始日志默认不外发。
8. **B/S 桥接（P5）若实现**：本地 bridge server 必须 `127.0.0.1` only + 鉴权 + 同源，严禁对外网监听。

---

## 9. 数据流与状态管理

- **`useAgentStore`（Zustand，新增）**：`sessions[]`、`activeSessionId`、`messagesBySession`、`connection: 'idle'|'connecting'|'streaming'|'error'`、`activeProvider`。**持久化**（localStorage 或经桌面端文件 I/O 存大体积会话）。
- **流式态**用 `useXChat` 管理（组件内），一条消息生成完成后镜像写入 `useAgentStore` 做持久化与跨视图共享。
- **与现有数据层的关系**：Agent 不走 `apiClient`/React Query（那是远端业务 API），是独立的本地能力子系统；但 `Conversations` 列表的加载/删除可复用 React Query 的缓存模式（query key `['agent','sessions']`）。
- **会话存储位置**：本地 Agent 的会话本质属于「本机」，优先委托 CLI 自身会话（Claude `--resume` / OpenCode session id），UI 只存「会话索引 + 元数据」，避免重复持久化大段消息。

---

## 10. 预加载桥（preload）扩展

在现有 `frontend/admin/desktop/electron/preload/index.ts` 的 `desktop` 对象上新增 `agent` 子能力（受 `contextBridge` 白名单约束，绝不暴露 `ipcRenderer` 原始对象）：

```ts
const desktop = {
  // …既有能力
  agent: {
    probe: (p) => ipcRenderer.invoke('agent:probe', p),
    start: (cfg) => ipcRenderer.invoke('agent:start', cfg),
    send: (sid, text, atts) => ipcRenderer.send('agent:send', { sid, text, atts }),
    answer: (sid, rid, dec) => ipcRenderer.send('agent:answer', { sid, rid, dec }),
    stop: (sid) => ipcRenderer.invoke('agent:stop', sid),
    onEvent: (cb) => {
      const h = (_e, payload) => cb(payload);
      ipcRenderer.on('agent:event', h);
      return () => ipcRenderer.removeListener('agent:event', h);
    },
  },
};
```

`react/types/desktop.d.ts` 同步补 `DesktopBridge.agent` 类型契约（与既有约定一致）。前端用 `window.desktop?.agent?.onEvent(...)` 可选链，B/S 自动降级。
