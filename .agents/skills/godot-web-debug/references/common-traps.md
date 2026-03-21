# Godot Web 常见陷阱

## ⚠️ 陷阱 0：使用 headless 模式（最重要！）

**Godot WebAssembly 必须使用 headed 模式渲染！**

```python
# ❌ 错误：headless 模式无法渲染 Godot WebAssembly
browser = p.chromium.launch(headless=True)
# 结果：永远卡在加载界面

# ✅ 正确：headed 模式可以正常渲染
browser = p.chromium.launch(headless=False)
# 结果：能看到完整的 3D 游戏界面
```

**原因**：Godot Web 使用 SharedArrayBuffer 和多线程，需要浏览器的完整支持。

**验证方法**：检查截图是否显示 Godot 加载画面（圆形进度条）还是 3D 世界。

---

## 陷阱 1：变量重复定义

```gdscript
# ❌ Web 导出会崩溃
var scale = 1.0
# ...
var scale = 2.0  # 编译错误！

# ✅ 安全写法
var scale = 1.0
# ...
scale = 2.0  # 重新赋值 OK
# 或
var label_scale = 2.0  // 使用不同名称
```

**影响**：Godot WebAssembly 编译失败，场景无法渲染

**检测**：截图文件大小 < 100KB，只有灰色背景

## 陷阱 2：假设布局立即更新

```gdscript
# ❌ 桌面版有效，Web 版延迟
label.add_theme_font_size_override("font_size", 20)
print(label.size)  // Web 中可能显示旧大小

# ✅ 使用 scale 实现立即效果
label.scale = Vector2(1.5, 1.5)
```

## 陷阱 3：截图时机错误

```python
# ❌ 太早（Godot WASM 需要更长时间）
page.goto(url)
page.wait_for_timeout(5000)  # 可能还是加载界面
page.screenshot(path="out.png")

# ✅ 正确：等待足够长时间
page.goto(url, wait_until="networkidle")
page.wait_for_timeout(15000)  # Godot WebAssembly 较大，需要更长等待
page.screenshot(path="out.png")
```

## 陷阱 4：定位时忘记考虑缩放

```gdscript
# ❌ 缩放后位置错误
label.scale = Vector2(2, 2)
label.position = screen_pos - label.size / 2

# ✅ 考虑缩放
label.position = screen_pos - (label.size * label.scale) / 2
```

## 陷阱 5：忘记重新构建 Web 导出

**症状**：截图中看不到代码更改的效果

**原因**：修改代码后忘记执行 `make web`

**修复**：始终运行 `make web && restart-server.sh`
