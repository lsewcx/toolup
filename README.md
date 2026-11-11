# toolup

快速在新服务器上安装常用工具的命令行工具。

`toolup` 会自动识别当前发行版和包管理器，为不同平台提供一致的安装体验：系统仓库收录的工具走包管理器，未收录的工具走官方安装脚本或官方二进制压缩包。

## 特性

- 自动识别发行版与包管理器（apt / dnf）
- 统一命令安装：`toolup install <tool>`
- 智能选择安装方式：包管理器优先，其次官方安装脚本 / 官方二进制压缩包
- 脚本安装自动选择 `curl` 或 `wget`，两者都缺失时输出 warning 提示安装
- 二进制压缩包安装会校验官方 SHA256，并用内置解压逻辑处理，无需额外安装 `unzip`
- 通过脚本或压缩包安装的工具已安装时自动跳过，可重复执行
- 结构化日志 + 彩色输出，支持 `NO_COLOR` 与调试模式

## 支持的环境

| 平台 | 包管理器 | 状态 |
| --- | --- | --- |
| Debian / Ubuntu / Linux Mint / Pop!_OS / Kali | apt | ✅ |
| Fedora / RHEL / CentOS / Rocky Linux / AlmaLinux | dnf | ✅ |
| macOS | - | 脚本 / 压缩包类工具可用（如 uv、ossutil） |
| Windows | - | ❌ 暂不支持 |

- 通过包管理器安装时，非 root 用户会自动使用 `sudo`；脚本或压缩包安装需要 root 权限时同样会提权
- ossutil 2.0 支持 Linux / macOS 的 386、amd64、arm、arm64 平台
- 从源码构建需要 Go 1.26+

## 已支持的工具

| 工具 | 安装方式 | 说明 |
| --- | --- | --- |
| [git](https://git-scm.com/) | apt-get / dnf | 系统包管理器安装 |
| [uv](https://docs.astral.sh/uv/) | 官方安装脚本 | 优先 `curl`，回退 `wget`；已安装则跳过 |
| [ossutil](https://www.alibabacloud.com/help/zh/oss/developer-reference/ossutil-overview/) | 官方二进制压缩包（2.4.1） | 校验 SHA256，安装到 `/usr/local/bin/ossutil` 并软链 `/usr/bin/ossutil`；需要 `sudo` |

## 安装 toolup

### 从源码构建

```bash
git clone git@github.com:lsewcx/toolup.git
cd toolup

# 静态编译，产物可直接拷到服务器使用
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o toolup .

# 可选：安装到 PATH
sudo install -m 0755 toolup /usr/local/bin/toolup
toolup --help
```

也可以使用 `go install`：

```bash
go install github.com/lsewcx/toolup@latest
```

## 使用

```bash
toolup install <tool>
```

示例：

```bash
toolup install git
toolup install uv
```

安装 uv 时的输出示例：

```text
INF 使用官方脚本安装 tool=uv script=https://astral.sh/uv/install.sh downloader=curl
downloading uv 0.13.7 x86_64-unknown-linux-gnu
INF 安装完成，如果当前终端找不到该命令，请重新打开终端或重新加载 shell 配置 tool=uv
```

> uv 官方脚本会安装到 `~/.local/bin` 并写入 shell 配置。如果当前终端还找不到 `uv`，请重新打开终端，或执行 `source ~/.bashrc`（其他 shell 同理）。

如果系统既没有 `curl` 也没有 `wget`，会输出 warning 并退出：

```text
WRN 缺少下载工具，无法安装 uv hint="请先安装 curl 或 wget，例如：apt install curl 或 dnf install wget"
Error: 未找到 curl 或 wget，请先安装其中之一后重试（例如：apt install curl 或 dnf install wget）
```

安装 ossutil 2.0（文档当前版本 2.4.1）：

```bash
toolup install ossutil
```

`toolup` 会按当前平台下载官方 zip、校验文档提供的 SHA256、解压后把 `ossutil` 安装到 `/usr/local/bin/`，并创建 `/usr/bin/ossutil` 软链接（非 root 用户会通过 `sudo` 执行）。安装完成后可执行 `ossutil config` 完成配置。

> ossutil 的下载地址与 SHA256 均取自阿里云官方文档；升级版本时需要同步更新工具目录中的版本、地址和校验和。

## 工作原理

执行 `toolup install <tool>` 时：

1. 在工具目录（catalog）中查找该工具；
2. 检测系统信息，得到发行版与包管理器；
3. 若工具声明了当前包管理器的包名，走包管理器安装（apt / dnf，必要时 `sudo`）；
4. 否则若工具声明了官方安装脚本地址，下载并执行官方脚本（优先 `curl`，其次 `wget`）；
5. 否则若工具声明了官方二进制压缩包，下载、校验 SHA256、解压并安装（必要时 `sudo`）；
6. 都没有则提示不支持。

## 项目结构

```text
.
├── cmd/                      # Cobra 命令入口（root、install）
├── internal/
│   ├── app/                  # 安装流程编排
│   ├── execx/                # 命令执行封装（含 sudo 提权）
│   ├── installer/            # apt / dnf / 官方脚本 / 官方压缩包安装器
│   │   └── catalog/          # 工具目录（git、uv、ossutil）
│   ├── logging/              # slog + tint 日志初始化
│   └── utils/                # 系统与包管理器检测
├── main.go
└── go.mod
```

## 开发

```bash
# 运行测试
go test ./...

# 静态检查与格式
go vet ./...
gofmt -l .

# 构建
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o toolup .
```

调试日志：

```bash
TOOLUP_DEBUG=1 ./toolup install uv
```

交叉编译示例：

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/toolup-linux-amd64 .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/toolup-linux-arm64 .
```

## 日志与环境变量

| 变量 | 作用 |
| --- | --- |
| `TOOLUP_DEBUG=1` | 输出调试日志（包含系统信息检测、源码位置） |
| `NO_COLOR=1` | 关闭彩色输出（遵循 [no-color.org](https://no-color.org/)） |
| `TERM=dumb` | 关闭彩色输出 |

## 安全提示

脚本安装方式会下载并执行工具官方的安装脚本，二进制压缩包会校验官方文档提供的 SHA256。请确认下载来源可信；安装时日志会打印实际使用的脚本地址或安装包地址。

## 许可证

[MIT](LICENSE) © 2026 lsewcx
