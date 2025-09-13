# DHGoHttp Roadmap / 增强规划

> 本文档用于跟踪未来特性、改进方向与技术设计要点。当前版本: `0.3.0-dev`
>
> Legend / 图例: ✅ 已实现 | 🟡 进行中 | 🔜 计划 | 💡 想法 / 待评估 | ❗ 需决策 | ⏳ 依赖外部
 
## 1. 愿景 (Vision)

提供一个「开箱即用 + 可按需加装安全与可 observability 模块」的静态/轻量资源分发工具：单文件、最少配置、适合临时/内网/调试/CI 场景，同时保持清晰扩展路径。

## 2. 当前已实现 (Baseline 0.3.x)

- 端口自动递增扫描 (`-port` + `-max-port-scan`)
- 自定义根目录 (`-dir`)
- Windows UAC 自动提升 + 防火墙规则幂等 & 退出清理
- 简单 Token 共享密钥 (`-token`)
- 只读目录禁止列目录 (`-readonly`)
- 访问日志（IP/方法/路径/状态/字节/耗时）
- 优雅退出 (Ctrl+C / SIGTERM) + 清理本进程创建的防火墙规则
- 自定义绑定地址 (`-bind`)
- 版本输出 (`-version`) 与自定义 `--help`
- 多平台构建矩阵 (GitHub Actions)

## 3. 里程碑分段 (Milestones)

| 里程碑 | 目标 | 关键输出 | 状态 |
|--------|------|----------|------|
| 0.4.0 | 基础可观测与下载安全增强 | JSON 日志、可选请求 ID、简单统计 | 🔜 |
| 0.5.0 | 访问控制细粒度 | 路径白/黑名单、全局只读+细粒度例外 | 🔜 |
| 0.6.0 | 性能与并发 | 可配置下载限速、并发连接上限、压缩 | 💡 |
| 0.7.0 | 可观测生态集成 | Prometheus 指标、健康检查 | 💡 |
| 0.8.0 | 交付体验 | Web UI 浏览、搜索、打包下载 | 💡 |
| 1.0.0 | 稳定发布 | 兼容性冻结、文档完整、测试覆盖 | 💡 |

## 4. Backlog 分类 (Categorized Backlog)

### 4.1 安全 / 访问控制

| 项目 | 描述 | 优先级 | 状态 | 备注 |
|------|------|--------|------|------|
| Path Allow/Deny 列表 | 配置文件或 flag 指定允许/拒绝路径前缀/模式 | 高 | 🔜 | 支持 glob `**/*.log` |
| 细粒度只读 | 目录只读 + 指定路径可下载 | 中 | 💡 | 与 deny/allow 组合 |
| 目录浏览开关 UI | 可启用带样式目录列表 | 中 | 💡 | 默认仍关闭 |
| 可选 basic-auth | 简单用户/密码 | 中 | 💡 | 避免重复造轮可选 htpasswd 解析 |
| JWT 校验 (可选) | 支持 HS256/RS256 验签 | 低 | 💡 | 需引入第三方库 |

### 4.2 日志 / 审计 / 可观测

| 项目 | 描述 | 优先级 | 状态 | 备注 |
|------|------|--------|------|------|
| JSON 日志模式 | `-log-json` 输出结构化 | 高 | 🔜 | 便于收集与分析 |
| 请求 ID 注入 | `X-Request-ID` 传入/生成 | 中 | 🔜 | 日志链路 |
| 下载计数器 | 每文件累计次数 | 中 | 💡 | 可导出 /metrics |
| Prometheus 指标 | qps、耗时直方图、字节计数 | 中 | 💡 | `-metrics` 监听独立端口 |
| 健康检查 | `/healthz` 简单 200 | 低 | 💡 | Kubernetes 就绪探针 |

### 4.3 传输 / 性能

