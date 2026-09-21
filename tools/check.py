#!/usr/bin/env python3
"""Read canonical UICL Profiles and validate structural contracts.
Checks shape, required/unknown properties, primary slots, child kinds, module exports,
reference targets and declared ports. Selected cross-field constraints are checked.
Does NOT check full runtime binding types, SQL semantics, effects, authorization, or execute anything.
Reads only explicit local sources within the selected package root; no network or secret resolution.
"""
from __future__ import annotations
import argparse, copy, fnmatch, hashlib, json, sys
from pathlib import Path
from urllib.parse import urlsplit, unquote
from syntax import read, parse, walk, prop, plain

class CheckError(ValueError):pass

def split_top(s,delimiter='|'):
    out=[];start=0;depth=0
    for i,c in enumerate(s):
        if c=='<':depth+=1
        elif c=='>':depth-=1
        elif c==delimiter and depth==0:out.append(s[start:i]);start=i+1
    out.append(s[start:]);return out

def refs(v):
    if isinstance(v,dict):
        if v.get('tag')=='ref':yield v
        for x in v.values():yield from refs(x)
    elif isinstance(v,list):
        for x in v:yield from refs(x)

def is_expr(v):return v.get('tag')=='expression'

class Catalog:
    def __init__(self,root:Path,extra: list[Path]|None=None):
        root=root.resolve();self.root=root;self.schemas={};self.records={};self.profiles={}
        paths=sorted((root/'profiles').glob('*.uicl'))+(extra or [])
        for path in paths:
            path=path.resolve()
            if root not in path.parents:raise CheckError('PROFILE_PATH_ESCAPE '+str(path))
            ast=read(path)
            for p in ast['nodes']:
                if p['kind']!='profile':continue
                pid=prop(p,'id')
                if pid in self.profiles:raise CheckError('PROFILE_CONFLICT '+str(pid))
                self.profiles[pid]={'path':str(path.relative_to(root)),'requires':prop(p,'requires',[]),'version':prop(p,'version')}
                for n in p['children']:
                    if n['kind'] not in ('schema','record_schema'):continue
                    key=n['label'] or prop(n,'kind' if n['kind']=='schema' else 'name')
                    dest=self.schemas if n['kind']=='schema' else self.records
                    if key in dest:raise CheckError('SCHEMA_CONFLICT '+key)
                    fields={};ports={}
                    for child in n['children']:
                        if child['kind']=='property':
                            pname=child['label'] or prop(child,'name')
                            if pname in fields:raise CheckError('DUPLICATE_SCHEMA_PROPERTY '+pname)
                            fields[pname]={k:plain(v) for k,v in child['props'].items()}
                        elif child['kind']=='port':ports[child['label'] or prop(child,'name')]=prop(child,'type')
                    dest[key]={'fields':fields,'primary':prop(n,'primary'),'identity':prop(n,'identity'),'children':prop(n,'children',[]),'ports':ports,'profile':pid}
        for pid,p in self.profiles.items():
            for dep in p['requires']:
                name,version=dep.rsplit('@',1)
                if name not in self.profiles or self.profiles[name]['version']!=version:raise CheckError('PROFILE_DEPENDENCY '+dep)
        visiting=set();done=set()
        def visit(name):
            if name in visiting:raise CheckError('PROFILE_CYCLE '+name)
            if name in done:return
            visiting.add(name)
            for dep in self.profiles[name]['requires']:visit(dep.rsplit('@',1)[0])
            visiting.remove(name);done.add(name)
        for name in self.profiles:visit(name)
    def type_ok(self,v,t):
        branches=split_top(t)
        if len(branches)>1:return any(self.type_ok(v,b) for b in branches)
        if t=='value':return True
        if t=='ref':return v['tag']=='ref'
        if t=='text':return v['tag']=='raw' or v['tag']=='literal' and v['type']=='text'
        if t=='number':return v['tag']=='literal' and v['type'] in ('int','decimal')
        if t in ('int','decimal','bool','null'):return v['tag']=='literal' and v['type']==t
        if t=='record':return v['tag']=='record'
        if t.startswith('list<') and t.endswith('>'):
            return v['tag']=='list' and all(self.type_ok(x,t[5:-1]) for x in v['items'])
        if t.startswith('record<') and t.endswith('>'):
            if v['tag']!='record':return False
            name=t[7:-1]
            if name not in self.records:raise CheckError('UNKNOWN_RECORD_TYPE '+name)
            try:self.fields(v['fields'],self.records[name]['fields'],name)
            except CheckError:return False
            return True
        raise CheckError('UNKNOWN_SHAPE_TYPE '+t)
    def fields(self,actual,expected,path):
        for k in actual:
            if k not in expected:raise CheckError('UNKNOWN_PROPERTY '+path+'.'+k)
        for k,p in expected.items():
            if k not in actual:
                if p.get('required') and 'default' not in p:raise CheckError('MISSING_PROPERTY '+path+'.'+k)
                continue
            v=actual[k]
            if is_expr(v):
                if not p.get('expression'):raise CheckError('EXPRESSION_FORBIDDEN '+path+'.'+k)
                continue
            if not self.type_ok(v,p['type']):raise CheckError('TYPE_SHAPE '+path+'.'+k+' requires '+p['type'])
            if 'enum' in p and plain(v) not in p['enum']:raise CheckError('ENUM '+path+'.'+k)
    def normalize(self,ast):
        ast=copy.deepcopy(ast)
        ids={}
        for n in walk(ast['nodes']):
            kind=n['kind'];s=self.schemas.get(kind)
            if not s:raise CheckError('UNKNOWN_NODE '+kind)
            if n['id']:
                if n['id'] in ids:raise CheckError('DUPLICATE_ID '+n['id'])
                ids[n['id']]=n
            if s['identity']=='required' and not n['id']:raise CheckError('MISSING_ID '+kind)
            if s['identity']=='forbidden' and n['id']:raise CheckError('FORBIDDEN_ID '+kind)
            if n['label'] is not None:
                primary=s['primary']
                if not primary:raise CheckError('PRIMARY_FORBIDDEN '+kind)
                if primary in n['props']:raise CheckError('PRIMARY_CONFLICT '+kind+'.'+primary)
                n['props'][primary]={'tag':'literal','type':'text','value':n['label']}
                n['label']=None
            self.fields(n['props'],s['fields'],kind)
            for ch in n['children']:
                if not any(fnmatch.fnmatchcase(ch['kind'],pattern) for pattern in s['children']):raise CheckError('CHILD_FORBIDDEN '+kind+' -> '+ch['kind'])
            self.selected(n)
        return ast,ids
    def selected(self,n):
        kind=n['kind'];p=n['props']
        def xor(a,b,required=False):
            count=int(a in p)+int(b in p)
            if count>1 or required and count!=1:raise CheckError('MUTUALLY_EXCLUSIVE '+kind+': '+a+'/'+b)
        if kind=='file':xor('source','body',True)
        if kind=='app':xor('target','targets')
        if kind=='endpoint':xor('action','query',True)
        if kind=='type':
            if bool(prop(n,'enum'))==bool(n['children']):raise CheckError('TYPE_RECORD_OR_ENUM '+str(n['id']))
        if kind=='collection' and prop(n,'storage')=='server' and 'access' not in p and not any(c['kind']=='policy' for c in n['children']):raise CheckError('SERVER_ACCESS_REQUIRED')
        if kind=='g.repeat':
            lo,hi=prop(n,'min'),prop(n,'max')
            if lo<0 or hi is not None and hi<lo or len(n['children'])!=1:raise CheckError('GRAMMAR_REPEAT_BOUNDS')
        if kind in ('ui.image',):pass
        if kind in ('image','video','design.canvas'):
            size=prop(n,'size')
            if size is not None and (len(size)!=2 or any(not isinstance(x,int) or isinstance(x,bool) or x<=0 for x in size)):raise CheckError('DIMENSION_PAIR')
        if kind=='video':
            fps=prop(n,'fps')
            if fps['num']<=0 or fps['den']<=0 or prop(n,'frames')<=0:raise CheckError('VIDEO_TIMEBASE')
            for track in n['children']:
                if track['kind']!='track':continue
                segments=[]
                for shot in track['children']:
                    start=prop(shot,'start_frame');length=prop(shot,'frames')
                    if start<0 or length<=0 or start+length>prop(n,'frames'):raise CheckError('TIMELINE_BOUNDS')
                    segments.append((start,start+length))
                segments.sort()
                if prop(track,'overlap','forbid')=='forbid' and any(a[1]>b[0] for a,b in zip(segments,segments[1:])):raise CheckError('TIMELINE_OVERLAP')
        if kind=='pg.table':
            cols=prop(n,'columns');names=[c['name'] for c in cols]
            if len(set(names))!=len(names):raise CheckError('DUPLICATE_COLUMN')
            for c in cols:
                if 'default' in c and 'default_sql' in c:raise CheckError('PG_DEFAULT_CONFLICT')
            if not set(prop(n,'primary_key',[]))<=set(names):raise CheckError('PG_UNKNOWN_KEY_COLUMN')
            for entry in prop(n,'indexes',[])+prop(n,'unique',[])+prop(n,'foreign_keys',[]):
                if not set(entry['columns'])<=set(names):raise CheckError('PG_UNKNOWN_COLUMN')
        if kind=='pg.migration' and prop(n,'atomic')=='required':
            if any(s.get('build')=='concurrently' for s in prop(n,'steps')):raise CheckError('PG_TRANSACTION_CONFLICT')
        if kind=='secret' and not prop(n,'scope'):raise CheckError('SECRET_EMPTY_SCOPE')

