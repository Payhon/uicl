<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_aigc "生成内容规格与任务 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# 生成内容规格与任务 Profile 1.0

标识：`uicl.aigc@1.0.0`。状态：规范整合稿。依赖：resource, design, testing。

文本、图片、视频、音频、演示文稿；规格与产物分离。


## `Ratio`

有理数比率。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `num` | `int` | 是 | 否 | 分子。 |
| `den` | `int` | 是 | 否 | 正分母。 |

## `text`

文本生成规格，与 ui.text 不同。

主文本：`label`；身份：`required`；允许子节点：require, review。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `label` | `text` | 否 | 否 | 规格名称。 |
| `prompt` | `text` | 是 | 允许，仍需业务检查 | 生成意图。 |
| `language` | `text` | 是 | 否 | 输出语言。 |
| `sources` | `list<ref>` | 否 | 否 | 显式输入资料。 |
| `prefer` | `record` | 否 | 否 | 语气、结构等偏好，不作硬保证。 |

## `image`

图像生成规格，不是已有文件。

主文本：`label`；身份：`required`；允许子节点：scene, overlay, require, review。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `label` | `text` | 否 | 否 | 规格名称。 |
| `prompt` | `text` | 是 | 允许，仍需业务检查 | 画面意图。 |
| `size` | `list<int>` | 是 | 否 | [width,height]。 |
| `color_space` | `text` | 否 | 否 | 色彩空间。 |
| `references` | `list<ref>` | 否 | 否 | 参考产物。 |
| `prefer` | `record` | 否 | 否 | 视觉偏好，适配器报告支持程度。 |

## `scene`

图像场景层。

主文本：`None`；身份：`optional`；允许子节点：subject。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `background` | `text` | 否 | 否 | 背景意图。 |

## `subject`

一个主要对象的期望描述。

主文本：`label`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `label` | `text` | 是 | 否 | 主体说明。 |
| `count` | `int` | 否 | 否 | 期望数量，须检查或评审。 |
| `region` | `list<number>` | 是 | 否 | 归一化 [x,y,w,h]。 |

## `overlay`

精确合成层。

主文本：`None`；身份：`optional`；允许子节点：overlay.text, overlay.image。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|

## `overlay.text`

确定性排版的文字图层。

主文本：`value`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `value` | `text` | 是 | 否 | 要精确排版的文字。 |
| `region` | `list<number>` | 是 | 否 | 归一化矩形。 |
| `font` | `text` | 是 | 否 | 锁定字体绑定。 |
| `font_px` | `int` | 是 | 否 | 字号。 |
| `align` | `text` | 否 | 否 | 对齐方式。 |

## `overlay.image`

已存在图片合成层。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `src` | `ref` | 是 | 否 | 图片产物。 |
| `region` | `list<number>` | 是 | 否 | 归一化矩形。 |

## `video`

视频生成与合成规格。

主文本：`label`；身份：`required`；允许子节点：track, require, review。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `label` | `text` | 否 | 否 | 视频名称。 |
| `size` | `list<int>` | 是 | 否 | 画幅。 |
| `fps` | `record<Ratio>` | 是 | 否 | 帧率有理数。 |
| `frames` | `int` | 是 | 否 | 总帧数。 |

## `track`

视频/字幕/音轨。

主文本：`None`；身份：`required`；允许子节点：shot, cue。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `kind` | `text` | 是 | 否 | 轨道类型。 |
| `overlap` | `text` | 否 | 否 | 同轨重叠处理。 |

## `shot`

位于整数时间轴的镜头。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `start_frame` | `int` | 是 | 否 | 起始帧。 |
| `frames` | `int` | 是 | 否 | 持续帧数。 |
| `prompt` | `text` | 是 | 否 | 镜头意图。 |
| `reference` | `ref` | 否 | 否 | 角色或产品参考。 |

## `cue`

字幕/音轨片段。

主文本：`text`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `text` | `text` | 是 | 否 | 文字。 |
| `start_frame` | `int` | 是 | 否 | 开始帧。 |
| `frames` | `int` | 是 | 否 | 持续帧数。 |

## `audio`

音频规格。

主文本：`label`；身份：`required`；允许子节点：require, review。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `label` | `text` | 否 | 否 | 名称。 |
| `text` | `text` | 是 | 否 | 朗读文字。 |
| `voice_binding` | `text` | 是 | 否 | 合规声音绑定，不假定已获克隆许可。 |
| `sample_rate` | `int` | 是 | 否 | 采样率。 |
| `channels` | `int` | 是 | 否 | 声道数。 |

## `deck`

可编辑演示文稿规格。

主文本：`label`；身份：`required`；允许子节点：slide, require, review。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `label` | `text` | 否 | 否 | 演示名称。 |
| `ratio` | `text` | 是 | 否 | 画幅比例。 |
| `editable` | `bool` | 是 | 否 | 对象可编辑要求。 |
| `design` | `ref` | 否 | 否 | 共享设计系统。 |

## `slide`

一页幻灯片。

主文本：`title`；身份：`required`；允许子节点：slide.text, slide.image, slide.chart。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `title` | `text` | 是 | 否 | 页面标题。 |
| `layout` | `text` | 是 | 否 | 锁定布局名。 |
| `notes` | `text` | 否 | 否 | 演讲备注。 |

## `slide.text`

幻灯片文字对象。

主文本：`value`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `value` | `text` | 是 | 否 | 文字内容。 |

## `slide.image`

幻灯片图片。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `src` | `ref` | 是 | 否 | 实际图片或任务产物。 |
| `alt` | `text` | 是 | 否 | 图片说明。 |

## `slide.chart`

有真实数据的图表。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `kind` | `text` | 是 | 否 | 图表类型。 |
| `data` | `list<record>` | 是 | 否 | 图表数据；缺失不能臆造。 |
| `editable` | `bool` | 是 | 否 | 是否要求原生可编辑图表。 |

## `generate`

把规格转成实际产物的任务。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `from` | `ref` | 是 | 否 | 规格节点。 |
| `phase` | `text` | 是 | 否 | 执行阶段。 |
| `output` | `text` | 是 | 否 | 输出媒体类型。 |
| `binding` | `text` | 是 | 否 | 实际生成/合成适配器。 |
| `limits` | `record<CallLimits>` | 是 | 否 | 时间、次数与费用预算。 |
| `approval` | `text` | 是 | 否 | 外部生成授权。 |

输出端口：`output: Artifact`，可用阶段 `execution`。

## 规则 AIGC_CONTRACT

Spec、Generator、Task、Artifact、Slot 不可隐式互换。@generate.output 仅建立依赖；读取文件、预览、检查或页面重绘不得自动触发生成或收费。


验证类别：`type-effect-runtime`。

## 规则 AIGC_GATES

require 为注册检查，prefer 为偏好，review 为评审。不能把硬要求悄悄降级成 prompt；适配器不支持时报告 UNSUPPORTED_CAPABILITY。生成结果必须本地验证且绑定证据。


验证类别：`adapter-validation`。

## 规则 AIGC_TIMELINE

视频时间用整数帧及有理数帧率；检查边界与重叠。editable deck 的文字/图表不能用整页截图冒充。字体和输入资产必须锁定且有使用权。


验证类别：`artifact-validation`。
