# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Layout

GoWind Admin (风行) is a full-stack enterprise admin scaffold — a **multi-package workspace with no root `package.json`**. One Go backend serves three interchangeable frontends that all consume the same backend API.

```
go-wind-admin/
├── backend/                    # Go + Kratos backend (Go module: go-wind-admin)
│   ├── api/                    # ← Protobuf source of truth + ALL generated code
│   │   ├── protos/             # .proto files (source — edit here)
│   │   └── gen/go/             # buf-generated Go (do not edit)
│   ├── app/admin/service/      # the admin service (single process, Kratos-based)
│   └── pkg/                    # shared Go packages (scripting, oss, eventbus, ...)
├── frontend/admin/
│   ├── react/                  # React 19 + Ant Design v6
│   ├── vue-element/            # Vue 3 + Element Plus
│   └── vue-vben/               # Vue 3 + Ant Design Vue + Vben (pnpm/turbo monorepo)
└── docs/
```

Each sub-project has its own detailed CLAUDE.md — **read the relevant one before editing in that area**:
- `backend/CLAUDE.md` — three-layer architecture, Ent/GORM repos, Wire DI, end-to-end "add a CRUD module" walkthrough
- `frontend/admin/react/CLAUDE.md`
- `frontend/admin/vue-element/CLAUDE.md`
- `frontend/admin/vue-vben/CLAUDE.md`

## The Contract-First Pipeline (the most important thing to understand)

A single Protobuf source drives code generation across the **entire stack**. The source of truth is `backend/api/protos/`. Never hand-edit any `generated/` or `gen/` directory.

```
backend/api/protos/**/*.proto
   │
   │  make api     (buf generate → buf.gen.yaml)
   ▼
backend/api/gen/go/             # Go server types, gRPC/HTTP services, errors, validators
   │
   │  make ts     (3 buf templates, plugin: protoc-gen-typescript-http)
   ▼
frontend/admin/{react, vue-element, vue-vben}/.../src/api/generated/   # TS service clients
```

**Proto two-layer architecture** (detailed in `backend/CLAUDE.md`):
- **Source domain layer** — `protos/<domain>/service/v1/` (e.g. `identity`, `permission`, `dict`, `task`, `audit`, `storage`, `internal_message`): defines messages + full gRPC service, **no** `google.api.http` annotations.
- **BFF layer** — `protos/admin/service/v1/`: defines the REST surface **with** `google.api.http` routes, imports domain messages without redefining them. **Frontends only consume this BFF layer** — every TS buf template restricts its input to `protos/admin/service/v1`.

So adding or changing an API means: edit proto → `make api` (regenerate Go) → `make ts` (regenerate all three TS clients) → implement backend Service/Repo → implement frontend hooks + page.

## Commands

### Backend — run from `backend/`

The `gow` CLI is the recommended wrapper (`go install github.com/tx7do/go-wind-toolkit/gowind/cmd/gow@latest`); Make targets are equivalent. `gow` commands can run from anywhere; Make runs from `backend/` (root-level targets recurse into services) or from `backend/app/admin/service/` (service-level targets).

| Task | gow | make (from `backend/`) |
|---|---|---|
| Run the admin service | `gow run admin` | `make run`¹ |
| Generate Go API code | `gow api` | `make api` |
| Generate Ent ORM | `gow ent` | `make ent` |
| Generate Wire DI | `gow wire` | `make wire` |
| Generate OpenAPI docs | — | `make openapi` |
| Generate **all 3 frontend TS clients** | — | `make ts` |
| Generate everything (ent+wire+api+openapi) | — | `make gen` |
| Tests | — | `make test` |
| Lint (golangci-lint) | — | `make lint` |
| Build binaries | — | `make build` |

¹ `make run` from `backend/app/admin/service/` runs `go run ./cmd/server -c ./configs` (regenerating api+openapi first).

Go 1.25+ (per `go.mod`). Backend HTTP server + Swagger UI on **http://localhost:7788** (`/docs/`).

### Frontends — run from each frontend dir

| Frontend | Dir | Dev | Build | Port |
|---|---|---|---|---|
| React | `frontend/admin/react` | `pnpm dev` | `pnpm build` | 7000 |
| Vue Element | `frontend/admin/vue-element` | `pnpm dev` | `pnpm build` | 3000 |
| Vue Vben | `frontend/admin/vue-vben` | `pnpm dev:antd` | `pnpm build:antd` | 5666 |

Install deps per-frontend with `pnpm install` (Vue Vben is itself a pnpm/turbo workspace; the app lives in `apps/admin`). All three enforce pnpm via a `preinstall` guard and use Conventional Commits (Husky + commitlint). Demo login: `admin` / `admin`.

### Local environment

Middleware (PostgreSQL, Redis, MinIO, Jaeger) runs in Docker; the app runs locally from the IDE:
- **Redis must be ≥ 8.0** — the backend uses `HExpire` / `HSETEX`, unavailable before 8.0.
- Middleware only (recommended for dev): `scripts/docker/libs_only.sh` (`.ps1` on Windows) or `make docker-libs`.
- Full Docker deploy (middleware + app): `scripts/docker/full_deploy.sh` or `make docker-up`.
- If connecting by hostname, map `postgres redis minio jaeger consul` → `127.0.0.1` in your hosts file.

## Shared Frontend Data-Layer Pattern

All three frontends implement the **same** data-layer architecture against the generated TS clients. Framework-specific templates and rules are in each sub-CLAUDE.md, but the shape is identical:

```
api/generated/   (buf-generated via make ts — never edit)
       ↓
api/client.ts    — apiClient singleton: a transport adapter wrapping axios,
                  exposing lazy-loaded service clients (apiClient.userService, …)
       ↓
api/hooks/ (react) · api/composables/ (vue)  — React Query / Vue Query integration
```

Conventions common to all three:
- **Components** call `useListXxx()` / `useGetXxx()` hooks; **non-component code** (stores, route guards, utilities) calls `fetchListXxx()` / `fetchXxx()` plain functions — never a hook outside a component.
- **List queries** always go through a `PaginationQuery` helper → `apiClient.xxxService.List(query.toRawParams())`.
- **Create** mutations wrap the body as `{ data: { ... } }` (gRPC convention).
- **Update** mutations build a field mask with `makeUpdateMask` so only changed fields are sent.
- **Query keys**: `['listXxx', query]` / `['getXxx', req]`.
- **Role codes and permission codes are stored separately** (`userRoles` vs `accessCodes`); route `meta.authority` is a mixed array of both. The superadmin role `*:*:*` bypasses all checks.

## Working in This Repo

- **CodeGraph is indexed** (`.codegraph/` exists at the root) — prefer `codegraph_explore` over grep/Read for locating symbols or tracing call paths across the layered architecture.
- Never hand-edit generated code: `backend/api/gen/go/`, `backend/app/.../internal/data/ent/`, any `*_wire_gen.go`, and every frontend `api/generated/` dir.
- House comment style is bilingual: `// 中文说明 / English description`.
- End-to-end feature flow: proto → `make api && make ts` → backend Service + Repo (register in the Wire provider sets, then `make wire`) → frontend hooks/composables + page + route + i18n.
