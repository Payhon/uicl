#!/usr/bin/env python3
"""Limited native-host example checker, not a complete WHATWG/CST editor.
Markdown extraction uses markdown-it-py in CommonMark mode. HTML extraction uses
Python HTMLParser and is deliberately restricted to ordinary well-formed examples.
No scripts, model calls, resource loads, or document writes are performed.
"""
from __future__ import annotations
import argparse, base64, json, re, sys, textwrap
from collections import Counter
from html.parser import HTMLParser
from pathlib import Path
from syntax import parse, walk, prop
from check import Catalog

class HostError(ValueError): pass

HEADER_MD='<!-- uicl:markdown 1.0 -->'
HEADER_HTML='uicl:html 1.0'

def decode_payload(text):
    compact=re.sub(r'[ \t\r\n]', '', text)
    if len(compact)>16*1024*1024:raise HostError('HOST_LIMIT')
    if not re.fullmatch(r'[A-Za-z0-9_-]*',compact):raise HostError('HOST_ENCODING')
    try:
        b=base64.urlsafe_b64decode(compact+'='*((-len(compact))%4))
        if base64.urlsafe_b64encode(b).decode().rstrip('=')!=compact:raise HostError('HOST_NONCANONICAL')
        return b.decode('utf-8')
    except (ValueError,UnicodeError) as exc:raise HostError('HOST_ENCODING') from exc

def annotation(data):
    """Return a fragment, None for ordinary comments. Header is handled separately."""
    data=textwrap.dedent(data).strip()
    if data.startswith('uicl:base64url\n'):return decode_payload(data.split('\n',1)[1])
    if data=='uicl' or data.startswith('uicl\n'):
        value=data[4:].lstrip('\r\n')
        if '<!--' in value or '-->' in value:raise HostError('HOST_COMMENT_BOUNDARY')
        return textwrap.dedent(value)
    return None

def fragment(text,complete=False):
    return parse(text if complete else 'uicl "1.0"\n'+text+'\n')

def markdown(source):
    try:from markdown_it import MarkdownIt
    except ImportError as exc:raise HostError('markdown-it-py is required for Markdown host checks') from exc
    source=source.removeprefix('\ufeff')
    if not source.startswith(HEADER_MD) or source[len(HEADER_MD):len(HEADER_MD)+1] not in ('\n','\r',''):
        raise HostError('HOST_MARKER_REQUIRED')
    engine=MarkdownIt('commonmark',{'html':True});ts=engine.parse(source);islands=[];headers=0
    for i,t in enumerate(ts):
        if t.type!='html_block' or t.level!=0:continue
        raw=t.content.strip()
        if raw==HEADER_MD:headers+=1;continue
        if raw.startswith('<!-- uicl:') and not raw.startswith('<!-- uicl:base64url\n'):
            raise HostError('HOST_MODE_OR_VERSION_CONFLICT')
        if not (raw.startswith('<!-- uicl\n') or raw.startswith('<!-- uicl:base64url\n')):continue
        if not raw.endswith('-->') or raw.count('<!--')!=1 or raw.count('-->')!=1:
            raise HostError('HOST_COMMENT_BOUNDARY')
        # Source column zero is mandatory for the first Markdown-host subset.
        if t.map and source.splitlines()[t.map[0]].startswith(' '):raise HostError('HOST_ANNOTATION_COLUMN')
        text=annotation(raw[4:-3]);ast=fragment(text);next_block=None
        for following in ts[i+1:]:
            if following.level!=0 or following.nesting==-1:continue
            if following.type=='html_block' and following.content.lstrip().startswith('<!-- uicl'):continue
            next_block=following;break
        for n in walk(ast['nodes']):
            if n['kind']=='doc.block':
                if prop(n,'target')!='next':raise HostError('MARKDOWN_TARGET_REQUIRES_NEXT')
                if next_block is None:raise HostError('HOST_TARGET_MISSING')
        islands.append({'line':t.map[0]+1 if t.map else None,'ast':ast,'target_line':next_block.map[0]+1 if next_block and next_block.map else None})
    if headers!=1:raise HostError('HOST_DUPLICATE_MARKER')
    return {'mode':'markdown','islands':islands,'rendered':engine.render(source),'host_engine':'markdown-it-py/CommonMark'}

