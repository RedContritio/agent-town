# GDScript：Web 版与桌面版行为差异

## 布局系统

| 特性 | 桌面版 | Web 版 |
|------|--------|--------|
| `reset_size()` | 立即生效 | 延迟（1 帧后） |
| `size` 赋值 | 有效 | 可能被布局覆盖 |
| `scale` | 立即生效 | 立即生效 ✅ |
| `custom_minimum_size` | 影响布局 | 影响布局 |

## 建议

对于 Web 导出，建议：
1. 使用 `scale` 进行视觉缩放
2. 手动计算位置时考虑缩放
3. 避免依赖字体更改后的 `size` 更新

## 示例：动态大小标签

**桌面版做法**（Web 版不可靠）：
```gdscript
label.add_theme_font_size_override("font_size", target_size)
label.reset_size()
var pos = screen_pos - label.size / 2
```

**兼容 Web 的做法**：
```gdscript
var scale = target_size / float(BASE_FONT_SIZE)
label.scale = Vector2(scale, scale)
var pos = screen_pos - (label.size * label.scale) / 2
```

## 字体处理

两个平台都支持：
- `LabelSettings.font_size`
- `add_theme_font_size_override`

但布局时机不同。需要立即视觉更新时使用 `scale`。
