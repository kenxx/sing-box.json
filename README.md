# sing-box JSON Schema Generator

这是一个用于自动生成 sing-box 配置文件的 JSON Schema 的 Go 工具。

## 功能

- 使用 `github.com/invopop/jsonschema` 从 sing-box 的配置结构体自动生成 JSON Schema
- 通过 GitHub Actions 自动生成和更新 Schema 文件

## 依赖

- `github.com/sagernet/sing-box` - sing-box 核心库
- `github.com/invopop/jsonschema` - JSON Schema 生成库

## 使用方法

### 本地运行

1. 安装依赖：
```bash
go mod tidy
```

2. 运行生成器：
```bash
go run main.go
```

这将生成 `sing-box-config-schema.json` 文件。

### GitHub Actions

项目配置了 GitHub Actions 工作流，在以下情况下会自动运行：

- 推送到 `main` 或 `master` 分支
- 创建 Pull Request
- 手动触发（workflow_dispatch）

工作流会自动：
1. 检出代码
2. 设置 Go 环境
3. 下载并验证依赖
4. 生成 JSON Schema
5. 如果有变更，自动提交并推送

## 输出文件

生成的 JSON Schema 文件：`sing-box-config-schema.json`

该文件可以用于：
- IDE 自动补全和验证
- 配置文件的验证工具
- 文档生成

## Release 发布

当推送以 `v` 开头的 tag（如 `v1.0.0`）时，会自动创建 GitHub Release：

```bash
git tag v1.0.0
git push origin v1.0.0
```

Release 会自动包含：
- JSON Schema 文件作为附件
- 版本说明

## GitHub Pages

项目会自动部署到 GitHub Pages，可以通过以下方式访问：

- **主页**: https://kenxx.github.io/sing-box.json/
- **Schema 文件**: https://kenxx.github.io/sing-box.json/sing-box-config-schema.json

### 在配置文件中使用

在你的 sing-box 配置文件中添加 `$schema` 字段：

```json
{
  "$schema": "https://kenxx.github.io/sing-box.json/sing-box-config-schema.json",
  "log": {
    "level": "info"
  },
  ...
}
```

这样 IDE（如 VS Code）就能提供自动补全和验证功能。

## 许可证

本项目遵循与 sing-box 相同的许可证。

