#!/usr/bin/env python3
"""Derive read-only grammar views from meta/grammar.uicl. Not a parser generator."""
from __future__ import annotations
import argparse, json
from pathlib import Path
from syntax import read, prop

def derive(ast):
    roots=[n for n in ast['nodes'] if n['kind']=='grammar']
    if len(roots)!=1:raise ValueError('Expected one grammar')
    rules={};order=[]
    for n in roots[0]['children']:
        if n['kind']!='g.rule':raise ValueError('Unexpected grammar child')
        if n['id'] in rules:raise ValueError('Duplicate grammar rule')
        rules[n['id']]=n;order.append(n['id'])
    start=prop(roots[0],'start')['id']
    if start not in rules:raise ValueError('Unresolved start')
    def tree(n):
        k=n['kind'];children=n['children']
        if k in ('g.seq','g.choice'):
            if not children:raise ValueError('Empty sequence/choice')
            return {'kind':k[2:],'items':[tree(c) for c in children]}
        if children and k not in ('g.repeat',):raise ValueError('Unexpected grammar children')
        if k=='g.ref':
            target=prop(n,'target')['id']
            if target not in rules:raise ValueError('Unresolved grammar reference '+target)
            return {'kind':'ref','name':rules[target]['label'] or target}
        if k=='g.token':return {'kind':'token','name':prop(n,'name')}
        if k=='g.literal':return {'kind':'literal','text':prop(n,'text')}
        if k=='g.repeat':
            lo,hi=prop(n,'min'),prop(n,'max')
            if lo<0 or hi is not None and hi<lo or len(children)!=1:raise ValueError('Invalid repetition bounds')
            return {'kind':'repeat','min':lo,'max':hi,'item':tree(children[0])}
        raise ValueError('Unknown grammar kind '+k)
    result={'start':rules[start]['label'] or start,'rules':{}}
    for name in order:
        n=rules[name]
        if len(n['children'])!=1:raise ValueError('Rule requires one root')
        result['rules'][n['label'] or name]=tree(n['children'][0])
    def eb(t):
        k=t['kind']
        if k in ('ref','token'):return t['name']
        if k=='literal':return json.dumps(t['text'],ensure_ascii=False)
        if k=='seq':return ', '.join(eb(x) for x in t['items'])
        if k=='choice':return '('+' | '.join(eb(x) for x in t['items'])+')'
        lo,hi=t['min'],t['max'];s=eb(t['item'])
        if (lo,hi)==(0,1):return '[ '+s+' ]'
        if (lo,hi)==(0,None):return '{ '+s+' }'
        if (lo,hi)==(1,None):return '( '+s+' ), { '+s+' }'
        return '('+s+'){'+str(lo)+','+('unbounded' if hi is None else str(hi))+'}'
    text='(* Generated from meta/grammar.uicl; layout and lexical side conditions also apply. *)\n\n'
    text+='\n\n'.join(k+' = '+eb(v)+' ;' for k,v in result['rules'].items())+'\n'
    return result,text

def main():
    root=Path(__file__).resolve().parents[1];a=argparse.ArgumentParser(description=__doc__);a.add_argument('--check',action='store_true');args=a.parse_args()
    graph,ebnf=derive(read(root/'meta/grammar.uicl'))
    outputs={root/'generated/grammar-tree.json':json.dumps(graph,ensure_ascii=False,indent=2)+'\n',root/'generated/core.ebnf':ebnf}
    for path,text in outputs.items():
        if args.check:
            if not path.exists() or path.read_text(encoding='utf-8')!=text:raise ValueError('Stale view '+str(path))
        else:path.write_text(text,encoding='utf-8')
    print(json.dumps({'rules':len(graph['rules']),'status':'PASS','mode':'check' if args.check else 'derive'}))
if __name__=='__main__':main()
