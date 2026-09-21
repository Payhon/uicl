"""Executed tests of the new package tools, not domain runtimes."""
import copy, json, sys, tempfile, unittest
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
sys.path.insert(0,str(ROOT/'tools'))
from syntax import parse, read, plain, prop, walk
from check import Catalog, PackageChecker, CheckError
from grammar_views import derive
from host import extract, decode_payload, HostError

CAT=Catalog(ROOT)

def node(body):return parse('uicl "1.0"\n'+body)['nodes'][0]
def v(body):return node('testnode\n  value: '+body+'\n')['props']['value']
def shape(body):return CAT.normalize(parse('uicl "1.0"\n'+body))

class SyntaxTests(unittest.TestCase):
    def test_minimal(self):self.assertEqual(node('app "你好"\n')['label'],'你好')
    def test_block_list_equivalence(self):self.assertEqual(v('[read, create]'),v('\n    - read\n    - create'))
    def test_compact_object_equivalence(self):self.assertEqual(v('[{name: "a", value: 1}, {name: "b", value: 2}]'),v('\n    - name: "a"\n      value: 1\n    - name: "b"\n      value: 2'))
    def test_nested_map(self):self.assertEqual(plain(v('\n    nested:\n      values:\n        - 1\n        - 2')),{'nested':{'values':[1,2]}})
    def test_nested_list(self):self.assertEqual(plain(v('\n    -\n      - 1\n      - 2\n    - [3, 4]')),[[1,2],[3,4]])
    def test_raw_is_inert(self):self.assertEqual(plain(v('|\n    @ref $data #id\n    - no')), '@ref $data #id\n- no\n')
    def test_fence_is_inert(self):self.assertEqual(plain(v('```python\n    print("@x")\n  ```')),'print("@x")\n')
    def test_block_raw_list(self):self.assertEqual(plain(v('\n    - |\n      first\n      second\n    - "third"')),['first\nsecond\n','third'])
    def test_negative_list(self):self.assertEqual(plain(v('\n    - -10\n    - 0')),[-10,0])
    def test_exponent_is_decimal(self):self.assertEqual(v('1e3')['type'],'decimal')
    def test_no_implicit_scalar_coercion(self):self.assertEqual(plain(v('[yes, no, on, true, false, null]')),['yes','no','on',True,False,None])
    def test_reference_string_distinction(self):self.assertEqual(v('"@x"')['tag'],'literal');self.assertEqual(v('@x')['tag'],'ref')
    def test_binding_and_list_not_flattened(self):self.assertEqual(v('$x')['tag'],'expression');self.assertEqual(v('[$x]')['tag'],'list')
    def test_pure_expression_precedence(self):
        tree=v('= $x + 2 * 3')['tree'];self.assertEqual(tree['op'],'+');self.assertEqual(tree['right']['op'],'*')
    def test_not_precedence(self):self.assertEqual(v('= not $a == $b')['tree']['operand']['op'],'==')
    def test_null_coalesce_right_associative(self):self.assertEqual(v('= $a ?? $b ?? $c')['tree']['right']['op'],'??')
    def test_comparison_chain_rejected(self):
        with self.assertRaises(ValueError):v('= $a < $b < $c')
    def test_implicit_division_rejected(self):
        with self.assertRaises(ValueError):v('= 1 / 3')
    def test_no_bare_expression_variable(self):
        with self.assertRaises(ValueError):v('= status == "active"')
    def test_duplicate_keys_rejected(self):
        for source in ('{a:1,a:2}','\n    a: 1\n    a: 2'):
            with self.subTest(source=source),self.assertRaises(ValueError):v(source)
    def test_duplicate_inline_and_block_rejected(self):
        with self.assertRaises(ValueError):node('app (title: "a")\n  title: "b"\n')
    def test_no_empty_block_value(self):
        with self.assertRaises(ValueError):v('')
    def test_mixed_collection_rejected(self):
        with self.assertRaises(ValueError):v('\n    - 1\n    a: 2')
    def test_node_inside_record_rejected(self):
        with self.assertRaises(ValueError):v('\n    name: "a"\n    ui.button "x"')
    def test_indent_jump_rejected(self):
        with self.assertRaises(ValueError):node('app\n    ui.text "x"\n')
    def test_tab_rejected(self):
        with self.assertRaises(ValueError):node('app\n\tui.text "x"\n')
    def test_unclosed_fence_rejected(self):
        with self.assertRaises(ValueError):v('```python\n    x = 1')
    def test_wrong_header_rejected(self):
        with self.assertRaises(ValueError):parse('aml "1.0"\napp "x"\n')
    def test_future_version_rejected(self):
        with self.assertRaises(ValueError):parse('uicl "2.0"\napp "x"\n')
    def test_invalid_unicode_rejected(self):
        with self.assertRaises(ValueError):v('"\\ud800"')
    def test_bom_crlf(self):self.assertEqual(parse('\ufeffuicl "1.0"\r\napp "你好"\r\n')['version'],'1.0')

