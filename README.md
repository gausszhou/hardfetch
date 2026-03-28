# hardfetch

类似 fastfetch/neofetch 的 Go CLI 系统信息工具。

## 特性

- 系统信息、硬件信息、网络信息、电池状态
- 跨平台支持 (Windows/Linux/macOS)
- 高性能并发采集
- 可配置显示

## 安装

```bash
go install github.com/gausszhou/hardfetch/cmd/hardfetch@latest
```

或从源码：

```bash
git clone https://github.com/gausszhou/hardfetch.git
cd hardfetch
make install
```

## 使用

```bash
hardfetch           # 默认显示
hardfetch -d        # 调试模式（性能分析）
hardfetch -v        # 版本
hardfetch -h        # 帮助
```

## 构建

```bash
make build   # 构建
make test    # 测试
make install # 安装
```
