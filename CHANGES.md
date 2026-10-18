## v0.3.0 - 2026-10-18

### 集成版本：历史优点汇总 + 新功能

**整合的历史优点（v0.1.x ~ v0.2.x）：**
- 三套纠错策略：编辑距离（Damerau-Levenshtein）+ 拼音匹配（850常用字）+ 同义词映射（316组技术词）
- 简繁互转（678组字对）与中英混输空格规范化
- 多策略投票置信度评分与候选列表生成
- 自定义词典支持（~/.tars/dict/custom.txt）
- MCP Server 三种传输：stdio / SSE / streamable HTTP
- 五大 MCP 工具：voice_correction / pinyin_match / synonym_lookup / fuzzy_search / lang_detect
- 自动更新（tars-update）：版本比对 + SHA256校验 + 备份自重启 + 稳定版回退
- 零外部依赖：全部字典数据编译期嵌入 Go 源码，单二进制分发
- 渗透测试覆盖（路径遍历 / XSS 注入防护）
- 模糊测试覆盖（FuzzCorrectorCorrect / FuzzPinyinMatch / FuzzFuzzySearch）

**v0.3.0 新增功能：**
- 批量纠错（BatchCorrect）：一次处理多段文本，并发 worker 池可控（默认4，上限16）
- 纠错结果缓存（ResultCache）：256 条 LRU 缓存，重复输入毫秒级命中
- MCP 新增 batch_correction 工具：JSON 数组输入，批量返回纠错结果

## v0.2.2 - 2026-10-18
  - Patch version bump
