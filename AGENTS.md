# AGENTS.md

本文档为在此仓库中工作的 AI 代理提供指南。

**语言说明**: 在与用户对话时使用中文。所有代码、命令和文件路径保持英文。

## 项目概述

这是一个名为 `hardfetch` 的 Go CLI 工具，类似于 fastfetch/neofetch 的系统信息工具。该项目使用 Go 1.22+，遵循标准 Go 项目结构。

## 项目目标

Hardfetch 旨在成为一个快速、可定制的系统信息工具，类似于 fastfetch/neofetch，但使用 Go 编写以获得更好的跨平台兼容性和性能。主要特性包括：

1. **系统信息**: 显示操作系统、内核、主机名、运行时间
2. **硬件信息**:
   - CPU: 型号、核心数、线程数、频率、架构
   - GPU: 名称、供应商、显存、驱动版本（Windows WMI 支持）
   - 内存: 总量、已用、可用、剩余
   - 磁盘: 多磁盘支持，Windows 下显示盘符，Linux/macOS 下显示挂载点
3. **网络信息**: 本地IP/公网IP、网络接口
4. **可定制显示**: ASCII 徽标、颜色主题、输出格式
5. **跨平台**: 支持 Windows、Linux、macOS
6. **高性能**: 并发信息采集
7. **可配置**: JSON/YAML 配置文件、命令行选项
8. **模块化设计**: 选择要显示的特定模块

## 构建命令

### 基础构建
```bash
# 构建二进制文件
make build
# 或
go build -o hardfetch cmd/hardfetch/main.go

# 全局安装
make install
# 或
go install ./cmd/hardfetch

# 清理构建产物
make clean
```

### 开发构建
```bash
# 使用 race 检测器构建
go build -race -o hardfetch cmd/hardfetch/main.go

# 为多个平台构建
make build-all

# 为特定操作系统/架构构建
GOOS=linux GOARCH=amd64 go build -o hardfetch-linux-amd64 cmd/hardfetch/main.go
GOOS=darwin GOARCH=arm64 go build -o hardfetch-darwin-arm64 cmd/hardfetch/main.go
GOOS=windows GOARCH=amd64 go build -o hardfetch-windows-amd64.exe cmd/hardfetch/main.go
```

## 测试命令

### 运行所有测试
```bash
# 运行所有测试
make test
# 或
go test ./...

# 运行测试并显示详细输出
go test -v ./...

# 使用 race 检测器运行测试
go test -race ./...
```

> 注意：测试产物会放到 `dist/` 目录下

### 运行特定测试
```bash
# 在特定包中运行测试
go test ./internal/cli

# 运行特定测试函数
go test -run TestFunctionName ./internal/cli

# 运行带覆盖率的测试
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 性能基准测试
```bash
# 运行基准测试
go test -bench=. ./...

# 运行带内存分析的基准测试
go test -bench=. -benchmem ./...
```

## 代码质量工具

### Go 工具
```bash
# 格式化代码
go fmt ./...

# 检查可疑代码结构
go vet ./...

# 运行静态分析
go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
shadow ./...

