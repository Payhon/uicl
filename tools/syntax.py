#!/usr/bin/env python3
"""UICL 1.0 structural syntax prototype. No execution or resource access.
Parses nodes, block/inline values, raw text and expression syntax; not a semantic compiler.
The original source bytes are retained by read(); there is no formatting or patch API.
"""
from __future__ import annotations
import argparse, json, re, sys
from pathlib import Path
from typing import Iterable
from scalar import Parser, Token, UICLError, scan_line, MAX_BYTES, MAX_DEPTH


def tokens(text: str, line: int, offset: int=0):
    return scan_line(text,line,offset)+[Token('EOF','',line,len(text)+offset+1)]

def error(message: str,line=1,code='UICL-SYNTAX'):
    raise UICLError(code,message,line,1)

def plain(v, default=None):
    if v is None: return default
    if v['tag']=='literal':
        return int(v['value']) if v['type']=='int' else v['value']
    if v['tag']=='raw': return v['text']
    if v['tag']=='list': return [plain(x) for x in v['items']]
    if v['tag']=='record': return {k:plain(x) for k,x in v['fields'].items()}
    return v

def prop(n,name,default=None): return plain(n['props'].get(name),default)

def walk(nodes: Iterable[dict]):
    for n in nodes:
        yield n
        yield from walk(n['children'])

