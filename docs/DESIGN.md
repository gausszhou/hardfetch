# Hardfetch 设计文档

## 概述

Hardfetch 是一个跨平台的 Go CLI 工具，用于获取和显示系统/硬件信息，类似于 fastfetch 或 neofetch。它使用 [gopsutil](https://github.com/shirou/gopsutil) 进行跨平台系统信息采集。

## 架构

### 核心组件

```
hardfetch/
├── main.go              # 入口点，CLI 参数解析
├── internal/
│   ├── detect/           # 检测协调层
│   │   ├── detect.go     # Detector 接口，并发采集
│   │   ├── detect_type.go # 数据类型和格式化器
│   │   └── detect_platform.go # 平台特定检测
│   ├── modules/          # 硬件信息采集器
│   │   ├── cpuinfo/      # CPU 信息
│   │   ├── memory/       # 内存和交换空间
│   │   ├── disk/         # 磁盘分区和使用情况
│   │   ├── gpuinfo/      # GPU 信息 (NVIDIA, AMD, Intel)
│   │   ├── network/      # 网络接口
│   │   ├── battery/      # 电池状态
│   │   └── sys/          # 系统信息 (OS, 内核, 主机名)
│   ├── display/         # 输出格式化
│   ├── logger/          # 日志工具
│   ├── info/            # 版本信息
│   └── cli/             # CLI 常量
```

### 数据流

1. **入口点** (`main.go`)
   - 解析 CLI 标志 (`-d`, `-p`, `-v`, `-h`)
   - 初始化日志器
   - 调用 `detect.Detect(detect.GetCoreDetectors()...)` 采集所有信息
   - 将结果传递给 `display.PrintResult()`
   - 支持 `--pprof` 生成性能分析文件

2. **检测层** (`internal/detect/`)
   - `Detector` 接口，包含 `Name()` 和 `Detect()` 方法
   - 三个核心检测器：`system`、`hardware`、`network`
   - 使用 `sync.Once` 实现单次初始化（记忆化）
   - 使用 `sync.WaitGroup` 实现并发采集

3. **硬件模块** (`internal/modules/`)
   - 每个模块实现平台特定检测
   - 使用 gopsutil 进行跨平台 API 调用
   - GPU 检测使用外部 CLI 工具 (nvidia-smi, rocm-smi, lspci)
   - 所有模块返回类型化结构体

4. **显示层** (`internal/display/`)
   - 使用 ANSI 颜色代码格式化数据
   - 渲染：系统、CPU、GPU、内存、磁盘、网络、电池
   - 使用 bytes.Buffer 高效构建字符串

5. **信息模块** (`internal/info/`)
   - 存储应用元数据：版本号、作者、仓库地址
   - 集中管理版本信息，便于更新

## 关键设计模式

### 检测器模式

```go
type Detector interface {
    Name() string
    Detect() (any, error)
}
```

每个检测器作为独立的 goroutine 运行，并发采集数据。

### 记忆化

```go
var result     *Result
var resultOnce sync.Once

func Detect(detectors ...Detector) *Result {
    resultOnce.Do(func() {
        result = &Result{}
        collectAll(detectors)
    })
    return result
}
```

使用 `sync.Once` 确保每次调用只运行一次检测。

### 平台特定检测

模块检测 `runtime.GOOS` 并分派到平台特定函数：

- `sys/sys.go`: Windows (PowerShell), Darwin (system_profiler), Linux (/proc)
- `gpuinfo/gpuinfo.go`: nvidia-smi, rocm-smi, lspci
- `battery/battery.go`: Windows (WMI), Darwin (pmset), Linux (/sys/class/power_supply)

### 性能计时器

使用 `logger.StartTimer()` 在调试模式下测量检测时间：

```go
t := logger.StartTimer("hardware:cpu")
defer t.Stop()
// ... 检测代码
```

## 数据类型

### SystemInfo
- OS, Arch, Kernel, Hostname, Host, Uptime

### HardwareInfo
- CPU (Model, Cores, Threads, Frequency)
- Memory (Total, Used, Available, Free)
- Swap (Total, Used, Free)
- Disks (Drive, Total, Used, Free, FileSystem)
- GPUs (Name, VRAM, Frequency, Type, DriverVersion)
- Battery (Percentage, Status)

### NetworkInfo
- Hostname, LocalIP, PublicIP
- Interfaces (Name, IPAddress, MACAddress)

## 并发模型

1. **顶层**: 3 个 goroutine (system, hardware, network)
2. **硬件层**: 5 个 goroutine (cpu, memory, gpu, battery, disk)
3. **GPU 检测**: 3 个 goroutine (NVIDIA, AMD, Intel) 在 Windows 上

总计：最多 11 个并发 goroutine 以获得最佳性能。

## 错误处理

- 每个模块返回错误但使用零值继续
- 调试模式下记录错误
- 显示层优雅处理 nil/缺失数据

## 依赖

- `github.com/shirou/gopsutil/v4` - 跨平台系统信息
- `golang.org/x/sys` - 底层系统调用
- 标准库: `sync`, `context`, `runtime`, `bytes`, `fmt`, `strings`, `runtime/pprof`

## 构建目标

- Windows (amd64, arm64)
- Linux (amd64, arm64)
- macOS (amd64, arm64)