# 检查未使用的依赖
go mod tidy -v
```

### 推荐代码检查工具
```bash
# 安装 golangci-lint（如果未安装）
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行 golangci-lint
golangci-lint run ./...
```

## 代码风格指南

### 导入组织
- 首先使用标准库导入，然后是第三方导入，最后是本地导入
- 各组导入之间用空行分隔
- 使用 `goimports` 自动格式化导入

示例:
```go
import (
    "fmt"
    "os"
    "strings"
    "github.com/spf13/cobra"
    "golang.org/x/text/cases"
    "hardfetch/internal/cli"
)
```

### 命名约定
- **包名**: 使用简短的小写单词 (例如: `cli`, `utils`)
- **变量**: 使用 camelCase (例如: `userName`, `maxRetries`)
- **常量**: 导出常量使用 CamelCase 或 UPPER_SNAKE_CASE
- **函数**: 使用 camelCase；导出的函数首字母大写
- **接口**: 适当使用 `-er` 后缀 (例如: `Reader`, `Writer`)

### 错误处理
- 函数调用后立即检查错误
- 使用 `fmt.Errorf` 和 `%w` 包装错误
- 在错误消息中提供上下文信息
- 错误时返回零值

示例:
```go
func ReadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read config %s: %w", path, err)
    }
    // ... 解析配置
}
```

### 函数设计
- 保持函数小而专注（最好少于 50 行）
- 提前返回以减少嵌套
- 有帮助时使用命名返回作为文档
- 为导出的函数添加完整句子文档

### 类型安全
- 尽可能使用具体类型而非 `interface{}`
- 为领域概念定义自定义类型
- 使用 comma-ok 形式的类型断言

### 注释和文档
- 为所有导出的函数、类型和变量添加文档
- 使用以句号结尾的完整句子
- 优先使用自文档化的代码而非注释
- 为非显而易见的逻辑添加注释

### 项目结构
```
hardfetch/
├── cmd/hardfetch/          # 主应用程序入口
├── internal/
│   ├── cli/                # CLI 特定逻辑
│   ├── detect/             # 信息采集模块（核心模块）
│   │   ├── detector.go     # 采集器接口和实现
│   │   ├── system.go       # 系统信息结构体
│   │   ├── hardware.go     # 硬件信息结构体
│   │   ├── network.go      # 网络信息结构体
│   │   └── collector/      # 平台特定采集实现
│   │       └── windows_collector.go  # Windows 采集实现
│   ├── display/            # 显示格式化
│   └── utils/              # 工具函数
├── configs/                # 配置文件
├── logos/                  # ASCII 徽标文件
├── go.mod                  # Go 模块定义
├── Makefile                # 构建自动化
└── README.md               # 项目文档
```

## 开发工作流程

### 添加新功能
1. 在 `internal/` 或 `pkg/` 下创建适当的目录结构
2. 为新功能编写测试
3. 更新 `cmd/hardfetch/main.go` 以集成新功能
4. 运行测试并确保通过
5. 如需要更新文档

### 添加依赖
```bash
# 添加新依赖
go get github.com/example/package

# 更新所有依赖
go get -u ./...

# 清理未使用的依赖
go mod tidy
```

### 版本管理
- 在 `internal/cli/version.go` 中更新版本
- 使用语义化版本 (MAJOR.MINOR.PATCH)
- 使用 `git tag v0.1.0` 标记发布

## Git 指南

### 提交信息
- 使用 conventional commits 格式: `type(scope): description`
- 类型: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
- 第一行保持在 50 个字符以内
- 需要时在提交正文中提供详细描述

### 分支策略
- `main`: 生产就绪代码
- `develop`: 功能集成分支
- 功能分支: `feature/description`
- 修复分支: `fix/description`

## 性能考虑

### 构建优化
```bash
# 使用优化构建
go build -ldflags="-s -w" -o hardfetch cmd/hardfetch/main.go

# 剥离调试信息以减小二进制文件
go build -trimpath -o hardfetch cmd/hardfetch/main.go
```

### 运行时性能
- 避免热路径中的不必要分配
- 为频繁分配的对象使用 `sync.Pool`
- 性能关键时使用 `go tool pprof` 进行分析

## 常见任务

### 添加新模块
1. 在 `internal/modules/` 中创建模块实现
2. 将模块添加到适当类别 (system, hardware, network)
3. 更新 `cmd/hardfetch/main.go` 将模块包含在显示函数中
4. 为模块编写测试
5. 更新帮助文本和文档

### 调试
```bash
# 使用调试日志运行
DEBUG=1 ./hardfetch

# 使用 delve 调试器
dlv debug cmd/hardfetch/main.go
```

### 跨平台编译
```bash
# 为多个平台构建
GOOS=darwin GOARCH=arm64 go build -o hardfetch-darwin-arm64 cmd/hardfetch/main.go
GOOS=linux GOARCH=amd64 go build -o hardfetch-linux-amd64 cmd/hardfetch/main.go
GOOS=windows GOARCH=amd64 go build -o hardfetch-windows-amd64.exe cmd/hardfetch/main.go
```

## 质量保证检查清单

提交代码前:
- [ ] 所有测试通过
- [ ] 代码已使用 `go fmt` 格式化
- [ ] 无 `go vet` 警告
- [ ] 无代码检查问题
- [ ] 如需要更新文档
- [ ] 保持向后兼容

## 资源

- [Go 文档](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go 代码审查注释](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go 谚语](https://go-proverbs.github.io/)
