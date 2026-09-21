#!/usr/bin/env python3
"""Rebuild deterministic catalog, example anthology, grammar and single-file edition.
Canonical sources: spec/UICL-1.0.uicl, profiles/*.uicl, meta/grammar.uicl,
independent examples/*.uicl. Does not execute examples or rewrite their sources.
"""
from __future__ import annotations
import argparse, hashlib, json, re
from pathlib import Path
from syntax import read, walk, prop
from check import Catalog, PackageChecker
from grammar_views import derive
from host import extract
ROOT=Path(__file__).resolve().parents[1]

def q(x):return json.dumps(x,ensure_ascii=False)
def put(path,text,verify=False):
    if verify:
        if not path.exists() or path.read_text(encoding='utf-8')!=text:raise ValueError('STALE_DERIVED '+str(path.relative_to(ROOT)))
    else:path.write_text(text,encoding='utf-8')

def main():
    arg=argparse.ArgumentParser(description=__doc__);arg.add_argument('--check',action='store_true');args=arg.parse_args()
    cat=Catalog(ROOT);extra=Catalog(ROOT,[ROOT/'examples/profile-authoring/measurement-profile.uicl'])
    # Source titles and normative prose come from the hosted reading standard.
    source=(ROOT/'spec/UICL-1.0.uicl').read_text(encoding='utf-8')
    sections=list(re.finditer(r'^## (\d+)\. (.+)$',source,re.M));old=read(ROOT/'meta/core-rules.uicl');rules=[n for n in old['nodes'] if n['kind']=='rule']
    if len(sections)!=len(rules):raise ValueError('Core rule/source count mismatch: update stable rule ID mapping first')
    core='uicl "1.0"\n\ndocument #core_rules "UICL Core 规范性条款"\n  version: "1.0"\n  status: "consolidated-specification"\n  authority: "从 spec/UICL-1.0.uicl 派生的规则镜像，不替代完整领域验证。"\n'
    for i,(m,n) in enumerate(zip(sections,rules)):
        end=sections[i+1].start() if i+1<len(sections) else source.index('## 配套文件与使用示例',m.end())
        statement=m.group(2)+'\n\n'+source[m.end():end].strip()+'\n'
        core+='\nrule #'+n['id']+' '+q(n['label'])+'\n  level: '+q(prop(n,'level'))+'\n  verification: '+q(prop(n,'verification'))+'\n  origin: '+q(prop(n,'origin'))+'\n  statement: |\n'
        core+=''.join('    '+line+'\n' for line in statement.splitlines())
    put(ROOT/'meta/core-rules.uicl',core,args.check)
    put(ROOT/'spec/UICL-1.0.md',source,args.check)
    def closure(pids,c):
        out=set(pids)
        for pid in list(pids):
            for dep in c.profiles[pid]['requires']:out.update(closure({dep.rsplit('@',1)[0]},c))
        return out
    catalog='uicl "1.0"\n';entrydata=[]
    for i,path in enumerate(sorted((ROOT/'examples').rglob('*.uicl')),1):
        text=path.read_text(encoding='utf-8');c=extra if path.parent.name=='profile-authoring' else cat
        if text.startswith('<!--') or text.lstrip().lower().startswith('<!doctype'):
            ex=extract(text);nodes=[n for island in ex['islands'] for n in walk(island['ast']['nodes'])];pids={c.schemas[n['kind']]['profile'] for n in nodes}|{'uicl.document'}
        else:
            check=PackageChecker(ROOT,c);check.load(path)
            pids={c.schemas[n['kind']]['profile'] for mod in check.cache.values() for n in walk(mod['ast']['nodes'])}
        pids=closure(pids,c);ps=sorted(pid+'@'+c.profiles[pid]['version'] for pid in pids)
        rel=str(path.relative_to(ROOT));entrydata.append({'path':rel,'profiles':ps,'sha256':hashlib.sha256(text.encode()).hexdigest()})
        longest=max([len(m.group()) for m in re.finditer(r'`+',text)]+[2]);fence='`'*max(3,longest+1)
        catalog+='\nexample #example_'+str(i)+' '+q(path.stem)+'\n  path: '+q(rel)+'\n  profiles: '+q(ps)+'\n  readiness: "contract-example; tools-checked; real-runtime-not-executed"\n  body: '+fence+'uicl\n'
        catalog+=''.join('    '+line+'\n' for line in text.splitlines())+'  '+fence+'\n'
    put(ROOT/'meta/examples.uicl',catalog,args.check)
    put(ROOT/'generated/example-catalog.json',json.dumps(entrydata,ensure_ascii=False,indent=2)+'\n',args.check)
    graph,ebnf=derive(read(ROOT/'meta/grammar.uicl'))
    put(ROOT/'generated/grammar-tree.json',json.dumps(graph,ensure_ascii=False,indent=2)+'\n',args.check);put(ROOT/'generated/core.ebnf',ebnf,args.check)
    # All explicit identities are module-unique; no executable example is inlined as active nodes.
    sources=[ROOT/'meta/core-rules.uicl',ROOT/'meta/grammar.uicl']+sorted((ROOT/'profiles').glob('*.uicl'))+[ROOT/'meta/examples.uicl']
    book='uicl "1.0"\n\n// Generated structured edition. Edit canonical sources, then rebuild.\n'
    for path in sources:
        text=core if path.name=='core-rules.uicl' else catalog if path.name=='examples.uicl' else path.read_text(encoding='utf-8')
        book+='\n// source: '+str(path.relative_to(ROOT))+'\n'+text.split('\n',1)[1].strip()+'\n'
    put(ROOT/'UICL-1.0-Structured.uicl',book,args.check)
    print(json.dumps({'status':'PASS','mode':'check' if args.check else 'rebuild','core_rules':len(rules),'grammar_rules':len(graph['rules']),'examples':len(entrydata)},ensure_ascii=False))
if __name__=='__main__':main()
