# Pipeline 文档

这是 Pipeline 项目的文档网站，使用 VitePress 构建。

## 开发

```bash
# 进入 docs 目录
cd docs

# 安装依赖
pnpm install

# 启动开发服务器
pnpm run dev

# 构建文档
pnpm run build

# 预览构建结果
pnpm run preview
```

## 部署

文档网站可以部署到 GitHub Pages、Netlify、Vercel 等平台。

### GitHub Pages

1. 构建文档：
```bash
cd docs
pnpm run build
```

2. 将 `.vitepress/dist` 目录部署到 GitHub Pages。

### Netlify

创建 `netlify.toml`：

```toml
[build]
  command = "cd docs && pnpm run build"
  publish = "docs/.vitepress/dist"
```

### Vercel

Vercel 会自动检测 VitePress 项目并配置构建。

## 文档结构

```
docs/
├── .vitepress/          # VitePress 配置
│   ├── config.js        # 配置文件
│   └── theme/           # 主题配置
├── guide/               # 指南文档
├── commands/            # 命令文档
├── architecture/        # 架构文档
├── best-practices/      # 最佳实践
└── index.md            # 首页
```

## 配置

配置文件位于 `.vitepress/config.ts`，可以修改：

- 网站标题和描述
- 导航栏和侧边栏
- 主题配置
- 搜索配置

## 贡献

欢迎贡献文档！请：

1. Fork 本仓库
2. 创建特性分支
3. 提交更改
4. 开启 Pull Request
