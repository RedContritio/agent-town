---
name: godot-web-debug
description: 使用截图验证调试 Godot Web UI 问题。当用户报告 Godot Web UI 问题时触发（如"UI元素大小不对"、"标签没缩放"、"按钮位置错了"）。使用 Playwright 自动截图来观察修复前后的实际 UI 状态。
---

# Godot Web 调试

通过截图验证来调试和修复 Godot Web UI 问题。

## ⚠️ 重要：必须验证图形界面

**AI 必须实际看到图形界面才能确认问题已修复！**

- Godot WebAssembly 需要 **headed 模式**（非 headless）才能正常渲染
- headless 模式会卡在加载界面，无法验证 3D 渲染
- 每次修复后必须截图并用 Read 工具查看

## 何时使用

以下情况使用本技能：
- 用户报告 Godot Web UI 渲染问题
- UI 元素无法正确缩放/调整大小
- 视觉元素显示错位或大小不对
- 需要可视化验证 UI 修复效果
- 代码修改涉及 Godot 场景或脚本

## 快速开始

```bash
# 1. 重启服务器并截图（脚本会自动准备 Playwright 环境）
./.agents/skills/godot-web-debug/scripts/restart-and-capture.sh

# 2. 查看截图确认问题
# → 使用 Read 工具读取 /tmp/godot_screenshot.png

# 3. 或者单独截图（服务器已运行的情况下）
./.agents/skills/godot-web-debug/scripts/screenshot.sh /tmp/my_screenshot.png

# 4. 修复代码...

# 5. 验证修复
./.agents/skills/godot-web-debug/scripts/restart-and-capture.sh
# → 与之前的截图对比
```

## 工作流程

### 1. 复现问题

```bash
# 重启服务器（确保使用最新代码）
./.agents/skills/godot-web-debug/scripts/restart-server.sh

# 捕获当前状态
./.agents/skills/godot-web-debug/scripts/screenshot.sh /tmp/godot_before.png

# 查看截图（使用 Read 工具读取图片路径）
# 路径: /tmp/godot_before.png
```

### 2. 分析问题

基于截图观察：
- 元素位置是否正确？
- 大小是否符合预期？
- 缩放是否按预期工作？

复杂问题请参考 [references/debug-patterns.md](references/debug-patterns.md)

### 3. 修复

根据观察结果应用修复。常见模式：

```gdscript
# 标签缩放问题
var scale = target_size / base_size
label.scale = Vector2(scale, scale)
label.position = screen_pos - (label.size * label.scale) / 2
```

### 4. 验证

```bash
# 重新构建并截图
make web && ./.agents/skills/godot-web-debug/scripts/restart-server.sh
./.agents/skills/godot-web-debug/scripts/screenshot.sh /tmp/godot_after.png

# 对比（使用 Read 工具读取图片路径）
# 路径: /tmp/godot_after.png
```

## 常见问题

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| 标签背景不缩放 | Web 布局延迟 | 使用 `label.scale` 而非 `font_size` 覆盖 |
| 场景消失 | GDScript 编译错误 | 检查变量重复定义 |
| 截图为灰色 | WASM 崩溃 | 检查文件大小（<100KB = 崩溃） |

更多内容见 [references/common-traps.md](references/common-traps.md)

## 脚本

所有 AI 调试脚本位于 `.agents/skills/godot-web-debug/scripts/`：

- `restart-server.sh` - 重启 Godot Web 服务器
- `restart-and-capture.sh` - 重启+截图组合命令
- `screenshot.sh` - 截图工具（调用 capture.py）
- `capture.py` - Playwright 截图实现

## 共享库

- `lib.sh` - 技能内部共享的工具函数（路径查找、虚拟环境管理等）

## 参考文档

- [references/debug-patterns.md](references/debug-patterns.md) - 调试工作流模式
- [references/common-traps.md](references/common-traps.md) - Godot Web 常见陷阱
- [references/gdscript-web-diffs.md](references/gdscript-web-diffs.md) - Web 与桌面版行为差异
