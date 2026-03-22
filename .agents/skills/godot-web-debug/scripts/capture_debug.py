#!/usr/bin/env python3
"""带交互的 Godot Web 截图工具

用法：
    capture_debug.py <output> <action> [params]

示例：
    # 默认截图（当前视角）
    capture_debug.py /tmp/shot.png

    # 缩放（负数缩小，正数放大）
    capture_debug.py /tmp/shot.png zoom -50

    # 拖拽旋转相机（dx, dy）
    capture_debug.py /tmp/shot.png drag 100 50

    # 按键盘（1-5 切换视角，+/- 缩放）
    capture_debug.py /tmp/shot.png key 1

    # 移动到坐标位置并截图（先缩小，移动，再放大）
    capture_debug.py /tmp/shot.png goto 100 50

环境要求：
    - 需要 Playwright 已安装（chromium）
    - 必须使用 headed 模式
"""

from playwright.sync_api import sync_playwright
import sys

DEFAULT_ZOOM = 0  # 滚轮每次缩放量
DEFAULT_DRAG_STEP = 100  # 每次拖拽的像素量

def capture_with_action(output, action, *args):
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=False)
        page = browser.new_page(viewport={"width": 1280, "height": 720})

        print(f"→ 正在访问 http://localhost:8080...")
        page.goto("http://localhost:8080", wait_until="networkidle", timeout=60000)

        print("→ 等待 Godot WebAssembly 加载（15秒）...")
        page.wait_for_timeout(15000)

        # 根据 action 执行交互
        if action == "zoom":
            delta = int(args[0]) if args else DEFAULT_ZOOM
            print(f"→ 滚轮缩放: {delta}")
            page.mouse.wheel(0, delta)

        elif action == "drag":
            dx = int(args[0]) if len(args) > 0 else DEFAULT_DRAG_STEP
            dy = int(args[1]) if len(args) > 1 else 0
            print(f"→ 拖拽旋转: dx={dx}, dy={dy}")
            # 获取页面中心点
            page.mouse.move(640, 360)
            page.mouse.down()
            page.mouse.move(640 + dx, 360 + dy)
            page.mouse.up()

        elif action == "key":
            key = args[0] if args else "1"
            print(f"→ 按键: {key}")
            page.keyboard.press(key)

        elif action == "move":
            x = int(args[0]) if len(args) > 0 else 640
            y = int(args[1]) if len(args) > 1 else 360
            print(f"→ 移动鼠标到: ({x}, {y})")
            page.mouse.move(x, y)

        elif action == "click":
            x = int(args[0]) if len(args) > 0 else 640
            y = int(args[1]) if len(args) > 1 else 360
            print(f"→ 点击: ({x}, {y})")
            page.mouse.click(x, y)

        elif action == "scroll":
            x = int(args[0]) if len(args) > 0 else 640
            y = int(args[1]) if len(args) > 1 else 360
            delta = int(args[2]) if len(args) > 2 else -100
            print(f"→ 滚动到 ({x}, {y}) 滚动量: {delta}")
            page.mouse.move(x, y)
            page.mouse.wheel(0, delta)

        # 等待操作生效
        page.wait_for_timeout(1000)

        print(f"→ 保存截图到 {output}...")
        page.screenshot(path=output, full_page=False)

        browser.close()
        print(f"✓ 截图已保存：{output}")
        return output

if __name__ == "__main__":
    output = sys.argv[1] if len(sys.argv) > 1 else "/tmp/godot_screenshot.png"
    action = sys.argv[2] if len(sys.argv) > 2 else ""
    args = sys.argv[3:] if len(sys.argv) > 3 else []

    capture_with_action(output, action, *args)