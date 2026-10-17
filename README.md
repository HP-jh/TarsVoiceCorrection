# TarsVoiceCorrection

**Tars 语音纠错引擎** — 单一纠错来源，被 TarsSecureGuard (B33) 与 TarsEnterprise (B36) 复用的 MCP Server。

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

## 零外部依赖

全部字典数据编译期嵌入 Go 源码，单二进制分发，无 Python runtime、无外部模型文件。

## 版本

当前版本：**v0.1.0**

## License

Apache-2.0
