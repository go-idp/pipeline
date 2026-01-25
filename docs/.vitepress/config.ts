import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Pipeline',
  description: 'Powerful workflow execution engine',
  base: '/pipeline/',
  
  head: [
    ['link', { rel: 'icon', href: '/favicon.ico' }],
    ['meta', { name: 'theme-color', content: '#3eaf7c' }],
  ],

  locales: {
    root: {
      label: 'English',
      lang: 'en',
      title: 'Pipeline',
      description: 'Powerful workflow execution engine',
      themeConfig: {
        nav: [
          { text: 'Home', link: '/' },
          { text: 'Guide', link: '/guide/' },
          { text: 'Commands', link: '/commands/' },
          { text: 'Architecture', link: '/architecture/' },
          { text: 'Best Practices', link: '/best-practices/' },
          { text: 'GitHub', link: 'https://github.com/go-idp/pipeline' },
        ],
        sidebar: {
          '/guide/': [
            {
              text: 'Getting Started',
              items: [
                { text: 'Introduction', link: '/guide/' },
                { text: 'Installation', link: '/guide/installation' },
                { text: 'First Pipeline', link: '/guide/getting-started' },
                { text: 'Configuration', link: '/guide/configuration' },
                { text: 'Core Concepts', link: '/guide/concepts' },
              ],
            },
          ],
          '/commands/': [
            {
              text: 'Commands',
              items: [
                { text: 'Overview', link: '/commands/' },
                { text: 'run', link: '/commands/run' },
                { text: 'server', link: '/commands/server' },
                { text: 'client', link: '/commands/client' },
              ],
            },
          ],
          '/architecture/': [
            {
              text: 'Architecture',
              items: [
                { text: 'Overview', link: '/architecture/' },
                { text: 'System Architecture', link: '/architecture/system' },
                { text: 'Error Handling', link: '/architecture/error-handling' },
                { text: 'Optimization', link: '/architecture/optimization' },
              ],
            },
          ],
          '/best-practices/': [
            {
              text: 'Best Practices',
              items: [
                { text: 'Best Practices', link: '/best-practices/' },
              ],
            },
          ],
        },
        editLink: {
          pattern: 'https://github.com/go-idp/pipeline/edit/main/docs/:path',
          text: 'Edit this page on GitHub',
        },
        lastUpdated: {
          text: 'Last updated',
        },
      },
    },
    zh: {
      label: '中文',
      lang: 'zh-CN',
      title: 'Pipeline',
      description: '强大的工作流执行引擎，支持本地执行和服务化部署',
      link: '/zh/',
      themeConfig: {
        nav: [
          { text: '首页', link: '/zh/' },
          { text: '快速开始', link: '/zh/guide/' },
          { text: '命令文档', link: '/zh/commands/' },
          { text: '架构设计', link: '/zh/architecture/' },
          { text: '最佳实践', link: '/zh/best-practices/' },
          { text: 'GitHub', link: 'https://github.com/go-idp/pipeline' },
        ],
        sidebar: {
          '/zh/guide/': [
            {
              text: '快速开始',
              items: [
                { text: '介绍', link: '/zh/guide/' },
                { text: '安装', link: '/zh/guide/installation' },
                { text: '第一个 Pipeline', link: '/zh/guide/getting-started' },
                { text: '配置文件', link: '/zh/guide/configuration' },
                { text: '核心概念', link: '/zh/guide/concepts' },
              ],
            },
          ],
          '/zh/commands/': [
            {
              text: '命令文档',
              items: [
                { text: '命令概述', link: '/zh/commands/' },
                { text: 'run 命令', link: '/zh/commands/run' },
                { text: 'server 命令', link: '/zh/commands/server' },
                { text: 'client 命令', link: '/zh/commands/client' },
              ],
            },
          ],
          '/zh/architecture/': [
            {
              text: '架构设计',
              items: [
                { text: '架构概述', link: '/zh/architecture/' },
                { text: '系统架构', link: '/zh/architecture/system' },
                { text: '错误处理', link: '/zh/architecture/error-handling' },
                { text: '性能优化', link: '/zh/architecture/optimization' },
              ],
            },
          ],
          '/zh/best-practices/': [
            {
              text: '最佳实践',
              items: [
                { text: '最佳实践', link: '/zh/best-practices/' },
              ],
            },
          ],
        },
        editLink: {
          pattern: 'https://github.com/go-idp/pipeline/edit/main/docs/:path',
          text: '在 GitHub 上编辑此页',
        },
        lastUpdated: {
          text: '最后更新',
        },
      },
    },
  },

  themeConfig: {
    logo: '/logo.png',
    socialLinks: [
      { icon: 'github', link: 'https://github.com/go-idp/pipeline' },
    ],
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024 Pipeline Team',
    },
    search: {
      provider: 'local',
    },
  },
})
