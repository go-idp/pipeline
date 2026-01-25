# 文档部署说明

本文档说明如何使用 GitHub Actions 自动部署 Pipeline 文档到 GitHub Pages。

## 自动部署

文档已经配置了 GitHub Actions 自动部署。每次推送到 `main` 分支时，会自动构建并部署文档。

### 工作流文件

`.github/workflows/docs.yml` - 自动部署工作流

### 部署步骤

1. **构建**: 安装依赖并构建 VitePress 文档
2. **上传**: 将构建产物上传为 artifact
3. **部署**: 部署到 GitHub Pages

## 手动触发

你也可以手动触发部署：

1. 进入 GitHub Actions 页面
2. 选择 "Deploy VitePress site to Pages" workflow
3. 点击 "Run workflow"

## GitHub Pages 配置

### 1. 启用 GitHub Pages

1. 进入仓库 Settings → Pages
2. Source 选择 "GitHub Actions"

### 2. Base 路径配置

文档的 base 路径在 `docs/.vitepress/config.js` 中配置：

```js
base: '/pipeline/'
```

如果你的仓库名是 `pipeline`，那么：
- GitHub Pages URL: `https://go-idp.github.io/pipeline/`
- Base 路径: `/pipeline/` ✅ (当前配置)

如果使用自定义域名或部署到根路径，需要修改为：

```js
base: '/'
```

## 本地测试

在部署前，可以在本地测试构建：

```bash
# 进入 docs 目录
cd docs

# 安装依赖
pnpm install

# 构建文档
pnpm run build

# 预览构建结果
pnpm run preview
```

## 故障排查

### 构建失败

1. 检查 Node.js 版本（需要 20+）
2. 检查 `docs/package.json` 中的依赖
3. 查看 GitHub Actions 日志

### 页面无法访问

1. 检查 GitHub Pages 是否已启用
2. 检查 base 路径是否正确
3. 等待几分钟让 DNS 生效

### 资源路径错误

如果图片、CSS 等资源无法加载：

1. 检查 base 路径配置
2. 确保资源路径使用相对路径或绝对路径（以 base 开头）

## 自定义域名

如果要使用自定义域名：

1. 在仓库 Settings → Pages 中设置 Custom domain
2. 将 base 路径改为 `/`
3. 添加 CNAME 文件到 `docs/public/` 目录

## 更多信息

- [VitePress 部署文档](https://vitepress.dev/zh/guide/deploy)
- [GitHub Pages 文档](https://docs.github.com/en/pages)
