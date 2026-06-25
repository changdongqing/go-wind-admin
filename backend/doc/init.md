# 本地后端开发环境说明

> 检查日期：2026-06-25　机器：Windows 11 / GOOS=windows amd64
> 本文档记录本地后端开发环境的现状、所需配置，以及本机与项目默认配置之间的差异。

## 一句话结论

**代码生成工具链与运行时已全部就绪可用；PostgreSQL 18 / Redis 8.8.0 已在本机以原生服务运行。** 剩余工作只是统一本机中间件与项目配置之间的差异（Redis 密码、PG 凭据/库名、连接主机名），处理完即可 `gow run admin` 启动。

---

## 一、环境总览

图例：✅ 已就绪　⚠️ 部分/待确认　➖ 未装（可选）

### 1. 运行时与基础工具

| 工具 | 要求 | 本机 | 说明 |
|---|---|---|---|
| Go | ≥ 1.25（go.mod=1.25.7） | ✅ 1.26.4 | 满足 |
| Git | — | ✅ 2.54.0 | — |
| Docker | ≥ 20.0 | ✅ 29.5.3 | — |
| Docker Compose | ≥ v2 | ✅ v5.1.4 | — |
| Node.js | ≥ 20.10 | ✅ 24.17.0 | 仅前端 / `make ts` 生成 TS 时用 |
| pnpm | ≥ 10 | ✅ 11.8.0 | 仅前端 |
| make | ≥ 3.8 | ➖ 未装 | 可用 `gow` 替代；`make ts` 可手动 `buf generate` |
| protoc（独立二进制） | — | ➖ 不需要 | `buf` 内置编译，项目只用插件 |

### 2. 代码生成工具链（已全部安装，PATH 已就绪）

> 安装位置 `C:\Users\chdq0306\go\bin`，该目录已在 Windows 用户 PATH 与系统 PATH 中，新开终端可直接调用 `buf` / `gow` / `protoc-gen-*`。

| 工具 | 版本 | 用途 |
|---|---|---|
| buf | 1.71.0 | 代码生成总入口（`make api` / `make ts`） |
| gow | v0.1.0 | 项目 CLI wrapper（api / ent / wire / run） |
| protoc-gen-go | latest | Go message 类型 |
| protoc-gen-go-grpc | latest | gRPC service |
| protoc-gen-go-http | latest | Kratos REST service |
| protoc-gen-go-errors | latest | Kratos 错误码 |
| protoc-gen-openapi | latest | OpenAPI / Swagger |
| protoc-gen-validate | latest | 参数校验（PGV） |
| protoc-gen-go-redact | latest | 日志脱敏 |
| protoc-gen-typescript-http | latest | 前端 TS 客户端（`make ts`） |

> 安装说明：官方代理 `proxy.golang.org` 本机无法直连（连接超时），实际使用国内代理 `GOPROXY=https://goproxy.cn,direct` 安装成功。本次为**命令级临时代理，未修改全局 `go env`**。今后任何 `go install` / `go mod download` 建议同样带上此代理；goproxy.cn 的 sumdb 偶发 `504`，重试即可。

### 3. 运行中间件（本机原生服务，非 Docker）

| 中间件 | 要求 | 本机现状 | 说明 |
|---|---|---|---|
| PostgreSQL | ≥ 14 | ✅ **18**，服务 `postgresql-x64-18`，监听 `0.0.0.0:5432` | 版本与端口均满足 |
| Redis | ≥ 8.0 | ✅ **8.8.0**，服务 `Redis`，监听 `127.0.0.1:6379`，standalone | 版本与端口满足；**当前未设密码** |
| MinIO | ≥ RELEASE.2024 | ⚠️ 1Panel 管理的实例在跑 | 端口/凭据待确认是否匹配项目 |
| Jaeger | ≥ 1.40 | ➖ 未运行 | 链路追踪，本地开发可选 |

---

## 二、本机中间件 vs 项目配置（差异表）

项目连接参数来自 `app/admin/service/configs/*.yaml`：

| 项 | 项目配置默认值 | 本机实际 | 是否匹配 / 处理建议 |
|---|---|---|---|
| PG host | `postgres`（容器名） | `127.0.0.1`（本机） | ❌ 需 hosts 映射，或改配置为 `localhost` |
| PG port | 5432 | 5432 | ✅ |
| PG user | `postgres` | `postgres`（PG18 默认） | ✅（密码待确认） |
| PG password | `*Abcd123456` | **未知**（PG18 安装时设定） | ❓ 需确认/统一 |
| PG dbname | `gwa` | **库是否存在未知** | ❓ 需先 `CREATE DATABASE gwa;`（`migrate:true` 只建表不建库） |
| Redis host | `redis`（容器名） | `127.0.0.1` | ❌ 同上，hosts 或改 `localhost` |
| Redis port | 6379 | 6379 | ✅ |
| Redis password | `*Abcd123456` | **无密码** | ❌ 需统一 |
| MinIO endpoint | `minio:9000` | 1Panel 实例（端口未知） | ❌ 需确认端口/凭据，或另起容器 |
| MinIO 凭据 | `root` / `*Abcd123456` | 1Panel 自定 | ❓ 待确认 |
| 服务端口 | REST `:7788`、SSE `:7789` | 未占用 | ✅ |
| authz | `noop` | — | ✅ 开发期无需 Casbin/OPA |

---

## 三、待处理配置项（应用启动前需统一）

1. **Redis 密码**：本机无密码，配置写了 `*Abcd123456`。二选一：
   - 给本机 Redis 设密码：改 `redis.windows.conf` 的 `requirepass *Abcd123456` 后重启服务；**或**
   - 把 `configs/data.yaml` 与 `server.yaml`（Asynq URI）里 Redis 的 `password` 置空。
2. **PG 凭据与库**：确认本机 `postgres` 用户密码是否为 `*Abcd123456`；并预先创建数据库 `gwa`（`CREATE DATABASE gwa;`）——`migrate:true` 只在库内建表，不创建库本身。
3. **连接主机名**：项目配置用的是容器名 `postgres` / `redis` / `minio`，本机服务在 `127.0.0.1`。二选一：
   - 在 `C:\Windows\System32\drivers\etc\hosts` 追加 `127.0.0.1 postgres redis minio`；**或**
   - 把配置里的 `host=postgres`→`localhost`、`addr=redis:6379`→`localhost:6379`、`endpoint=minio:9000`→`localhost:9000`。
   > 注意：本机 Redis 仅监听 `127.0.0.1`（非 `0.0.0.0`），应用必须在本机运行才能连上。
4. **MinIO**：确认 1Panel 实例的端口与 `root` / `*Abcd123456` 凭据是否可用；否则用 `scripts/docker/libs_only` 另起一组仅供本项目使用的容器，避免与 1Panel 实例混用。

---

## 四、可选工具（仍未安装，按需）

| 工具 | 用途 | 何时需要 |
|---|---|---|
| golangci-lint | `make lint` | 想跑静态检查时 |
| make | make 命令 | 习惯用 make 时（`gow` + 手动 `buf generate` 可替代） |
| kratos / ent / gnostic CLI | 各自代码生成 | 已可由 `gow` 子命令替代，一般不需要 |
| pm2 | 物理机进程托管 | 生产部署，开发无需 |
| Jaeger | 链路追踪 UI | 本地开发可选 |

---

## 五、启动验证路径

工具链就绪、第三节差异处理完后：

```bash
cd backend
gow run admin        # 或：cd app/admin/service && go run ./cmd/server -c ./configs
```

访问 http://localhost:7788/docs 验证 Swagger UI。
