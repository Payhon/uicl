#!/usr/bin/env python3
"""UICL 1.0 seed syntax reader. No network, imports, evaluation or program execution.
The trusted bootstrap reads UICL's own declarative grammar and schema files.
The separate grammar recognizer checks the token grammar again from UICL data.
This adapted token/value parser is used by syntax.py; its legacy line parser is not a public entry point. No runtime or effect checking.
"""
from __future__ import annotations
import argparse
from dataclasses import dataclass
import json
from pathlib import Path
import re
import sys
from typing import Any, Iterable

MAX_BYTES = 16 * 1024 * 1024
MAX_DEPTH = 96
IDENT = re.compile(r'[A-Za-z_][A-Za-z0-9_]*')
NUMBER = re.compile(r'(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?')

class UICLError(ValueError):
    def __init__(self, code: str, message: str, line: int = 1, column: int = 1):
        self.code, self.line, self.column = code, line, column
        super().__init__(f'{code} at {line}:{column}: {message}')

@dataclass(frozen=True)
class Token:
    kind: str
    text: str
    line: int
    column: int
    data: Any = None

def val(kind: str, value: Any) -> dict:
    return {'tag': 'literal', 'type': kind, 'value': value}

def scan_line(text: str, line: int, offset: int = 0) -> list[Token]:
    out: list[Token] = []
    i = 0
    while i < len(text):
        c = text[i]
        if c == ' ': i += 1; continue
        start = i
        if c == '"':
            i += 1
            escaped = False
            while i < len(text):
                d = text[i]; i += 1
                if d == '"' and not escaped: break
                if d == '\\' and not escaped: escaped = True
                else: escaped = False
            else: raise UICLError('UICL-S005', 'Unterminated string', line, start + offset + 1)
            raw = text[start:i]
            try:
                decoded = json.loads(raw)
                if any(0xD800 <= ord(d) <= 0xDFFF for d in decoded):
                    raise ValueError('Unpaired Unicode surrogate')
            except ValueError as e:
                raise UICLError('UICL-S005', str(e), line, start + offset + 1) from e
            out.append(Token('STRING', raw, line, start + offset + 1, decoded)); continue
        m = NUMBER.match(text, i)
        if m:
            raw = m.group(); i = m.end()
            out.append(Token('NUMBER', raw, line, start + offset + 1)); continue
        m = IDENT.match(text, i)
        if m:
            raw = m.group(); i = m.end()
            out.append(Token('IDENT', raw, line, start + offset + 1)); continue
        pair = text[i:i+2]
        if pair in ('::', '==', '!=', '<=', '>=', '??', '?.'):
            i += 2; out.append(Token(pair, pair, line, start+offset+1)); continue
        if c in '#@$():,.[]{}=+-*%<>':
            i += 1; out.append(Token(c, c, line, start+offset+1)); continue
        raise UICLError('UICL-S003', f'Unexpected character {c!r}', line, start+offset+1)
    return out

