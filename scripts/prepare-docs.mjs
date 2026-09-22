import assert from 'node:assert/strict';
import { existsSync, mkdirSync, readFileSync, readdirSync, rmSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('../', import.meta.url));
const github = 'https://github.com/Payhon/uicl/blob/main/';

export function filesIn(directory) {
  if (!existsSync(directory)) return [];
  return readdirSync(directory, { withFileTypes: true }).flatMap(entry => {
    const filename = path.join(directory, entry.name);
    return entry.isDirectory() ? filesIn(filename) : [filename];
  }).sort();
}

export function renderMarkdown(text, source, page, pages) {
  let fence;
  // Reading copies omit host metadata; code examples retain their original bytes.
  return text.replace(/^(?:<!-- uicl[\s\S]*?-->\s*)+/, '').split('\n').map(line => {
    const marker = line.match(/^ {0,3}(`{3,}|~{3,})(.*)$/);
    if (fence) {
      if (marker && marker[1][0] === fence[0] && marker[1].length >= fence.length && !marker[2].trim()) fence = undefined;
      return line;
    }
    if (marker) {
      fence = marker[1];
      return line.replace(/^( {0,3}(?:`{3,}|~{3,}))(uicl|aml)(?=\s|$)/, '$1text');
    }
    // ponytail: current prose uses single-line code spans; adopt a Markdown AST if richer source syntax is added.
    return line.split(/(`+[^`]*`+)/g).map((part, index) => {
      if (index % 2) return part;
      return part.replace(/\]\(([^\s)]+)([^)]*)\)/g, (match, href, title) => {
        if (/^(?:[a-z][\w+.-]*:|\/|#)/i.test(href)) return match;
        const [, destination, suffix] = href.match(/^([^?#]+)(.*)$/);
        const target = path.posix.normalize(path.posix.join(path.posix.dirname(source), decodeURIComponent(destination)));
        const mapped = pages.get(target);
        if (mapped) return `](${path.posix.relative(path.posix.dirname(page), mapped)}${suffix}${title})`;
        assert(!target.startsWith('../') && existsSync(path.join(root, target)), `Broken source link: ${source} -> ${href}`);
        return `](${github}${target.split('/').map(encodeURIComponent).join('/')}${suffix}${title})`;
      }).replace(/</g, '&lt;').replace(/\{/g, '&#123;').replace(/\}/g, '&#125;');
    }).join('');
  }).join('\n');
}

export function prepareDocs(output = path.join(root, 'website/docs')) {
  const result = new Map();
  const sources = ['README.zh-CN.md', ...['spec', 'guides'].flatMap(directory =>
    filesIn(path.join(root, directory)).filter(file => file.endsWith('.md')).map(file => path.relative(root, file)))];
  const examples = filesIn(path.join(root, 'examples')).filter(file => file.endsWith('.uicl')).map(file => path.relative(root, file));
  const pages = new Map(sources.map(source => [source, source === 'README.zh-CN.md' ? 'overview.md' : source]));
  for (const [source, page] of [...pages]) pages.set(source === 'README.zh-CN.md' ? 'README.uicl' : source.replace(/\.md$/, '.uicl'), page);
  pages.set('README.md', 'overview.md');
  for (const source of examples) pages.set(source, source.replace(/\.uicl$/, '.md'));

  for (const source of sources) {
    const text = readFileSync(path.join(root, source), 'utf8');
    const canonical = source === 'README.zh-CN.md' ? 'README.uicl' : source.replace(/\.md$/, '.uicl');
    assert.equal(text, readFileSync(path.join(root, canonical), 'utf8'), `Stale reading copy: ${source}`);
    result.set(pages.get(source), `${renderMarkdown(text, source, pages.get(source), pages).trimEnd()}\n\n---\n\n原始文档：[${canonical}](${github}${canonical})\n`);
  }

  let index = '# 示例库\n\n从实际 `.uicl` 文件生成的阅读目录。示例描述语言契约，当前工具只提供有限解析与静态检查；本页不会运行 UI、数据库、模型或部署任务。\n\n';
  index += '建议从 [Hello](01-hello.md)、[计数器](02-counter.md) 开始，再阅读 [全流程应用说明](../spec/Fullstack-Walkthrough.md)。\n\n';
  const groups = new Map([
    ['examples', '语言与领域示例'], ['examples/fullstack', '完整应用契约'],
    ['examples/hosted', 'Markdown 与 HTML 承载'], ['examples/profile-authoring', '第三方 Profile'],
  ]);
  for (const directory of [...new Set(examples.map(source => path.posix.dirname(source)))]) {
    index += `## ${groups.get(directory) || directory}\n\n`;
    for (const source of examples.filter(source => path.posix.dirname(source) === directory)) {
      const name = path.posix.basename(source, '.uicl');
      const text = readFileSync(path.join(root, source), 'utf8');
      const fence = '`'.repeat(Math.max(3, ...[...text.matchAll(/`+/g)].map(match => match[0].length + 1)));
      result.set(pages.get(source), `# ${name}\n\n这是 UICL 契约源码阅读页，不代表已实现或执行相应运行时。\n\n[查看 GitHub 原始源码](${github}${source}) · [返回示例库](${path.posix.relative(directory, 'examples/index.md')})\n\n${fence}text\n${text}${text.endsWith('\n') ? '' : '\n'}${fence}\n`);
      index += `- [${name}](${path.posix.relative('examples', pages.get(source))})\n`;
    }
    index += '\n';
  }
  result.set('examples/index.md', index);

  const authored = path.join(root, 'website/pages');
  for (const file of filesIn(authored)) {
    const page = path.relative(authored, file);
    assert(!result.has(page), `Authored page conflicts with generated page: ${page}`);
    result.set(page, readFileSync(file));
  }
  for (const file of filesIn(path.join(root, 'website/public'))) {
    result.set(path.join('public', path.relative(path.join(root, 'website/public'), file)), readFileSync(file));
  }

  const link = (file, label) => {
    assert(result.has(file), `Missing sidebar page: ${file}`);
    return {
      text: label || result.get(file).toString().match(/^# (.+)$/m)?.[1].replace(/^UICL /, '').replace(/ Profile 1\.0$/, ''),
      link: `/${file.replace(/\.md$/, '').replace(/\/index$/, '/')}`,
    };
  };
  const sidebar = [
    { text: '开始', items: [link('overview.md', '项目总览'), link('spec/Quickstart.md', '快速入门'), link('tooling.md', '工具与验证')] },
    { text: '语言规范', collapsible: true, items: [link('spec/UICL-1.0.md', '语言标准 1.0'), ...sources.filter(source => source.startsWith('spec/') && !['spec/UICL-1.0.md', 'spec/Quickstart.md', 'spec/Profile-Index.md'].includes(source)).map(source => link(source))] },
    { text: '领域 Profile', collapsible: true, collapsed: true, items: [link('spec/Profile-Index.md', 'Profile 索引'), ...sources.filter(source => source.startsWith('guides/')).map(source => link(source))] },
    { text: '示例', collapsible: true, collapsed: true, items: [link('examples/index.md', '浏览全部示例'), ...[...new Set(examples.map(source => path.posix.dirname(source)))].map(directory => ({
      text: groups.get(directory) || directory, collapsible: true, collapsed: true,
      items: examples.filter(source => path.posix.dirname(source) === directory).map(source => link(pages.get(source))),
    }))] },
  ];
  // Only this generated directory is replaced; canonical sources remain untouched.
  rmSync(output, { recursive: true, force: true });
  for (const [page, content] of result) {
    const target = path.join(output, page);
    mkdirSync(path.dirname(target), { recursive: true });
    writeFileSync(target, content);
  }
  return { pages: result, examples, sidebar };
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const { pages, examples, sidebar } = prepareDocs();
  writeFileSync(path.join(root, 'website/sidebar.json'), `${JSON.stringify(sidebar, null, 2)}\n`);
  console.log(`Prepared ${pages.size} documentation files, including ${examples.length} UICL examples.`);
}
