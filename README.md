# hardfetch

一个使用 Go 编写的类似 fastfetch/neofetch 的系统信息工具。

## 特性

- **系统信息**: 操作系统、内核、主机名、运行时间、Shell
- **桌面环境**: 窗口管理器、主题、字体、终端、区域设置
- **硬件信息**:
  - CPU: 型号、核心数、线程数、频率、架构
  - GPU: 名称、供应商、显存、驱动版本（Windows WMI 支持）
  - 内存: 总量、已用、可用、使用率
  - 交换空间: 总量、已用、使用率
  - 磁盘: 多磁盘支持，Windows 下显示盘符，Linux/macOS 下显示挂载点
  - 电池: 电量、充电状态
- **网络信息**: 本地IP、公网IP、网络接口
- **可定制显示**: ASCII 徽标、颜色主题、输出格式
- **跨平台**: 支持 Windows、Linux、macOS
- **高性能**: 并发信息采集
- **可配置**: JSON/YAML 配置文件、命令行选项
- **模块化设计**: 选择要显示的特定模块

## 安装

### 从源码安装

```bash
# 克隆仓库
git clone <repository-url>
cd hardfetch

# 构建并安装
make install
# 或
go install ./cmd/hardfetch
```

### 使用 go install

```bash
go install hardfetch/cmd/hardfetch@latest
```

## 使用方法

```bash
# 使用默认设置显示系统信息
hardfetch

# 显示特定模块
hardfetch --modules system,cpu,memory

# 显示所有可用模块
hardfetch --all

# 不显示 ASCII 徽标
hardfetch --no-logo

# 不显示颜色
hardfetch --no-colors

# 显示版本
hardfetch --version

# 显示帮助
hardfetch --help

# 生成配置文件
hardfetch --gen-config

# 列出所有可用模块
hardfetch --list-modules
```

## 可用模块

- **os**: 操作系统、版本、内核
- **host**: 主机名、主机型号
- **kernel**: 内核版本
- **uptime**: 系统运行时间
- **shell**: 默认 Shell
- **wm**: 窗口管理器
- **theme**: 桌面主题
- **font**: 系统字体
- **terminal**: 终端模拟器
- **locale**: 系统区域设置
- **cpu**: CPU 型号、架构、核心数、线程数、频率
- **gpu**: GPU 名称、供应商、显存、驱动版本
- **memory**: 总量、已用、可用、剩余内存
- **swap**: 交换空间使用情况
- **disk**: 多磁盘信息，支持盘符/挂载点
- **network**: 本地IP、公网IP、网络接口
- **battery**: 电池状态

## 开发

### 构建命令

```bash
# 构建二进制文件到 dist/ 目录
make build

# 运行测试
make test

# 清理构建产物（删除 dist/ 目录）
make clean

# 全局安装
make install

# 为多个平台构建（Linux、macOS、Windows）
make build-all

# 为特定平台构建
make build-linux    # 构建 Linux 二进制文件
make build-darwin   # 构建 macOS 二进制文件
make build-windows  # 构建 Windows 二进制文件
```

### 测试

```bash
# 运行所有测试
go test ./...

# 运行带覆盖率的测试
go test -cover ./...

# 运行特定测试
go test -run TestMainVersion ./cmd/hardfetch
```

### 代码质量

```bash
# 格式化代码
go fmt ./...

# 检查代码
go vet ./...

# 运行代码检查器（需要 golangci-lint）
golangci-lint run ./...
```

## 项目结构

```
hardfetch/
├── cmd/hardfetch/          # 主应用程序入口
│   ├── main.go             # CLI 入口点
│   └── main_test.go        # CLI 测试
├── internal/
│   ├── cli/                # CLI 特定逻辑
│   │   ├── version.go      # 版本常量
│   │   ├── config.go       # 配置管理
│   │   └── ...
│   ├── detect/             # 信息采集模块（核心模块）
│   │   ├── detector.go     # 采集器接口和实现
│   │   ├── system.go       # 系统信息结构体
│   │   ├── hardware.go     # 硬件信息结构体
│   │   ├── network.go      # 网络信息结构体
│   │   └── collector/      # 平台特定采集实现
│   │       └── windows_collector.go  # Windows 采集实现
│   ├── display/            # 显示格式化
│   │   ├── ascii.go        # ASCII 艺术渲染
│   │   ├── colors.go       # 颜色支持
│   │   └── formatter.go    # 输出格式化
│   └── utils/              # 工具函数
├── dist/                   # 构建输出（生成，已 gitignore）
├── configs/                # 配置文件
│   └── default.json        # 默认配置
├── logos/                  # ASCII 徽标文件
├── go.mod                  # Go 模块定义
├── go.sum                  # 依赖校验和
├── Makefile                # 构建自动化
├── .gitignore              # Git 忽略规则
├── .golangci.yml           # 代码检查器配置
├── AGENTS.md               # 开发指南
└── README.md               # 项目文档
```

## 平台特定实现

### Windows
- **CPU 检测**: 通过注册表查询获取 CPU 型号和频率
- **GPU 检测**: 使用 WMI 查询 (Win32_VideoController) 获取 GPU 信息
- **磁盘检测**: 使用 Windows API 支持多磁盘和盘符
- **内存**: 使用 Windows GlobalMemoryStatusEx API

### Linux/macOS
- **CPU 检测**: 使用 runtime.NumCPU() 的通用实现
- **GPU 检测**: 操作系统特定的占位符实现
- **磁盘检测**: 单磁盘占位符
- **内存**: 占位符值（将使用系统特定调用实现）

## GPU 信息检测

GPU 模块提供详细的显卡信息：

### Windows 实现
- 使用 Windows Management Instrumentation (WMI) 通过 PowerShell
- 查询 `Win32_VideoController` 类获取 GPU 详情
- 提取: 名称、供应商、显存、驱动版本
- 支持多个 GPU
- 自动供应商检测（NVIDIA、AMD、Intel、Microsoft）


## 添加新功能

1. 在 `internal/` 下创建适当的目录结构
2. 为新功能编写测试
3. 更新 `cmd/hardfetch/main.go` 以集成新功能
4. 运行测试并确保通过
5. 如需要更新文档

## 代码风格

请遵循 [AGENTS.md](AGENTS.md) 中的指南：
- 导入组织
- 命名约定
- 错误处理
- 函数设计
- 类型安全
- 注释和文档

## 许可证

MIT
