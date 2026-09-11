# CutX 图标候选选择

## 候选总览

| 编号 | 名称 | 设计方向 | SVG 源文件 | PNG 预览 |
|------|------|---------|-----------|----------|
| 01 | Scissors | 剪刀切割文件块，直观的"cut"隐喻 | `svg/option-01-scissors.svg` | `option-01-scissors.png` |
| 02 | Puzzle | 拼图碎片重组，对应"像拼图一样切割与合并" | `svg/option-02-puzzle.svg` | `option-02-puzzle.png` |
| 03 | Split Arrows | 一分为二的箭头，split/merge 双向 | `svg/option-03-split-arrows.svg` | `option-03-split-arrows.png` |
| 04 | File Stack | 分层文件块堆叠，chunk 可视化 | `svg/option-04-file-stack.svg` | `option-04-file-stack.png` |
| 05 | Hexagon | 六边形碎裂，几何技术感 | `svg/option-05-hexagon.svg` | `option-05-hexagon.png` |
| 06 | Blade | 刀片切割文档，极简切割线 | `svg/option-06-blade.svg` | `option-06-blade.png` |
| 07 | Merge Ring | 拼合圆环，merge 概念 | `svg/option-07-merge-ring.svg` | `option-07-merge-ring.png` |
| 08 | X-Split | 抽象 X 由分裂箭头组成，品牌标识 | `svg/option-08-x-split.svg` | `option-08-x-split.png` |

## 推荐方案（Top 3）

### 🥇 首选：08 X-Split
- **理由**：抽象的 X 形态直接呼应产品名 Cut**X**，蓝绿双色箭头分别代表 split（蓝）和 merge（绿），中心交点象征切割点。图形极简，在 32px favicon 下依然可辨认。作为 Windows 应用图标辨识度最高。
- **32px 可读性**：✅ 强 — X 形态在极小尺寸下依然清晰

### 🥈 次选：01 Scissors
- **理由**：剪刀是最直观的"切割"隐喻，无需文字解释即可理解工具用途。绿色切割线增加了视觉层次。
- **32px 可读性**：⚠️ 中等 — 剪刀细节在 32px 下稍显复杂

### 🥉 第三：03 Split Arrows
- **理由**：一分为二的箭头简洁有力，同时表达 split 和 merge 双向语义。蓝色和绿色分别对应两个方向，色彩对比清晰。
- **32px 可读性**：✅ 强 — 箭头形状在小尺寸下清晰

## 下一步

1. 选择一个候选编号（01-08）
2. 运行导出脚本生成 Windows ICO 格式：
   ```bash
   python3 .agents/skills/qiaomu-icon-generator/scripts/export_selected_icon.py \
     --source design/cutx-icon-options/option-08-x-split.png \
     --web-out design/cutx-icon-options/final \
     --web-prefix cutx-icon
   ```
3. 将选中的 PNG 转换为 Windows .ico 格式（多尺寸：16/32/48/64/128/256）