class Parser:
    BP={'??':10,'or':20,'and':30,'==':40,'!=':40,'<':40,'<=':40,'>':40,'>=':40,'in':40,'+':50,'-':50,'*':60,'%':60}
    COMP={'==','!=','<','<=','>','>=','in'}
    def __init__(self,tokens:list[Token]): self.ts=tokens; self.p=0; self.depth=0
    def peek(self,k:str|None=None): return self.ts[self.p] if k is None else self.ts[self.p].kind==k
    def take(self,k:str|None=None):
        t=self.ts[self.p]
        if k is not None and t.kind!=k: raise UICLError('UICL-S003',f'Expected {k}, got {t.kind} {t.text!r}',t.line,t.column)
        self.p+=1; return t
    def error(self,message:str,code='UICL-S003'):
        t=self.peek(); raise UICLError(code,message,t.line,t.column)
    def gap(self):
        prev=self.ts[self.p-1]; cur=self.peek()
        if prev.line==cur.line and cur.column<=prev.column+len(prev.text): self.error('Header fields require a space')
    def adjacent(self):
        prev=self.ts[self.p-1];cur=self.peek()
        if prev.line!=cur.line or cur.column!=prev.column+len(prev.text): self.error('No whitespace inside an anchor, reference or binding prefix')
    def qname(self):
        parts=[self.take('IDENT').text]
        while self.peek('.'):
            self.take();parts.append(self.take('IDENT').text)
        return '.'.join(parts)
    def key(self): return self.take().data if self.peek('STRING') else self.qname()
    def reference(self):
        self.take('@'); self.adjacent()
        if self.peek('STRING'): return {'tag':'ref','uri':self.take().data}
        first=self.take('IDENT').text; module=None
        if self.peek('::'):
            self.adjacent(); self.take(); self.adjacent(); module=first; first=self.take('IDENT').text
        ports=[]
        while self.peek('.'):
            self.adjacent();self.take();self.adjacent();ports.append(self.take('IDENT').text)
        return {'tag':'ref','module':module,'id':first,'ports':ports}
    def binding(self):
        self.take('$');self.adjacent()
        node={'tag':'binding','name':self.take('IDENT').text}
        while self.peek().kind in ('.','?.','['):
            t=self.take()
            if t.kind=='[':
                index=self.expr();self.take(']');node={'tag':'index','object':node,'index':index}
            else: node={'tag':'member','object':node,'name':self.take('IDENT').text,'optional':t.kind=='?.'}
        return node
    def pair(self,expression=False):
        k=self.key();self.take(':');return k,self.expr() if expression else self.value()
    def aggregate(self,record:bool,expression:bool=False):
        opening,closing=('{','}') if record else ('[',']');self.take(opening)
        fields={};items=[]
        if not self.peek(closing):
            while True:
                if record:
                    k,v=self.pair(expression)
                    if k in fields:self.error(f'Duplicate key {k}', 'UICL-S004')
                    fields[k]=v
                else:items.append(self.expr() if expression else self.value())
                if not self.peek(','):break
                self.take()
                if self.peek(closing):self.error('Trailing comma is forbidden')
        self.take(closing)
        return {'tag':'record','fields':fields} if record else {'tag':'list','items':items}
    def value(self):
        self.depth+=1
        try:
            if self.depth>MAX_DEPTH:self.error('Value nesting exceeds limit','UICL-LIMIT')
            t=self.peek()
            if t.kind=='=':self.take();return {'tag':'expression','tree':self.expr()}
            if t.kind=='$':return {'tag':'expression','tree':self.binding()}
            if t.kind=='@':return self.reference()
            if t.kind=='STRING':return val('text',self.take().data)
            if t.kind in ('NUMBER','-'):
                negative=''
                if self.peek('-'):self.take();negative='-'
                n=self.take('NUMBER').text
                return val('decimal' if any(c in n for c in '.eE') else 'int',negative+n)
            if t.kind=='IDENT':
                if t.text in ('true','false','null'):
                    self.take();return val('null' if t.text=='null' else 'bool',{'true':True,'false':False,'null':None}[t.text])
                return val('text',self.qname())
            if t.kind=='[':return self.aggregate(False)
            if t.kind=='{':return self.aggregate(True)
            self.error('Expected value')
        finally:self.depth-=1
    def expr(self,minimum:int=0):
        self.depth+=1
        try:
            if self.depth>MAX_DEPTH:self.error('Expression nesting exceeds limit','UICL-LIMIT')
            t=self.peek()
            if t.kind=='-' or t.text=='not':
                self.take();left={'tag':'unary','op':t.text,'operand':self.expr(35 if t.text=='not' else 70)}
            elif t.kind=='(':
                self.take();left=self.expr();self.take(')')
            elif t.kind=='$':left=self.binding()
            elif t.kind=='@':left=self.reference()
            elif t.kind=='[':left=self.aggregate(False,True)
            elif t.kind=='{':left=self.aggregate(True,True)
            elif t.kind in ('STRING','NUMBER') or t.text in ('true','false','null'):left=self.value()
            elif t.kind=='IDENT' and t.text not in ('and','or','in'):
                name=self.qname();self.take('(');args=[]
                if not self.peek(')'):
                    while True:
                        args.append(self.expr())
                        if not self.peek(','):break
                        self.take()
                self.take(')');left={'tag':'pure_call','name':name,'args':args}
            else:self.error('Expected expression; variables require $')
            while self.peek().kind in ('.','?.','['):
                p=self.take()
                if p.kind=='[':
                    index=self.expr(); self.take(']'); left={'tag':'index','object':left,'index':index}
                else:
                    left={'tag':'member','object':left,'name':self.take('IDENT').text,'optional':p.kind=='?.'}
            compared=False
            while True:
                t=self.peek();bp=self.BP.get(t.text)
                if bp is None or bp<minimum:break
                if t.text in self.COMP:
                    if compared:self.error('Comparison chains require explicit boolean operators')
                    compared=True
                self.take();right=self.expr(bp if t.text=='??' else bp+1)
                left={'tag':'binary','op':t.text,'left':left,'right':right}
            return left
        finally:self.depth-=1
    def property_ahead(self):
        old=self.p
        try:
            if self.peek().kind not in ('IDENT','STRING'):return False
            self.key();return bool(self.peek(':'))
        finally:self.p=old
