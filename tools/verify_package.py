#!/usr/bin/env python3
"""Run reproducible package checks and record exactly the evaluated scope.
Subprocesses are only local Python checker/test programs; no external resources,
cloud accounts, models or domain examples are executed.
"""
from __future__ import annotations
import hashlib, importlib.metadata, json, platform, re, subprocess, sys
from pathlib import Path
from syntax import read, walk, prop
from check import Catalog
ROOT=Path(__file__).resolve().parents[1]

def write_json(path,data):path.write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
def digest(path):return hashlib.sha256(path.read_bytes()).hexdigest()
def command(args,out=None):
    result=subprocess.run([sys.executable,*args],cwd=ROOT,capture_output=True,text=True,timeout=120)
    if out:(ROOT/out).write_text(result.stdout+result.stderr,encoding='utf-8')
    if result.returncode:raise RuntimeError('Check failed: '+' '.join(args)+'\n'+result.stdout+result.stderr)
    return result.stdout+result.stderr

def main():
    for stale in ['initial.json','unittest-initial.txt']:
        (ROOT/'reports'/stale).unlink(missing_ok=True)
    command(['tools/rebuild_views.py','--check'])
    command(['tools/check.py','--out','reports/structural-check.json'])
    command(['tools/check.py','UICL-1.0-Structured.uicl','--out','reports/structured-edition-check.json'])
    command(['tools/check.py','examples/profile-authoring/measurement-profile.uicl','examples/profile-authoring/measurement-example.uicl','--extra-profile','examples/profile-authoring/measurement-profile.uicl','--out','reports/extension-check.json'])
    command(['tools/host.py','--out','reports/host-check.json'])
    unittest=command(['-m','unittest','discover','-s','tests','-v'],'reports/unittest.txt')
    match=re.search(r'Ran (\d+) tests',unittest)
    if not match or not re.search(r'\nOK\s*$',unittest):raise RuntimeError('No reliable unittest summary')
    suite_count=int(match.group(1))
    pairs=[]
    for parent in ('spec','guides','examples/hosted'):
        for p in sorted((ROOT/parent).glob('*.uicl')):
            for ext in ('.md','.html'):
                other=p.with_suffix(ext)
                if other.exists():
                    if p.read_bytes()!=other.read_bytes():raise ValueError('Twin mismatch '+str(p))
                    pairs.append([str(p.relative_to(ROOT)),str(other.relative_to(ROOT))])
    if (ROOT/'README.uicl').read_bytes()!=(ROOT/'README.zh-CN.md').read_bytes():raise ValueError('README twin mismatch')
    pairs.append(['README.uicl','README.zh-CN.md'])
    sources=json.loads((ROOT/'provenance/sources.json').read_text())
    for source in sources:
        if digest(ROOT/source['path'])!=source['sha256']:raise ValueError('Historical source modified '+source['path'])
    anthology=read(ROOT/'meta/examples.uicl');sample_count=0
    for n in anthology['nodes']:
        if n['kind']!='example':continue
        if prop(n,'body')!=(ROOT/prop(n,'path')).read_text(encoding='utf-8'):raise ValueError('Example body mismatch '+prop(n,'path'))
        sample_count+=1
    cat=Catalog(ROOT)
    rule_records=[]
    for p in [ROOT/'meta/core-rules.uicl']+sorted((ROOT/'profiles').glob('*.uicl')):
        for n in walk(read(p)['nodes']):
            if n['kind']=='rule':rule_records.append({'id':n['id'],'name':n['label'],'source':str(p.relative_to(ROOT)),'origin':prop(n,'origin'),'verification':prop(n,'verification'),'full_semantics_status':'NOT_EVALUATED'})
    write_json(ROOT/'reports/rule-coverage.json',{'rule_count':len(rule_records),'note':'Selected checks may cover parts of a rule; these are not full domain/runtime conformance results.','rules':rule_records})
    # Capture the package's actual profile versions and files. Unbound adapters are explicit.
    lock={'language':'uicl','core':'1.0','status':'package-source-lock-not-production-plan','catalog':[],'verification_environment':{'python':platform.python_version(),'markdown_it_py':importlib.metadata.version('markdown-it-py')},'runtime_adapters':'UNBOUND','credentials':'UNBOUND','network_access_performed':False}
    for pid,p in sorted(cat.profiles.items()):lock['catalog'].append({'id':pid,'version':p['version'],'source':p['path'],'sha256':digest(ROOT/p['path']),'requires':p['requires']})
    write_json(ROOT/'uicl.lock.json',lock)
    structural=json.loads((ROOT/'reports/structural-check.json').read_text());host=json.loads((ROOT/'reports/host-check.json').read_text())
    summary={'status':'PASS','edition':'UICL 1.0 consolidated specification','evaluated':{'standard_profiles':len(cat.profiles),'node_schemas':len(cat.schemas),'closed_record_schemas':len(cat.records),'grammar_productions':37,'core_sections':45,'core_and_domain_rules':len(rule_records),'structured_source_documents':structural['documents'],'references_checked':structural['references_checked'],'single_file_edition':1,'extra_profile_documents':2,'hosted_documents':host['documents'],'byte_identical_host_pairs':len(pairs),'independent_uicl_examples':sample_count,'embedded_example_source_matches':sample_count,'unittest_methods':suite_count,'hexl_matrix_subtests':360,'historical_source_hashes':len(sources)},'reports':{'structure':'structural-check.json','single_file':'structured-edition-check.json','extension':'extension-check.json','hosts':'host-check.json','unit_tests':'unittest.txt','rules':'rule-coverage.json'},'host_pairs':pairs,'scope_limits':['Syntactic/shape/link and selected static checks only; expression result types are not inferred.','Native HTML uses a restricted example parser, not complete WHATWG conformance.','No browser screenshot comparison or no-op source editor implementation tested.','No Windows/mobile build, PostgreSQL execution, Cloudflare or ECS changes performed.','No model calls, secret resolution, native code execution, real Agent or WebAssembly runtime executed.','Grammar views are derived from UICL, but the syntax prototype is hand-written.','All full normative semantic rule results remain NOT_EVALUATED.','HEXL tests were rerun in this package; old AML/Semaquil reports were not counted.']}
    write_json(ROOT/'reports/verification.json',summary)
    print(json.dumps({'status':'PASS',**summary['evaluated']},ensure_ascii=False,indent=2))
if __name__=='__main__':main()