class Reader:
    def __init__(self, source: str):
        try: size=len(source.encode('utf-8'))
        except UnicodeEncodeError: error('Invalid Unicode scalar')
        if size>MAX_BYTES:error('Source byte limit exceeded',code='UICL-LIMIT')
        self.source=source
        source=source.removeprefix('\ufeff').replace('\r\n','\n')
        if '\r' in source or '\x00' in source:error('Bare CR or NUL')
        self.lines=source.split('\n');self.i=0; self.depth=0
    def row(self):
        while self.i<len(self.lines):
            raw=self.lines[self.i]; text=raw.lstrip(' '); indent=len(raw)-len(text)
            if not text or text.startswith('//'):self.i+=1;continue
            if '\t' in raw:error('Tab outside raw block',self.i+1,'UICL-LAYOUT')
            if indent%2:error('Indent must use two-space units',self.i+1,'UICL-LAYOUT')
            if indent//2>MAX_DEPTH:error('Indentation limit',self.i+1,'UICL-LIMIT')
            return indent,text,self.i+1
        return None
    def is_property(self,text,line=1,offset=0):
        # Only scan the key prefix; raw markers deliberately are not Core tokens.
        m=re.match(r'^("(?:[^"\\]|\\.)*"|[A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*)\s*:(?!:)',text)
        if not m:return False
        p=Parser(tokens(m.group(1),line,offset));p.key();p.take('EOF');return True
    def property_parts(self,text,line,indent):
        m=re.match(r'^("(?:[^"\\]|\\.)*"|[A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*)\s*:(?!:)(.*)$',text)
        if not m:error('Expected record key and colon',line)
        p=Parser(tokens(m.group(1),line,indent));k=p.key();p.take('EOF')
        return k,m.group(2).strip(),indent+m.start(2)
    def inline(self,text,line,offset=0):
        p=Parser(tokens(text,line,offset));v=p.value();p.take('EOF');return v
    def raw(self,marker,indent,line):
        content=[];prefix=' '*(indent+2)
        if marker=='|':
            while self.i<len(self.lines):
                r=self.lines[self.i]
                if not r.strip(' '):content.append('');self.i+=1;continue
                if not r.startswith(prefix):break
                content.append(r[len(prefix):]);self.i+=1
            if not content or not any(content):error('Use "" for empty text',line,'UICL-RAW')
            return {'tag':'raw','text':'\n'.join(content).rstrip('\n')+'\n','form':'pipe','language':None}
        m=re.fullmatch(r'(`{3,})([A-Za-z_][A-Za-z0-9_.]*)?',marker)
        if not m:error('Invalid raw fence',line,'UICL-RAW')
        closing=' '*indent+m.group(1)
        while self.i<len(self.lines):
            r=self.lines[self.i]
            if r==closing:
                self.i+=1
                return {'tag':'raw','text':''.join(c+'\n' for c in content),'form':'fence','language':m.group(2)}
            if not r:content.append('')
            elif r.startswith(prefix):content.append(r[len(prefix):])
            else:error('Raw content or closing fence indentation mismatch',self.i+1,'UICL-RAW')
            self.i+=1
        error('Unclosed raw fence',line,'UICL-RAW')
    def property(self,indent):
        row=self.row()
        if not row or row[0]!=indent:error('Property indentation mismatch',self.i+1,'UICL-LAYOUT')
        _,text,line=row;k,tail,offset=self.property_parts(text,line,indent);self.i+=1
        if not tail:
            row=self.row()
            if not row or row[0]!=indent+2:error('Missing block value',line,'UICL-MISSING_VALUE')
            v=self.block(indent+2)
        elif tail=='|' or tail.startswith('```'):v=self.raw(tail,indent,line)
        else:v=self.inline(tail,line,offset)
        return k,v
    def block(self,indent):
        self.depth+=1
        try:
            if self.depth>MAX_DEPTH:error('Block nesting limit',self.i+1,'UICL-LIMIT')
            r=self.row()
            if not r or r[0]!=indent:error('Block indentation mismatch',self.i+1,'UICL-LAYOUT')
            return self.block_list(indent) if r[1]=='-' or r[1].startswith('- ') else self.record(indent)
        finally:self.depth-=1
    def record(self,indent):
        fields={}
        while (r:=self.row()) and r[0]>=indent:
            if r[0]!=indent:error('Unexpected record indentation',r[2],'UICL-LAYOUT')
            k,v=self.property(indent)
            if k in fields:error('Duplicate record key '+k,r[2],'UICL-DUPLICATE_KEY')
            fields[k]=v
        if not fields:error('Empty block record',self.i+1)
        return {'tag':'record','fields':fields}
    def block_list(self,indent):
        items=[]
        while (r:=self.row()) and r[0]>=indent:
            if r[0]!=indent:error('Unexpected list indentation',r[2],'UICL-LAYOUT')
            _,text,line=r
            if text!='-' and not text.startswith('- '):error('Mixed list and record',line,'UICL-LAYOUT')
            tail=text[1:].strip()
            if tail and self.is_property(tail,line,indent+2):
                # Compact item record has a virtual key column two spaces after '-'.
                self.lines[self.i]=' '*(indent+2)+tail
                items.append(self.record(indent+2))
            else:
                self.i+=1
                if not tail:
                    nxt=self.row()
                    if not nxt or nxt[0]!=indent+2:error('Missing list item',line,'UICL-MISSING_VALUE')
                    items.append(self.block(indent+2))
                elif tail=='|' or tail.startswith('```'):items.append(self.raw(tail,indent,line))
                else:items.append(self.inline(tail,line,indent+2))
        return {'tag':'list','items':items}
    def node(self,indent):
        r=self.row()
        if not r or r[0]!=indent:error('Node indentation mismatch',self.i+1,'UICL-LAYOUT')
        _,text,line=r;p=Parser(tokens(text,line,indent));kind=p.qname();identity=None;label=None;props={}
        if p.peek('#'):
            p.gap();p.take();p.adjacent();identity=p.take('IDENT').text
        if p.peek().kind in ('IDENT','STRING'):
            p.gap();label=p.take().data if p.peek('STRING') else p.qname()
        if p.peek('('):
            p.gap();p.take()
            if not p.peek(')'):
                while True:
                    k,v=p.pair()
                    if k in props:error('Duplicate property '+k,line,'UICL-DUPLICATE_KEY')
                    props[k]=v
                    if not p.peek(','):break
                    p.take()
                    if p.peek(')'):error('Trailing comma',line)
            p.take(')')
        p.take('EOF');self.i+=1;children=[]
        while (r:=self.row()) and r[0]>indent:
            if r[0]!=indent+2:error('Skipped node indentation',r[2],'UICL-LAYOUT')
            if self.is_property(r[1],r[2],r[0]):
                k,v=self.property(indent+2)
                if k in props:error('Duplicate property '+k,r[2],'UICL-DUPLICATE_KEY')
                props[k]=v
            else:children.append(self.node(indent+2))
        return {'kind':kind,'id':identity,'label':label,'props':props,'children':children,'span':{'line':line,'column':indent+1}}
    def document(self):
        r=self.row()
        if not r or r[0]!=0:error('Missing version header')
        p=Parser(tokens(r[1],r[2]));t=p.take('IDENT')
        if t.text!='uicl':error('Expected uicl version header',r[2],'UICL-VERSION')
        p.gap();version=p.take('STRING').data;p.take('EOF')
        if version!='1.0':error('Unsupported UICL version',r[2],'UICL-VERSION')
        self.i+=1;nodes=[]
        while self.row():nodes.append(self.node(0))
        if not nodes:error('At least one declaration required')
        return {'language':'uicl','version':version,'nodes':nodes}

def parse(source:str)->dict:return Reader(source).document()
def read(path:Path)->dict:
    data=path.read_bytes()
    if len(data)>MAX_BYTES:error('Byte limit exceeded',code='UICL-LIMIT')
    try:return parse(data.decode('utf-8'))
    except UnicodeDecodeError:error('Source must be UTF-8')

if __name__=='__main__':
    cli=argparse.ArgumentParser(description=__doc__);cli.add_argument('source',type=Path);cli.add_argument('--out',type=Path)
    args=cli.parse_args()
    try:
        output=json.dumps(read(args.source),ensure_ascii=False,indent=2)+'\n'
        if args.out:args.out.write_text(output,encoding='utf-8')
        else:print(output,end='')
    except (ValueError,OSError) as exc:print(exc,file=sys.stderr);raise SystemExit(1)
