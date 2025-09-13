# DHGoHttp

[English Version](./README.en.md)

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Build](https://github.com/PeiKeSmart/DHGoHttp/actions/workflows/build.yml/badge.svg)

一个极简的跨平台文件目录 HTTP 服务器，主打“放到目录里就能直接下载”。在 Windows 下可自动尝试提权并添加防火墙放行规则，方便局域网或外部机器访问。

## 功能特性

- 直接把当前工作目录作为静态文件根目录
- 免配置启动：`go run .` 或构建后的可执行文件即可
- Windows：
  - 自动检测是否管理员，若不是尝试触发 UAC 提升
  - 管理员模式自动添加（且避免重复添加）防火墙入站规则：`DHGoHttp-<端口>`
  - 已存在同名规则时不重复创建
- 可使用 `PORT` 环境变量自定义监听端口（默认 8080）
- 提供 `-no-firewall` 参数跳过提权与防火墙逻辑（便于快速本机调试）

## 快速开始

### 克隆或下载

```bash
# 方式一：git 克隆
git clone <your-repo-url> dhgohttp
cd dhgohttp

# 方式二：直接把 main.go 放到任意目录
```

### 运行（开发）

```bash
go run .
```

### 构建与运行（推荐在 Windows 测试提权逻辑时使用）

```bash
go build -o dhgohttp.exe .
./dhgohttp.exe
```

首次在 Windows 下非管理员运行时，会弹出 UAC 询问；同意后新进程（管理员）会添加防火墙规则并监听端口，原进程退出。

## 访问文件

假设你的目录结构：

```text
E:/Project/DHGoHttp/
  ├── main.go
  ├── README.md
  └── example-download.sh
```

访问：

```text
http://<主机IP>:8080/example-download.sh
```

本机可直接：

```text
http://localhost:8080/README.md
```

Curl 下载：

```bash
curl -O http://localhost:8080/example-download.sh
```

## 命令行参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `-no-firewall` | 跳过自动提权 & 防火墙规则添加 | `./dhgohttp.exe -no-firewall` |
| `-elevated` | 内部使用标记，防止重复提权 | （不手动使用） |
| `-port` | 指定起始端口（优先级高于 PORT 环境变量；若被占用自动递增） | `./dhgohttp.exe -port 9000` |
| `-dir` | 指定共享根目录（默认当前工作目录） | `./dhgohttp.exe -dir C:/Files` |
| `-max-port-scan` | 端口占用递增最大尝试次数（默认 50） | `./dhgohttp.exe -port 9000 -max-port-scan 30` |
| `-bind` | 绑定地址（默认所有网卡 0.0.0.0） | `./dhgohttp.exe -bind 127.0.0.1` |
| `-token` | 启用 Token 校验 (Header: X-Token 或 ?token=) | `./dhgohttp.exe -token SECRET123` |
| `-readonly` | 禁止目录列出（仅按具体文件路径访问） | `./dhgohttp.exe -readonly` |

## 环境变量

| 变量 | 作用 | 示例 |
|------|------|------|
| `PORT` | 指定监听端口 | `PORT=9000 go run .` |

> 注意：`PORT` 只写端口号，不含冒号。程序内部会自动补 `:`。

## Windows 防火墙与提权说明

1. 非管理员运行会尝试 `ShellExecuteW` + `runas` 触发 UAC 提升。
2. 若用户拒绝：继续普通权限运行，但不会自动添加防火墙放行规则，此时只能本机或已经被系统允许的网络访问。
3. 防火墙规则名格式：`DHGoHttp-<端口号>`，例如：`DHGoHttp-8080`。
4. 检查规则是否存在：

```powershell
netsh advfirewall firewall show rule name="DHGoHttp-8080"
```

1. 手动添加（管理员 PowerShell）：

```powershell
netsh advfirewall firewall add rule name="DHGoHttp-8080" dir=in action=allow protocol=TCP localport=8080
```

1. 手动删除：

```powershell
netsh advfirewall firewall delete rule name="DHGoHttp-8080"
```

## 典型使用场景

- 临时分享某个目录里的脚本、安装包、日志
- 内网机器之间快速传输
- 在 CI/CD 或容器中临时暴露构建产物（生产环境不推荐直接裸用）

## 常见问题 (FAQ)

### 1. 为什么我外网/局域网另一台机器访问不了？

- 可能未提权 → 未创建防火墙放行规则
- 端口被系统或安全软件阻拦
- 服务器监听的是默认 8080，检查是否被占用
- 目标机器防火墙策略限制

### 2. 如何指定端口？自动递增是怎样的？

方式一：使用 flag：

```bash
./dhgohttp.exe -port 9000
```

若 9000 被占用，会依次尝试 9001、9002 ...，最多尝试 `-max-port-scan` 次（默认 50）。

方式二：使用环境变量：

```bash
PORT=9000 go run .
```

或 Windows PowerShell：

```powershell
$env:PORT=9000; go run .
```

### 3. 我不想每次弹 UAC，怎么办？

- 直接以管理员打开终端再运行
- 或使用 `-no-firewall`，然后手动提前创建好规则

### 4. 为什么重复运行显示“规则已存在，跳过创建”？

程序启动前会检测规则，存在就不再调用 `netsh`，避免多余日志。

### 5. 自定义目录与自动递增端口支持了吗？

已支持：`-dir` 指定目录；端口会自动递增寻找可用端口（日志会显示尝试序列）；`-readonly` 禁止列目录；`-token` 开启简单鉴权。

### 6. 如何优雅退出？

直接 Ctrl+C（或发送 SIGTERM）程序会：

1. 停止接受新连接
2. 等待最多 5 秒处理中的请求
3. 若本进程曾成功创建防火墙规则，则尝试删除该规则
4. 输出退出日志后结束

## 代码结构

- `main.go`：全部逻辑（启动、提权、防火墙、静态文件托管）
- `example-download.sh`：示例文件
- `README.md`：使用说明

## 安全注意事项

- 不做访问控制，任何能连到端口的客户端都可下载目录中所有文件。
- 不建议直接暴露到公网；若需公网使用，请添加：反向代理 / 认证 / 访问白名单。
- 提权仅在 Windows 下执行；Linux / macOS 需手动防火墙放行（如 `ufw` 或安全组）。

## 后续可选增强方向

- 访问日志 / 下载统计
- 优雅退出删除防火墙规则（当前仅添加不删除）
- 简单的只读 Token 访问控制
- 目录访问控制（黑/白名单）
- 多平台发布产物（在 CI 矩阵构建后生成）

## 版本与更新日志

请查看 [CHANGELOG](./CHANGELOG.md)。当前初始版本：`0.1.0`。

## 许可证

本项目采用 MIT License，详见 [LICENSE](./LICENSE)。

---
欢迎提出需求或直接继续拓展哪一项，我可以帮你补齐。😉