class ShapeTests(unittest.TestCase):
    def test_primary_normalization(self):
        a,_=shape('ui.text "欢迎"\n');b,_=shape('ui.text\n  value: "欢迎"\n');self.assertEqual(a,b)
    def test_primary_conflict(self):
        with self.assertRaisesRegex(CheckError,'PRIMARY_CONFLICT'):shape('ui.text "欢迎"\n  value: "欢迎"\n')
    def test_missing_id(self):
        with self.assertRaisesRegex(CheckError,'MISSING_ID'):shape('state\n  value: 0\n')
    def test_duplicate_id(self):
        with self.assertRaisesRegex(CheckError,'DUPLICATE_ID'):shape('state #x\n  value: 0\nstate #x\n  value: 1\n')
    def test_unknown_property(self):
        with self.assertRaisesRegex(CheckError,'UNKNOWN_PROPERTY'):shape('ui.text "x"\n  typo: true\n')
    def test_unknown_node(self):
        with self.assertRaisesRegex(CheckError,'UNKNOWN_NODE'):shape('mystery "x"\n')
    def test_mandatory_property(self):
        with self.assertRaisesRegex(CheckError,'MISSING_PROPERTY'):shape('state #x\n')
    def test_server_access_required(self):
        with self.assertRaisesRegex(CheckError,'SERVER_ACCESS_REQUIRED'):shape('collection #x\n  storage: server\n  field "name"\n    type: text\n')
    def test_secret_not_plain_value(self):
        with self.assertRaisesRegex(CheckError,'UNKNOWN_PROPERTY'):shape('secret #password\n  source: runtime_prompt\n  scope: [ssh]\n  value: "not-allowed"\n')
    def test_nonempty_secret_scope(self):
        with self.assertRaisesRegex(CheckError,'SECRET_EMPTY_SCOPE'):shape('secret #password\n  source: runtime_prompt\n  scope: []\n')
    def test_image_dimension(self):
        with self.assertRaisesRegex(CheckError,'DIMENSION_PAIR'):shape('image #x "x"\n  prompt: "test"\n  size: [0, 1200]\n')
    def test_timeline_bounds(self):
        ast=read(ROOT/'examples/14-aigc-video.uicl')
        shot=next(n for n in walk(ast['nodes']) if n['id']=='detail');shot['props']['frames']=v('999')
        with self.assertRaisesRegex(CheckError,'TIMELINE_BOUNDS'):CAT.normalize(ast)
    def test_timeline_overlap(self):
        ast=read(ROOT/'examples/14-aigc-video.uicl')
        shot=next(n for n in walk(ast['nodes']) if n['id']=='detail');shot['props']['start_frame']=v('20')
        with self.assertRaisesRegex(CheckError,'TIMELINE_OVERLAP'):CAT.normalize(ast)
    def test_pg_transaction_conflict(self):
        text=(ROOT/'examples/09-postgresql.uicl').read_text().replace('atomic: split_allowed','atomic: required')
        with self.assertRaisesRegex(CheckError,'PG_TRANSACTION_CONFLICT'):CAT.normalize(parse(text))
    def test_pg_unknown_index_column(self):
        text=(ROOT/'examples/09-postgresql.uicl').read_text().replace('columns: [customer_id]','columns: [missing]')
        with self.assertRaisesRegex(CheckError,'PG_UNKNOWN_COLUMN'):CAT.normalize(parse(text))
    def test_schema_controls_acceptance(self):
        cat=copy.deepcopy(CAT);cat.schemas['ui.text']['fields']['value']['type']='int'
        with self.assertRaisesRegex(CheckError,'TYPE_SHAPE'):cat.normalize(parse('uicl "1.0"\nui.text "x"\n'))
    def test_extension_requires_explicit_profile(self):
        ast=read(ROOT/'examples/profile-authoring/measurement-example.uicl')
        with self.assertRaisesRegex(CheckError,'UNKNOWN_NODE'):CAT.normalize(ast)
        cat=Catalog(ROOT,[ROOT/'examples/profile-authoring/measurement-profile.uicl']);cat.normalize(ast)

