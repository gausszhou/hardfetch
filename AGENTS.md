# AGENTS.md

AI 代理工作指南。

## 语言

与用户对话使用中文，代码/命令/文件路径使用英文。

## 项目

Go CLI 工具，类似 fastfetch/neofetch。使用 Go 1.22+。

## 构建

```bash
make build   # 构建
make install # 安装
make test    # 测试
```

## 代码风格

- 导入: 标准库 → 第三方 → 本地包
- 命名: 包名小写，变量 camelCase，导出首字母大写
- 错误: 立即检查，用 `fmt.Errorf` + `%w`

## 结构

```
hardfetch/
├── cmd/hardfetch/  # 入口
├── internal/
│   ├── cli/        # CLI
│   ├── detect/     # 信息采集
│   ├── display/    # 显示
│   └── logger/    # 日志
├── configs/       # 配置
└── logos/         # ASCII 徽标
```

## Git

使用 conventional commits: `type(scope): description`

类型: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

## 提交前

- 测试通过
- `go fmt`
- `go vet`
