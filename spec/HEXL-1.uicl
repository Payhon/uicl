<!-- uicl:markdown 1.0 -->

<!-- uicl
document #hexl_spec "HEXL/1 分行字节封装规范"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# HEXL/1 分行字节封装规范

以下协议正文沿用历史 HEXL/1，名称不随 UICL 更名；本包 codec 测试重新执行。源码/CLI 位置为 tools/hexl.py。

# HEXL/1：分行二进制文本封装工作草案

## 1. 目标

把任意有限字节流封装成可逐行处理的 ASCII 文本。每个数据行独立编码固定数量的原始字节，最后一行可缩短。HEXL 是本提案的编码容器名称，不宣称首次发明十六进制转储、Base64 或分块传输。

它是 AML 可引用的独立格式，不要求每个数据行都写成 AML 节点或 JSON 对象。

## 2. 规范格式

```text
!hexl 1 <encoding> <chunk_bytes>
<16位小写十六进制字节偏移>:<encoded_payload>
...
!end <total_bytes> <data_lines> sha256:<64位小写十六进制摘要>
```

实际文件不包含省略号。编码为 ASCII，所有行必须以 LF 结束，无 BOM、CR、注释、空白行或末尾附加数据。

header 与 footer 的各字段之间恰为一个 ASCII 空格。十进制整数不允许多余前导零。偏移固定为 16 个小写十六进制字符，0 起算。最大文件字节数为 2^64-1。

chunk_bytes 的协议原型范围为 1..1,048,576，默认 48。更大块需要后续版本或扩展，不得让未知 chunk_size 无限制分配内存。最大文件大小还受宿主 max_bytes 限制。

## 3. 四种编码

| encoding | 字母表和填充 | 建议扩展名 |
|---|---|---|
| hex | 小写 0-9a-f，不填充 | .hexl |
| base64 | RFC 4648 标准字母表，规范 `=` 填充 | .b64l |
| base64url | RFC 4648 URL-safe 字母表，本协议规定无填充 | .b64l |
| base32 | RFC 4648 大写 A-Z2-7，本协议规定无填充 | .b32l |

以 header 为准，不能根据文件扩展名猜解码方式。

Base64url 的 '-' 和 '_' 替代 '+' 和 '/'；这不是从已编码字符串中删除字符。只有本协议明确可推导的 padding 可以省略。所有 unused bits 必须为零。解码器必须拒绝非字母表字符、错误填充和可解码但非规范的表示，不能“尽量猜出来”。

纯字母数字需求可使用 base32。Base62 没有在本草案中注册：需要另外固定字母顺序、前导零、块长度、终块规则和算法，不能只写 basexx 让实现者自由选择。

## 4. 固定的是原始字节，不是编码字符

把整份文件先 Base64 后任意切字符串，可能使每行无法独立解码。本协议先把原始字节切块，再分别编码。

默认 48 字节对应 96 个 hex 字符或 64 个 Base64 字符。Base32 的 48 字节独立块有 77 个无填充字符。因此不允许直接串接所有 base32/base64url 载荷后仅做一次解码；应按行独立解码并拼接原始字节。

第 i 个完整数据块的偏移必须为 i * chunk_bytes。任何重复、乱序、间隙都必须报错。只有最终数据行允许少于 chunk_bytes 且至少 1 字节。

空文件没有数据行，只有 header 和 footer，total_bytes=0，data_lines=0，摘要是空字节串的 SHA-256。

## 5. 完整示例

原始内容为 `AML` 后跟 LF，共四字节。

```text
!hexl 1 hex 4
0000000000000000:414d4c0a
!end 4 1 sha256:e93ef79161e260f8bd71148ee3541839cb8f6b29bd3dcfbaffa62c3db82985db
```

同一文件的 Base64url 表示：

```text
!hexl 1 base64url 4
0000000000000000:QU1MCg
!end 4 1 sha256:e93ef79161e260f8bd71148ee3541839cb8f6b29bd3dcfbaffa62c3db82985db
```

两个 footer 摘要相同，因为摘要计算对象是原始字节，不是 ASCII 封装。

## 6. 校验顺序

校验 header、版本和额度；有界读取每行；验证偏移与编码；更新累计摘要和字节数；验证 footer 的字节数、行数、摘要；要求 EOF；然后允许发布输出。

footer 之前的输出是未验证临时字节。解析错误、提前 EOF、取消或 checksum mismatch 时必须丢弃临时输出，不能留下看起来完整的文件。

摘要确保与声明内容一致，但同文件里的摘要不证明作者身份。攻击者可同时修改载荷与 footer。可信交付需要外部锁文件、已验证的签名或受信内容摘要。

仅检验整个文件的 footer 不能立即认证一个随机读出的块。需要随机块认证时，另行定义签名的 chunk manifest / Merkle 索引；不是基础 HEXL 的隐含保证。

## 7. 流式与随机访问

编码器无需预先知道总字节数，可以一次顺序读取并在末尾输出汇总。解码器可按行返回原始块，但调用者在总摘要通过前不得把它们视作已验证产物。

对 canonical HEXL 的完整中间块，记录宽度固定，可从 header 长度、编码、chunk_bytes 计算行位置。最终短块和 footer 另行处理。压缩后的传输流不能直接使用该随机偏移公式。

基础版本不定义压缩。压缩和加密可由明确的外部容器层承担，必须限制解压资源并记录各层摘要；不能偷偷压缩后仍宣称是原始字节的固定块映射。

基础编码不承诺增量编辑高效。文件前部插入字节会使后续固定分块变化；内容寻址的分块/CAS 包可作为另一个存储 Profile，不能伪装成基础 HEXL 的原有行为。

## 8. 长度与用途

忽略行框架开销，hex 长度为 2N，标准 Base64 为 4*ceil(N/3)，大块无填充 Base32 约为 1.6N。每行独立编码会产生额外尾块对齐开销，文件实际长度必须逐块计算。

这是文本化，不是压缩。ASCII 更短不自动等于模型 token 更少。大型媒体通常通过旁挂对象/CAS 引用传递；单文件文本快照、教学、小型嵌入、字节级测试和流式封装适合 HEXL。

## 9. 与 AML 集成

```aml
file #payload
  media: "application/octet-stream"
  representation: hexl
  source: @"./payload.b64l"
```

或者将完整 HEXL 文本放入 `payload` 原文围栏。本地引用或内嵌 payload 二者择一，不能制造两个相互冲突的事实来源。

`media` 指解码后的原始资源。representation 描述传输封装。解码得到 EXE、DLL、SO、脚本或带宏文档也不会自动获得执行权限。

## 10. 参考原型

```bash
python reference/hexl.py pack input.bin output.hexl
python reference/hexl.py pack input.bin output.b64l --encoding base64url --chunk 48
python reference/hexl.py unpack output.b64l restored.bin --max-bytes 1073741824
python -m unittest discover -s tests -v
```

CLI 默认拒绝覆盖现有目标，只有 --force 才允许替换。临时文件仅在校验成功后发布。示例实现依赖 Python 标准库，不需要安装外部包。

该实现不是经过正式安全审计的生产文件服务。服务化时还需要并发、隔离、权限、配额和运行环境审计。

## 参考依据

RFC 4648, The Base16, Base32, and Base64 Data Encodings:
https://www.rfc-editor.org/rfc/rfc4648

本草案自行定义分行、偏移、header/footer、无填充 Profile 和严格规范化规则。不得把完整 HEXL 协议称为 RFC 4648 已规定的格式。