class LinkTests(unittest.TestCase):
    def temp_check(self,files,entry):
        with tempfile.TemporaryDirectory(dir=ROOT/'reports') as d:
            d=Path(d)
            for name,body in files.items():(d/name).write_text('uicl "1.0"\n'+body,encoding='utf-8')
            checker=PackageChecker(ROOT,CAT);return checker.load(d/entry)
    def test_fullstack_links(self):
        checker=PackageChecker(ROOT,CAT);checker.load(ROOT/'examples/fullstack/project.uicl');self.assertEqual(len(checker.cache),9)
    def test_exported_reference(self):
        self.temp_check({'lib.uicl':'module #lib\n  exports: [x]\nstate #x\n  value: 0\n','main.uicl':'use #lib\n  source: @"./lib.uicl"\nset\n  target: @lib::x\n  value: 1\n'},'main.uicl')
    def test_private_reference_rejected(self):
        with self.assertRaisesRegex(CheckError,'NOT_EXPORTED'):
            self.temp_check({'lib.uicl':'module #lib\n  exports: []\nstate #x\n  value: 0\n','main.uicl':'use #lib\n  source: @"./lib.uicl"\nset\n  target: @lib::x\n  value: 1\n'},'main.uicl')
    def test_missing_reference(self):
        with self.assertRaisesRegex(CheckError,'UNRESOLVED_REF'):self.temp_check({'main.uicl':'set\n  target: @missing\n  value: 1\n'},'main.uicl')
    def test_unknown_port(self):
        with self.assertRaisesRegex(CheckError,'UNKNOWN_PORT'):self.temp_check({'main.uicl':'state #x\n  value: 0\nset\n  target: @x.missing\n  value: 1\n'},'main.uicl')
    def test_import_cycle(self):
        with self.assertRaisesRegex(CheckError,'IMPORT_CYCLE'):self.temp_check({'a.uicl':'use #b\n  source: @"./b.uicl"\n','b.uicl':'use #a\n  source: @"./a.uicl"\n'},'a.uicl')
    def test_import_path_escape(self):
        with self.assertRaisesRegex(CheckError,'PATH_ESCAPE'):self.temp_check({'main.uicl':'use #bad\n  source: @"../../../../etc/passwd"\n'},'main.uicl')
    def test_network_import_not_automatic(self):
        with self.assertRaisesRegex(CheckError,'IMPORT_NONLOCAL'):self.temp_check({'main.uicl':'use #remote\n  source: @"https://example.org/module.uicl"\n'},'main.uicl')
    def test_uri_is_not_fetched(self):
        result=self.temp_check({'main.uicl':'file #remote\n  media: "text/plain"\n  source: @"https://invalid.example/no-request"\n'},'main.uicl');self.assertIn('remote',result['ids'])