class PackageChecker:
    def __init__(self,root:Path,catalog:Catalog):self.root=root.resolve();self.catalog=catalog;self.cache={};self.visiting=set();self.deferred=[];self.ref_count=0
    def local_uri(self,origin:Path,ref):
        uri=ref.get('uri')
        if uri is None:raise CheckError('IMPORT_REQUIRES_URI')
        u=urlsplit(uri)
        if u.scheme or u.netloc or u.query or u.fragment:raise CheckError('IMPORT_NONLOCAL_NOT_SUPPORTED '+uri)
        p=(origin.parent/unquote(u.path)).resolve()
        if p!=self.root and self.root not in p.parents:raise CheckError('PATH_ESCAPE '+uri)
        return p
    def load(self,path):
        path=path.resolve()
        if path in self.cache:return self.cache[path]
        if path in self.visiting:raise CheckError('IMPORT_CYCLE '+str(path))
        self.visiting.add(path)
        ast,ids=self.catalog.normalize(read(path)); imports={};modules=[n for n in ast['nodes'] if n['kind']=='module']
        if len(modules)>1:raise CheckError('DUPLICATE_MODULE')
        exports=set(prop(modules[0],'exports',[])) if modules else set()
        if not exports<=set(ids):raise CheckError('UNKNOWN_EXPORT '+str(exports-set(ids)))
        for n in ast['nodes']:
            if n['kind']=='use':imports[n['id']]=self.load(self.local_uri(path,n['props']['source']))
        info={'path':path,'ast':ast,'ids':ids,'imports':imports,'exports':exports}
        for n in walk(ast['nodes']):
            schema=self.catalog.schemas[n['kind']]
            for key,v in n['props'].items():
                targets=schema['fields'][key].get('targets',[])
                for r in refs(v):
                    self.ref_count+=1
                    if 'uri' in r:
                        # URI denotes identity/location only. Non-imports are intentionally NOT read.
                        if n['kind']!='use':self.deferred.append({'file':str(path.relative_to(self.root)),'node':n['id'],'uri':r['uri'],'status':'NOT_RESOLVED_BY_CHECKER'})
                        continue
                    dest=info
                    if r.get('module'):
                        alias=r['module']
                        if alias not in imports:raise CheckError('UNKNOWN_IMPORT '+alias)
                        dest=imports[alias]
                        if r['id'] not in dest['exports']:raise CheckError('NOT_EXPORTED '+r['id'])
                    if r['id'] not in dest['ids']:raise CheckError('UNRESOLVED_REF '+r['id']+' in '+str(path.relative_to(self.root)))
                    target=dest['ids'][r['id']]
                    if r.get('ports'):
                        if len(r['ports'])!=1 or r['ports'][0] not in self.catalog.schemas[target['kind']]['ports']:raise CheckError('UNKNOWN_PORT '+str(r))
                    elif targets and target['kind'] not in targets:raise CheckError('REF_TARGET_KIND '+n['kind']+'.'+key+' -> '+target['kind'])
        self.cache[path]=info;self.visiting.remove(path);return info

