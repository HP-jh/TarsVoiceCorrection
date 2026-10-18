# TarsVoiceCorrection

**Tars 语音纠错引擎** — 单一纠错来源，被 TarsSecureGuard (B33) 与 TarsEnterprise (B36) 复用的 MCP Server。

## 关于作者

> 你好！我是这个项目的作者，一名正在读小学的学生。
> 这个项目从规划到代码实现，**完全由我自己主导设计，并借助 AI 工具完成**。每一行代码、每一个功能模块，都凝聚了我对技术的热爱与坚持。
> 如果你认可这个项目，或者觉得它对你有帮助，希望能给我一个 **Star** 支持，这对我意义重大！

## 欢迎支持

- 如果你在使用过程中遇到问题，欢迎提交 [Issue](https://github.com/HP-jh) 或参与讨论
- 如果你觉得项目不错，请点一个 **Star**，这是对我最大的鼓励
- 也欢迎分享给身边有需要的朋友，让更多人看到这个项目
- 后续功能迭代会继续推进，你的支持是我持续更新的动力

## ⚠️ 维护说明

> **本项目依赖 AI 服务套餐，若因套餐额度耗尽或服务商调整，可能导致更新暂时停滞。**
> 如遇较长时间未更新，请理解这是 AI 服务限制所致，并非项目放弃维护。
> 恢复服务后会第一时间继续推进，感谢你的耐心与包容。

---

## 核心定位

语音/文本输入常含：拼写错、同音字、同义词替换、多字少字、中英混输。本引擎输出纠错后的标准化文本，附带置信度与候选列表。

## 三套纠错策略

| 策略 | 说明 | 典型场景 |
|------|------|----------|
| 编辑距离 | Damerau-Levenshtein 距离 + 归一化相似度 | 拼写错误（电恼→电脑） |
| 拼音匹配 | 内置 850 常用字拼音映射 | 同音字（中过→中国） |
| 同义词映射 | 内置 316 组技术领域同义词 | 说法不一（服务器→主机） |

外加：简繁互转（678 组字对）、中英混输空格规范化、多策略投票置信度评分。

## 自定义词典

用户可在 `~/.tars/dict/custom.txt` 按行填入自定义词汇，引擎优先匹配。

## MCP Server

支持 stdio / SSE / streamable 三种传输：

```bash
# stdio（默认，用于 MCP Client 子进程）
./tars-voice-correction -transport=stdio

# SSE over HTTP
./tars-voice-correction -transport=sse -addr=:8080

# Streamable HTTP
./tars-voice-correction -transport=streamable -addr=:8080
```

### 工具清单

- `voice_correction` — 主纠错（输入文本 → 纠错结果 + 候选 + 置信度）
- `pinyin_match` — 拼音同音词查找
- `synonym_lookup` — 同义词查询
- `fuzzy_search` — 模糊搜索
- `lang_detect` — 语言检测

### 直接模式（不启 MCP Server）

```bash
./tars-voice-correction -text="电恼" -mode=correct
./tars-voice-correction -word="中国" -mode=pinyin
./tars-voice-correction -word="电脑" -mode=synonym
```

## 自动更新

内置 `tars-update` 模块，启动时自动检测新版本，支持一键升级：

```bash
# 检查是否有新版本
./tars-voice-correction update check

# 下载并安装最新版本（自动校验 SHA256、备份旧版本、自重启）
./tars-voice-correction update apply
```

- 访问 GitHub Releases 比对版本
- 下载失败或版本解析异常时自动回退到 `stable_versions.json` 记录的最后稳定版本
- 六平台二进制全覆盖（Linux / macOS / Windows，amd64 / arm64）

## 零外部依赖

全部字典数据编译期嵌入 Go 源码，单二进制分发，无 Python runtime、无外部模型文件。

## 版本

当前版本：**v0.2.0**

## License

Apache-2.0

## Changelog

### vx.y.z
- Version bump with stability improvements