| 项目 | 描述 | 优先级 | 状态 | 备注 |
|------|------|--------|------|------|
| Gzip/Br 压缩 | 针对文本类型响应 | 中 | 💡 | 避免二进制浪费 CPU |
| 限速功能 | `-rate` 全局或/IP 限速 | 中 | 💡 | token bucket |
| 并发连接上限 | `-max-conns` 拒绝过载 | 中 | 💡 | 保护主机 |
| Sendfile 优化 | 大文件使用 `http.ServeContent` / `io.Copy` 调优 | 低 | 💡 | Benchmark 验证 |

### 4.4 功能体验

| 项目 | 描述 | 优先级 | 状态 | 备注 |
|------|------|--------|------|------|
| Web UI | 简单文件浏览+下载按钮 | 中 | 💡 | 纯静态模板 |
| 递归打包下载 | `?zip=1` 打包目录 | 中 | 💡 | zip 流式输出 |
| 单文件 SHA256 | `?sha256=1` 返回哈希 | 低 | 💡 | 结合下载计数 |
| 多文件选择下载 | 批量打包 | 低 | 💡 | UI 支持 |

### 4.5 部署 / 交付

| 项目 | 描述 | 优先级 | 状态 | 备注 |
|------|------|--------|------|------|
| Release 自动产物校验 | 生成 SHA256SUMS / 签名 | 中 | 🔜 | GitHub Actions job |
| Homebrew 公式 | macOS 安装 | 低 | 💡 | tap 仓库 + goreleaser 可选 |
| Scoop manifest | Windows 安装 | 低 | 💡 | buckets 提交 |
| Docker 镜像 | 提供 `docker run` 快速共享 | 中 | 💡 | 多架构 buildx |

### 4.6 质量保障

| 项目 | 描述 | 优先级 | 状态 | 备注 |
|------|------|--------|------|------|
| 单元测试覆盖 | 端口探测/中间件 | 高 | 🔜 | 首批核心逻辑 |
| 集成测试 | 启动+请求真实文件 | 中 | 💡 | 临时端口+临时目录 |
| Benchmark | 大文件传输性能基准 | 低 | 💡 | go test -bench |
| 静态分析扩展 | govulncheck 集成 | 中 | 💡 | CI 增加步骤 |

## 5. 设计草案 / 实现要点 (Draft Notes)

### 5.1 JSON 日志

- 新 flag: `-log-json`
- 结构: `{ts, level, msg, method, path, status, bytes, dur_ms, ip, req_id}`
- 复用当前 loggingMiddleware，抽象一个 `logger` 接口

### 5.2 路径 Allow/Deny

- Flag 或配置文件：`-allow "prefix:/foo" -deny "glob:**/*.secret"`
- 解析阶段构建匹配器：前缀 / 后缀 / 通配 glob
- 中间件在进入 FileServer 前拒绝 (403)

### 5.3 下载计数器 & Prometheus

- 计数 map[string]*stats
- 原子自增；暴露 `/metrics` 或独立监听 (避免混合安全策略)
- 指标：`http_download_total{file=""}`, `http_bytes_total`, `request_duration_seconds` histogram

### 5.4 限速

- 全局 token bucket：burst = 2*rate，循环 goroutine 补充
- 每次写前 `Wait(n)`，或分块对齐 32KB
- 对大文件影响：可配置 chunk 大小

### 5.5 打包下载

- 遍历目录构建 zip.Writer 流式写入
- 安全：路径清理防止目录穿越
- 可限制最大总大小 (防内存/时间消耗) `-zip-max-mb`

## 6. 风险与注意事项 (Risks)

| 风险 | 说明 | 缓解 |
|------|------|------|
| 大文件内存占用 | zip/哈希若全读内存 | 流式处理 |
| 目录穿越 | 用户传入 `..` | `filepath.Clean` + 前缀校验 |
| 日志敏感信息 | 路径可能含隐私 | 可添加 `-log-redact` |
| Prometheus 暴露 | 未加 auth 可泄露内部信息 | 支持独立监听 + 可选 token |