class GrammarTests(unittest.TestCase):
    def test_grammar_is_data(self):
        tree,text=derive(read(ROOT/'meta/grammar.uicl'));self.assertEqual(len(tree['rules']),37);self.assertIn('"uicl"',text)
    def test_mutated_grammar_changes_view(self):
        ast=read(ROOT/'meta/grammar.uicl');n=next(n for n in walk(ast['nodes']) if n['kind']=='g.literal' and prop(n,'text')=='uicl');n['props']['text']=v('"other"');self.assertIn('"other"',derive(ast)[1])
    def test_bad_grammar_reference_rejected(self):
        ast=read(ROOT/'meta/grammar.uicl');n=next(n for n in walk(ast['nodes']) if n['kind']=='g.ref');n['props']['target']=v('@absent')
        with self.assertRaisesRegex(ValueError,'Unresolved'):derive(ast)

class HostTests(unittest.TestCase):
    def test_markdown_code_fence_inert(self):
        result=extract((ROOT/'examples/hosted/article.uicl').read_text());ids=[n['id'] for i in result['islands'] for n in walk(i['ast']['nodes'])];self.assertEqual(ids,['summary'])
    def test_markdown_rename_bytes(self):self.assertEqual((ROOT/'examples/hosted/article.uicl').read_bytes(),(ROOT/'examples/hosted/article.md').read_bytes())
    def test_html_rename_bytes(self):self.assertEqual((ROOT/'examples/hosted/page.uicl').read_bytes(),(ROOT/'examples/hosted/page.html').read_bytes())
    def test_markdown_no_marker_no_activation(self):
        with self.assertRaises(HostError):extract('# ordinary\n',mode='markdown')
    def test_markdown_wrong_version(self):
        with self.assertRaises(HostError):extract('<!-- uicl:markdown 2.0 -->\n# x\n',mode='markdown')
    def test_markdown_duplicate_marker(self):
        with self.assertRaises(HostError):extract('<!-- uicl:markdown 1.0 -->\n\n<!-- uicl:markdown 1.0 -->\n\n# x\n')
    def test_markdown_missing_target(self):
        with self.assertRaisesRegex(HostError,'TARGET_MISSING'):extract('<!-- uicl:markdown 1.0 -->\n\n<!-- uicl\ndoc.block #x\n  target: next\n-->\n')
    def test_truncated_annotation(self):
        with self.assertRaisesRegex(HostError,'COMMENT_BOUNDARY'):extract('<!-- uicl:markdown 1.0 -->\n\n<!-- uicl\ndoc.block #x\n  target: next\n')
    def test_html_annotation(self):self.assertEqual(len(extract((ROOT/'examples/hosted/page.uicl').read_text())['islands']),1)
    def test_html_data_block(self):self.assertEqual(len(extract((ROOT/'examples/hosted/data-block.uicl').read_text())['islands']),1)
    def test_html_target_duplicate(self):
        source=(ROOT/'examples/hosted/page.uicl').read_text().replace('</body>','<p id="summary">duplicate</p></body>')
        with self.assertRaisesRegex(HostError,'TARGET_MISSING_OR_DUPLICATE'):extract(source)
    def test_html_bad_data_type(self):
        source=(ROOT/'examples/hosted/data-block.uicl').read_text().replace('type="text/plain"','type="text/javascript"')
        with self.assertRaisesRegex(HostError,'SCRIPT_TYPE'):extract(source)
    def test_html_template_inert(self):
        source=(ROOT/'examples/hosted/page.uicl').read_text().replace('</body>','<template><!-- uicl\ndoc.block #hidden\n  target: next\n--></template></body>')
        self.assertEqual(len(extract(source)['islands']),1)
    def test_base64_canonical(self):self.assertEqual(decode_payload('SGVsbG8'),'Hello')
    def test_base64_noncanonical(self):
        with self.assertRaises(HostError):decode_payload('Zh')
    def test_base64_forbidden_chars(self):
        with self.assertRaises(HostError):decode_payload('YWJj$')

if __name__=='__main__':unittest.main()
