# Godot Web 调试模式

## 模式 1：基于距离的缩放

**问题**：UI 元素应根据相机距离缩放。

**截图分析**：
- 在远处截图
- 拉近后截图
- 对比元素大小

**解决方案**：
```gdscript
var scale_factor = REFERENCE_DISTANCE / distance
scale_factor = clamp(scale_factor, MIN_SCALE, MAX_SCALE)
label.scale = Vector2(scale_factor, scale_factor)
# 定位时必须考虑缩放：
label.position = screen_pos - (label.size * label.scale) / 2
```

**验证**：
1. 远处截图：元素较小
2. 拉近截图：元素成比例变大
3. 文件大小均为 ~200-300KB（无崩溃）

## 模式 2：大小 vs 缩放

**桌面版 Godot**（有效）：
```gdscript
label.add_theme_font_size_override("font_size", new_size)
label.reset_size()
```

**Web 版 Godot**（布局延迟）：
```gdscript
# size 延迟一帧更新 - 无法用于立即定位
```

**Web 解决方案**（立即生效）：
```gdscript
label.scale = Vector2(scale, scale)
# 视觉和边界都立即缩放
```

## 模式 3：编译错误检测

**症状**：截图只显示灰色背景 + HTML UI

**文件大小检查**：
```bash
ls -lh /tmp/godot_screenshot.png
# 正常：200-300KB
# 崩溃：<100KB
```

**常见原因**：
- 变量重复定义
- 引用缺失（如 `@onready` 节点未找到）
- 除以零

**修复**：检查 Godot 控制台或添加 print 语句