def main():
    cli=argparse.ArgumentParser(description=__doc__);cli.add_argument('files',nargs='*',type=Path);cli.add_argument('--root',type=Path,default=Path(__file__).resolve().parents[1]);cli.add_argument('--extra-profile',action='append',type=Path,default=[]);cli.add_argument('--out',type=Path)
    args=cli.parse_args()
    try:
        cat=Catalog(args.root,args.extra_profile);checker=PackageChecker(args.root,cat)
        paths=args.files or sorted((args.root/'profiles').glob('*.uicl'))+sorted((args.root/'meta').glob('*.uicl'))+sorted((args.root/'examples').glob('*.uicl'))+sorted((args.root/'examples/fullstack').glob('*.uicl'))
        for path in paths:checker.load(path)
        report={'status':'PASS','scope':'syntax, registered shapes, primary slots, module exports, reference identity/port existence, selected static rules','documents':len(checker.cache),'node_schemas':len(cat.schemas),'record_schemas':len(cat.records),'profiles':len(cat.profiles),'references_checked':checker.ref_count,'external_resources':checker.deferred,'not_evaluated':['runtime binding type system','generic port type compatibility','permission/effect soundness','domain runtimes','real build/deployment','full host-document conformance','complete grammar-driven independent parser']}
        output=json.dumps(report,ensure_ascii=False,indent=2)+'\n'
        if args.out:args.out.write_text(output,encoding='utf-8')
        else:print(output,end='')
        return 0
    except (ValueError,OSError,RecursionError) as exc:print(exc,file=sys.stderr);return 1
if __name__=='__main__':raise SystemExit(main())
