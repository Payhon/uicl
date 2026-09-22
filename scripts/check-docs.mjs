import assert from 'node:assert/strict';
import { mkdtempSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { prepareDocs, renderMarkdown } from './prepare-docs.mjs';

const root = fileURLToPath(new URL('../', import.meta.url));
const sample = '<!-- uicl:markdown 1.0 -->\n\n# 标题\n\n<script> Artifact<ui.tree> {html_id:1} `list<T>`\n\n[入门](Quickstart.uicl#开始) [源码](../profiles/00-meta.uicl)\n\n```uicl\n<!-- uicl -->\nvalue: {a: 1}\n[保留](Quickstart.uicl)\n```\n';
const rendered = renderMarkdown(sample, 'spec/UICL-1.0.md', 'spec/UICL-1.0.md', new Map([['spec/Quickstart.uicl', 'spec/Quickstart.md']]));
assert(rendered.startsWith('# 标题'));
assert(rendered.includes('&lt;script> Artifact&lt;ui.tree> &#123;html_id:1&#125; `list<T>`'));
assert(rendered.includes('[入门](Quickstart.md#开始)'));
assert(rendered.includes('https://github.com/Payhon/uicl/blob/main/profiles/00-meta.uicl'));
assert(rendered.includes('```text\n<!-- uicl -->\nvalue: {a: 1}\n[保留](Quickstart.uicl)\n```'));
assert.throws(() => renderMarkdown('[missing](missing.uicl)', 'spec/test.md', 'spec/test.md', new Map()), /Broken source link/);

const output = mkdtempSync(path.join(tmpdir(), 'uicl-docs-'));
try {
  const { pages, examples, sidebar } = prepareDocs(output);
  assert(pages.has('overview.md'));
  assert(pages.has('spec/UICL-1.0.md'));
  assert(pages.has('guides/19-lifecycle.md'));
  assert(pages.get('overview.md').endsWith('原始文档：[README.uicl](https://github.com/Payhon/uicl/blob/main/README.uicl)\n'));
  assert(pages.get('spec/Quickstart.md').includes('原始文档：[spec/Quickstart.uicl]'));
  assert(pages.get('guides/01-core.md').includes('原始文档：[guides/01-core.uicl]'));
  assert.deepEqual(sidebar.map(group => group.text), ['开始', '语言规范', '领域 Profile', '示例']);
  for (const source of examples) {
    const page = pages.get(source.replace(/\.uicl$/, '.md'));
    const text = readFileSync(path.join(root, source), 'utf8');
    const fence = page.match(/^(`{3,})text$/m)?.[1];
    assert(fence, `Missing text fence: ${source}`);
    assert(page.includes(`${fence}text\n${text}${text.endsWith('\n') ? '' : '\n'}${fence}\n`), `Changed example content: ${source}`);
    assert(pages.get('examples/index.md').includes(source.slice('examples/'.length).replace(/\.uicl$/, '.md')));
  }
  // Validate generated local Markdown links, excluding code examples and external URLs.
  for (const [file, content] of pages) {
    if (!file.endsWith('.md') || Buffer.isBuffer(content)) continue;
    const prose = content.replace(/^(`{3,}|~{3,})[^\n]*\n[\s\S]*?^\1\s*$/gm, '');
    for (const [, href] of prose.matchAll(/\]\(([^\s)]+)\)/g)) {
      if (/^(?:[a-z][\w+.-]*:|\/|#)/i.test(href)) continue;
      const target = path.posix.normalize(path.posix.join(path.posix.dirname(file), href.split(/[?#]/)[0]));
      assert(pages.has(target), `Broken page link: ${file} -> ${href}`);
    }
  }
  console.log(`Documentation checks passed: ${pages.size} files, ${examples.length} complete source examples.`);
} finally {
  rmSync(output, { recursive: true, force: true });
}