## 7. 决策待定 (Open Decisions)

| 项目 | 问题 | 选项 | 当前倾向 |
|------|------|------|----------|
| 配置来源 | Flag vs 文件 | 仅 flag / 支持 YAML | 先 flag, 后可选文件 |
| 认证扩展 | Token / Basic / JWT | 分层中间件 | 逐步迭代：Token→Basic→JWT |
| 多监听端口 | /metrics 分离 | 同端口路径 vs 独立端口 | 独立端口更清晰 |

## 8. 扩展生态 (Integration)

- goreleaser: 自动生成 checksum / 发布到 GitHub Release
- 容器镜像: 多架构 (amd64/arm64) + 非 root 运行
- 包管理：Homebrew / Scoop / AUR（待社区反馈）

## 9. 快速任务池 (Quick Wins)

| 任务 | 价值 | 预估 | 备注 |
|------|------|------|------|
| `-log-json` | 可观测扩展基础 | 小 | 重用现有中间件 |
| Request ID | 链路追踪辅助 | 小 | 可直接 UUID |
| `/healthz` | K8s 集成 | 极小 | 直接 200 |
| hash 输出 | 验证下载 | 小 | `?sha256=1` |

## 10. 贡献指引 (Contribution Hints)

- 新增 flag：保持简短、语义清晰；README & README.en 同步
- 修改构建：更新 CI workflow 与 CHANGELOG
- 新依赖：评估许可证（优先 MIT / BSD / Apache-2.0）
- 性能相关：提供 Benchmark 或说明数据

---

欢迎根据优先级勾选 / 调整；如果想我先实现其中某个模块，直接告诉我编号或表格条目即可。

## 11. 进阶演进建议 (Advanced Evolution)

> 面向中长期提升，涵盖架构抽象、扩展性、性能深挖、安全加固与生态整合。此区建议可在核心功能稳定后择优推进。

### 11.1 架构与可扩展性

- 中间件链抽象：将现有 logging / token / readonly 重构为统一 `type Middleware func(http.Handler) http.Handler` 列表，支持动态启用/排序。
- 模块化构建标签：把可选功能（如 zip 打包、JWT、Prometheus）用 build tags 隔离，减少最小发行版体积。
- 插件机制 (实验)：通过 Go `plugin` (仅 *nix) 或简单进程回调（HTTP/STDIO）方式加载外部扩展（如自定义鉴权）。
- 配置层抽象：引入轻量配置解析（YAML/TOML/JSON 可选），实现 flag → config 叠加优先级策略。
- 统一错误响应格式：可选 `-error-json` 输出 `{code,message}`，便于脚本消费。

### 11.2 性能与资源优化

- 零拷贝/Sendfile 优化：针对大文件在支持平台上利用 `http.ServeContent` + 缓存 stat 信息减少系统调用。
- 目录索引缓存：在启用目录浏览模式下缓存文件列表（含 ETag），增量刷新（基于 mtime）。
- 分块预读：为顺序下载大文件提供可选 readahead（注意与内核页缓存策略协调）。
- PGO（Profile-Guided Optimization）：在构建 Release 前采集真实访问型 profile 进行 `go build -pgo`。
- 压缩策略自适应：根据文件类型+大小/阈值决定是否压缩（避免小文件压缩开销）。

### 11.3 安全强化

- mTLS（双向 TLS）支持：用于敏感内网场景；新增 `-tls-cert/-tls-key/-tls-ca`。
- 访问速率控制：在 token 或 IP 维度应用简单令牌桶+惩罚时间。
- 审计日志分流：安全相关事件（认证失败、拒绝访问、限速触发）输出到独立 channel / 文件。
- 临时一次性下载链接：生成带签名 + 过期时间的 URL（`/dl/<id>?sig=...`）。
- Hash 验证头：在响应中追加 `X-File-SHA256`（可配置），帮助客户端校验完整性。

