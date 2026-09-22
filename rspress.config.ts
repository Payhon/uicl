import { defineConfig } from '@rspress/core';
import sidebar from './website/sidebar.json';

export default defineConfig({
  root: 'website/docs',
  base: '/uicl/',
  siteOrigin: 'https://payhon.github.io',
  title: 'UICL',
  description: 'UICL 通用意图与契约语言：语言规范、领域 Profile、示例与参考检查工具。',
  lang: 'zh',
  logo: '/logo.svg',
  logoText: 'UICL 1.0',
  icon: '/logo.svg',
  markdown: {
    link: { checkDeadLinks: true, checkAnchors: true },
  },
  themeConfig: {
    sidebar: { '/': sidebar },
    darkMode: 'light',
    nav: [
      { text: '开始', link: '/spec/Quickstart', activeMatch: '^/(overview|tooling|spec/Quickstart)' },
      { text: '语言规范', link: '/spec/UICL-1.0' },
      { text: 'Profiles', link: '/spec/Profile-Index', activeMatch: '^/(guides|spec/Profile-)' },
      { text: '示例', link: '/examples/', activeMatch: '^/examples/' },
    ],
    socialLinks: [{ icon: 'github', mode: 'link', content: 'https://github.com/Payhon/uicl' }],
    footer: {
      message: 'UICL · Universal Intent and Contract Language<br />规范、示例与工具的实现范围，以文档及实际检查结果为准。',
    },
  },
});