class HTMLExtract(HTMLParser):
    BLOCKED={'pre','code','textarea','template','script','style'}
    VOID={'area','base','br','col','embed','hr','img','input','link','meta','param','source','track','wbr'}
    def __init__(self):
        super().__init__(convert_charrefs=False);self.stack=[];self.headers=0;self.islands=[];self.ids=[];self.capture=None
    def handle_starttag(self,tag,attrs):
        attrs=dict(attrs);blocked=any(x in self.BLOCKED for x in self.stack)
        if 'id' in attrs and not blocked:self.ids.append(attrs['id'])
        if tag=='script' and not blocked and ('data-uicl' in attrs or 'data-uicl-encoding' in attrs):
            if attrs.get('type','').lower()!='text/plain':raise HostError('HOST_SCRIPT_TYPE')
            enc=attrs.get('data-uicl-encoding')
            if enc not in (None,'base64url'):raise HostError('HOST_ENCODING')
            self.capture={'line':self.getpos()[0],'encoding':enc,'data':[]}
        if tag not in self.VOID:self.stack.append(tag)
    def handle_startendtag(self,tag,attrs):
        self.handle_starttag(tag,attrs)
        if tag not in self.VOID:self.handle_endtag(tag)
    def handle_endtag(self,tag):
        if tag=='script' and self.capture is not None:
            cap=self.capture;self.capture=None;body=''.join(cap['data'])
            if cap['encoding']:body=decode_payload(body)
            self.islands.append({'line':cap['line'],'ast':fragment(textwrap.dedent(body).strip(),complete=True)})
        if tag in self.stack:
            idx=len(self.stack)-1-self.stack[::-1].index(tag);self.stack=self.stack[:idx]
    def handle_data(self,data):
        if self.capture is not None:self.capture['data'].append(data)
    def handle_comment(self,data):
        if self.capture is not None:self.capture['data'].append('<!--'+data+'-->');return
        if any(x in self.BLOCKED for x in self.stack):return
        clean=data.strip()
        if clean==HEADER_HTML:
            if 'head' not in self.stack:raise HostError('HOST_MARKER_OUTSIDE_HEAD')
            self.headers+=1;return
        if clean.startswith('uicl:') and not clean.startswith('uicl:base64url\n'):raise HostError('HOST_MODE_OR_VERSION_CONFLICT')
        body=annotation(data)
        if body is not None:self.islands.append({'line':self.getpos()[0],'ast':fragment(body)})

def html(source):
    p=HTMLExtract();p.feed(source);p.close()
    if p.capture is not None:raise HostError('HOST_UNCLOSED_DATA_BLOCK')
    if p.headers!=1:raise HostError('HOST_MARKER_COUNT')
    ids=Counter(p.ids)
    bound=set()
    for island in p.islands:
        for n in walk(island['ast']['nodes']):
            if n['kind']!='doc.block':continue
            target=prop(n,'target')
            if not isinstance(target,dict) or set(target)!={'html_id'}:raise HostError('HTML_TARGET_REQUIRES_ID')
            key=target['html_id']
            if ids[key]!=1:raise HostError('HOST_TARGET_MISSING_OR_DUPLICATE')
            if key in bound:raise HostError('HOST_TARGET_CONFLICT')
            bound.add(key)
    return {'mode':'html','islands':p.islands,'html_ids':p.ids,'host_engine':'Python HTMLParser (restricted examples; not full WHATWG)'}

def extract(source,mode=None):
    if len(source.encode('utf-8'))>16*1024*1024:raise HostError('HOST_LIMIT')
    if mode is None:mode='markdown' if source.removeprefix('\ufeff').startswith(HEADER_MD) else 'html'
    if mode not in ('markdown','html'):raise HostError('HOST_MODE')
    return markdown(source) if mode=='markdown' else html(source)

def validate(path,cat):
    original=path.read_bytes();result=extract(original.decode('utf-8'));nodes=[]
    for island in result['islands']:nodes.extend(island['ast']['nodes'])
    ast={'language':'uicl','version':'1.0','nodes':nodes}
    if nodes:cat.normalize(ast)
    result.pop('rendered',None)
    return {'file':str(path.relative_to(cat.root)),'mode':result['mode'],'islands':len(result['islands']),'nodes':len(list(walk(nodes))),'status':'PASS','scope':'sample host extraction + shape, not full host conformance or runtime'}

def main():
    cli=argparse.ArgumentParser(description=__doc__);cli.add_argument('files',nargs='*',type=Path);cli.add_argument('--out',type=Path);args=cli.parse_args()
    root=Path(__file__).resolve().parents[1];cat=Catalog(root)
    paths=args.files or sorted((root/'spec').glob('*.uicl'))+sorted((root/'guides').glob('*.uicl'))+sorted((root/'examples/hosted').glob('*.uicl'))
    if not args.files and (root/'README.uicl').exists():paths.append(root/'README.uicl')
    # spec/layout and other helper files are hosted; ignore an optional structural file.
    paths=[p for p in paths if not p.read_text(encoding='utf-8-sig').startswith('uicl "')]
    try:
        results=[validate(p.resolve(),cat) for p in paths]
        report={'status':'PASS','documents':len(results),'results':results,'not_evaluated':['all CommonMark/GFM/WHATWG cases','browser rendering and script behavior','lossless edits','annotation execution authorization']}
        output=json.dumps(report,ensure_ascii=False,indent=2)+'\n'
        if args.out:args.out.write_text(output,encoding='utf-8')
        else:print(output,end='')
    except (ValueError,OSError) as exc:print(exc,file=sys.stderr);return 1
    return 0
if __name__=='__main__':raise SystemExit(main())