### 11.4 可观测与运维

- OpenTelemetry 集成：可选导出 trace/span（文件读取、发送耗时拆分）。
- 自适应指标聚合：对大量文件的高频访问做分桶统计（Top-N 热点）。
- Profiling 开关：`-pprof` 启用内置 `net/http/pprof`（默认 off，避免敏感信息泄露）。
- 慢请求日志：超过阈值（如 500ms）单独记录，支持 JSON 模式。
- 滚动日志策略：写入文件时支持大小轮转（避免单日志过大）。

### 11.5 分发与生态

- goreleaser 全自动发布：多平台产物 + SBOM + 签名（cosign / minisign）。
- Docker 镜像加固：非 root 用户、只读根文件系统、HEALTHCHECK、multi-stage 构建最小化。
- Helm Chart：发布至 OCI registry，简化 K8s 部署。
- Kustomize 支持：提供可覆盖 patch（资源限额/节点选择）。
- Nix flake 打包：便于 NixOS 用户快速集成。

### 11.6 高级功能延展

- 数据源适配层：支持除本地 FS 外的后端（只读）如 S3 / Azure Blob / GCS（抽象 `Storage` 接口）。
- Chunk/Ranged Upload (可选)：扩展为轻量“单向分发+回传”通道（需要明确安全策略）。
- WebSocket 事件：广播文件系统变化（新增/删除），用于自动刷新 UI。
- Delta/断点续传增强：对大文件提供校验块信息（SHA256 per chunk）支持外部同步工具。
- 可插拔鉴权链：Token→Basic→OIDC（`-oidc-issuer/-oidc-client-id`）。

### 11.7 稳定性与健壮性

- 混沌测试脚本：模拟端口占用、磁盘满、权限拒绝、长时间传输中断。
- 可选写入保护：监测目录突变（新增/修改数量激增）并告警（潜在加密软件行为）。
- 自恢复端口：监听失败后退避重试（限制次数），输出 JSON 事件。
- 运行时自检：启动后执行一次自检（目录可读/剩余磁盘空间/TLS 文件存在）。

### 11.8 文档与开发者体验

- 单页文档网站：使用 mkdocs 或 docusaurus 自动生成 API/Flags 索引。
- 交互式 `--help`（分页 / 分类显示）。
- 示例仓库：演示与 Nginx 反代 / systemd / Docker Compose 结合用法。
- 教程脚本：一键生成自签名 TLS + 运行示例。

### 11.9 风险 / 取舍提醒

| 方向 | 潜在风险 | 缓解建议 |
|------|----------|----------|
| 引入太多依赖 | 体积 & 供应链风险 | 严格评审 / 选择纯标准库优先 |
| 插件机制 | 跨平台复杂 / 不稳定 | 首选进程/HTTP 扩展方式 |
| 多数据源 | 复杂度上升 | 先本地 FS 稳定→逐步只读后端 |
| 高度可配置 | 入口学习成本变高 | 分层：核心 flags / 高级 config |
| 性能优化过早 | 维护成本 | 基于真实 profiling 数据推进 |

### 11.10 推荐推进顺序 (Suggested Order)

1. JSON 日志 + Request ID（奠定 observability 基础）
2. Allow/Deny 路径控制（核心安全）
3. Prometheus 指标 + `/healthz`
4. 下载计数 + Top-N 热点统计
5. 限速 & 并发上限（资源保护）
6. 目录浏览 UI + Zip 打包（体验提升）
7. goreleaser + Docker 镜像加固（交付体系）
8. OIDC / mTLS / 临时链接（强化安全）
9. 多后端存储适配（扩展场景）
10. OpenTelemetry / Profiling / 慢日志（高级运维）

---

若你确定下一阶段要启动哪一组（比如“先做 0.4.0 可观测”或“优先安全 Allow/Deny”），告诉我即可直接落地实现细化任务列表。
