# GoWind Admin｜风行，开箱即用的企业级前后端一体中后台脚手架

> **让中后台开发如风般自由 — GoWind Admin**

风行（GoWind Admin）是一款开箱即用的企业级Golang全栈后台管理系统的脚手架。

系统后端基于GO微服务框架[go-kratos](https://go-kratos.dev/)，前端有Vue3和React两个版本，兼顾微服务的扩展性与单体部署的便捷性。

尽管依托微服务框架设计，但系统前后端均支持单体架构模式开发与部署，灵活适配不同团队规模及项目复杂度需求，平衡灵活性与易用性。

产品具备上手简易、功能完备的核心优势，依托风行对企业级场景的深度适配能力，可助力开发者快速落地各类企业级管理系统项目，大幅提升开发效率。

[English](./README.en-US.md) | **中文** | [日本語](./README.ja-JP.md)

## 演示地址

> 演示地址入口：<https://demo.admin.gowind.cloud>
> 
> Vue3 Vben 演示地址：<https://vben.admin.gowind.cloud>  
> Vue3 Element Plus 演示地址：<https://ele.admin.gowind.cloud>  
> React 演示地址：<https://react.admin.gowind.cloud>
>
> 后端Swagger地址：<https://api.demo.admin.gowind.cloud/docs/>
>
> 默认账号密码: `admin` / `admin`

## 风行·核心技术栈

秉持高效、稳定、可扩展的技术选型理念，系统核心技术栈如下：

- **后端**：`Golang`、`go-kratos`、`Wire`、`Ent ORM` / `Gorm`、`MySQL`、`Redis`、`Docker`
- **公共基础能力**：`JWT 鉴权`、`Casbin` /`OPA` / `Zanzibar` 权限控制、`SSE 消息推送`、`Swagger 接口文档`
- **Vue Vben 版**：`Vue3` + `TypeScript` + `Vite` + `Ant Design Vue` + `Vben Admin`
- **Vue Element 版**：`Vue3` + `TypeScript` + `Vite` + `Element Plus`（轻量纯净版）
- **React 版**：`React19` + `TypeScript` + `Vite` + `React Router` + `Zustand` + `Ant Design V6` + `@ant-design/pro-components`（**无 UMI 框架**）

## 风行·快速上手指南

### 环境脚本选型

- Linux / macOS 开发环境：`scripts/env/install_unix_dev.sh`
- Linux / macOS 生产环境：`scripts/env/install_unix_prod.sh`
- Windows 开发环境：`scripts/env/install_windows_dev.ps1`

### Docker 两种部署模式

- **full_deploy 完整模式**：同步启动中间件+后端应用，适用于一键演示、生产部署；
- **libs_only 依赖模式（推荐）**：仅启动中间件，应用本地IDE运行调试，适配日常开发。

### 后端启动命令

#### Linux / macOS

```shell
# 赋予脚本执行权限
chmod +x scripts/**/*.sh

# 开发环境（推荐）
./scripts/env/install_unix_dev.sh
./scripts/docker/libs_only.sh
gow run admin

# 生产环境
./scripts/env/install_unix_prod.sh
./scripts/docker/full_deploy.sh

# PM2 进程托管（生产进阶）
./scripts/deploy/pm2_service.sh
```

#### Windows（PowerShell 管理员）

```powershell
# 放行脚本策略（首次仅需执行一次）
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser

# 初始化环境
.\scripts\env\install_windows_dev.ps1

# 本地开发
.\scripts\docker\libs_only.ps1
gow run admin

# 一键完整部署
.\scripts\docker\full_deploy.ps1
```

### 前端启动说明

前端统一存放于 `frontend/admin` 目录，依赖安装命令统一，启动命令差异化配置：

- React：目录 `frontend/admin/react`，启动命令 `pnpm dev`，本地端口：`7000`
- Vue Element：目录 `frontend/admin/vue-element`，启动命令 `pnpm dev`，本地端口：`3000`
- Vue Vben：目录 `frontend/admin/vue-vben`，启动命令 `pnpm dev:antd`，本地端口：`5666`

```shell
# 安装依赖
pnpm install

# React版本
cd frontend/admin/react
pnpm dev

# Vue3 Element版本
cd frontend/admin/vue-element
pnpm dev

# Vue3 Vben版本
cd frontend/admin/vue-vben
pnpm dev:antd
```

## 风行·核心功能列表

| 功能   | 说明                                                                       |
|------|--------------------------------------------------------------------------|
| 用户管理 | 管理和查询用户，支持高级查询和按部门联动用户，用户可禁用/启用、设置/取消主管、重置密码、配置多角色、多部门和上级主管、一键登录指定用户等功能。 |
| 租户管理 | 管理租户，新增租户后自动初始化租户部门、默认角色和管理员。支持配置套餐、禁用/启用、一键登录租户管理员功能。                   |
| 角色管理 | 管理角色和角色分组，支持按角色联动用户，设置菜单和数据权限，批量添加和移除员工。                                 |
| 权限管理 | 管理权限分组、菜单、权限点，支持树形列表展示。                                                  |
| 组织管理 | 管理组织，支持树形列表展示。                                                           |
| 部门管理 | 管理部门，支持树形列表展示。                                                           |
| 职位管理 | 用户职务管理，职务可作为用户的一个标签。                                                           |
| 接口管理 | 管理接口，支持接口同步功能，主要用于新增权限点时选择接口，支持树形列表展示、操作日志请求参数和响应结果配置。                   |
| 菜单管理 | 配置系统菜单，操作权限，按钮权限标识等，包括目录、菜单、按钮。                                                                  |
| 字典管理 | 管理数据字典大类及其小类，支持按字典大类联动字典小类、服务端多列排序、数据导入和导出。                              |
| 任务调度 | 管理和查看任务及其任务运行日志，支持任务新增、修改、删除、启动、暂停、立即执行。                                 |
| 文件管理 | 管理文件上传，支持文件查询、上传到OSS或本地、下载、复制文件地址、删除文件、图片支持查看大图功能。                       |
| 消息分类 | 管理消息分类，支持2级自定义消息分类，用于消息管理消息分类选择。                                         |
| 消息管理 | 管理消息，支持发送指定用户消息，可查看用户是否已读和已读时间。                                          |
| 站内信  | 站内消息管理，支持消息详细查看、删除、标为已读、全部已读功能。                                          |
| 个人中心 | 个人信息展示和修改，查看最后登录信息，密码修改等功能。                                              |
| 缓存管理 | 缓存列表查询，支持根据缓存键清除缓存。                                                      |
| 登录日志 | 登录日志列表查询，记录用户登录成功和失败日志，支持IP归属地记录。                                        |
| 操作日志 | 操作日志列表查询，记录用户操作正常和异常日志，支持IP归属地记录，查看操作日志详情。                               |

## 风行·后台截图展示

<table>
    <tr>
        <td><img src="./docs/images/admin_login_page.png" alt="后台用户登录界面"/></td>
        <td><img src="./docs/images/admin_dashboard.png" alt="后台分析界面"/></td>
    </tr>
    <tr>
        <td><img src="./docs/images/admin_user_list.png" alt="后台用户列表界面"/></td>
        <td><img src="./docs/images/admin_user_create.png" alt="后台创建用户界面"/></td>
    </tr>
    <tr>
        <td><img src="./docs/images/admin_tenant_list.png" alt="后台租户列表界面"/></td>
        <td><img src="./docs/images/admin_tenant_create.png" alt="后台创建租户界面"/></td>
    </tr>
    <tr>
        <td><img src="./docs/images/admin_org_unit_list.png" alt="组织单位列表界面"/></td>
        <td><img src="./docs/images/admin_org_unit_create.png" alt="创建组织单位界面"/></td>
    </tr>
    <tr>
        <td><img src="./docs/images/admin_position_list.png" alt="后台职位列表界面"/></td>
        <td><img src="./docs/images/admin_position_create.png" alt="后台创建职位界面"/></td>
    </tr>
    <tr>
        <td><img src="./docs/images/admin_role_list.png" alt="后台角色列表界面"/></td>
        <td><img src="./docs/images/admin_role_create.png" alt="后台创建角色界面"/></td>
    </tr>
    <tr>
        <td><img src="./docs/images/admin_permission_list.png" alt="后台权限列表界面"/></td>
        <td><img src="./docs/images/admin_permission_create.png" alt="后台创建权限界面"/></td>
    </tr>
    <tr>
        <td><img src="./docs/images/admin_menu_list.png" alt="后台目录列表界面"/></td>
        <td><img src="./docs/images/admin_menu_create.png" alt="后台创建目录界面"/></td>
    </tr>
    <tr>
        <td><img src="./docs/images/admin_task_list.png" alt="后台调度任务列表界面"/></td>
        <td><img src="./docs/images/admin_task_create.png" alt="后台创建调度任务界面"/></td>
    </tr>
    <tr>
        <td><img src="./docs/images/admin_dict_list.png" alt="后台数据字典列表界面"/></td>
        <td><img src="./docs/images/admin_dict_entry_create.png" alt="后台创建数据字典条目界面"/></td>
    </tr>
    <tr>
        <td><img src="./docs/images/admin_internal_message_list.png" alt="后台站内信消息列表界面"/></td>
        <td><img src="./docs/images/admin_internal_message_publish.png" alt="后台发布站内信消息界面"/></td>
    </tr>
    <tr>
        <td><img src="./docs/images/admin_login_policy_list.png" alt="登录策略列表界面"/></td>
        <td><img src="./docs/images/admin_login_policy_create.png" alt="登录策略创建界面"/></td>
    </tr>
    <tr>
        <td><img src="./docs/images/admin_login_audit_log_list.png" alt="后台登录日志界面"/></td>
        <td><img src="./docs/images/admin_api_audit_log_list.png" alt="后台操作日志界面"/></td>
    </tr>
    <tr>
        <td><img src="./docs/images/admin_api_list.png" alt="API列表界面"/></td>
        <td><img src="./docs/images/api_swagger_ui.png" alt="后端内置Swagger UI界面"/></td>
    </tr>
</table>

## 联系我们

- 微信个人号：`yang_lin_bo`（备注：`go-wind-admin`）
- 掘金专栏：[go-wind-admin](https://juejin.cn/column/7541283508041826367)

## [感谢JetBrains提供的免费GoLand & WebStorm](https://jb.gg/OpenSource)

[![avatar](https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.svg)](https://jb.gg/OpenSource)
