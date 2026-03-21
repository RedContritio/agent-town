#!/usr/bin/env python3
"""捕获 Godot Web 截图

用法：
    capture.py [输出路径] [网址]

示例：
    capture.py                           # 默认：/tmp/godot_screenshot.png
    capture.py /tmp/before.png
    capture.py /tmp/test.png http://localhost:8080

环境要求：
    - 需要 Playwright 已安装（chromium）
    - 必须使用 headed 模式（headless=False）才能渲染 Godot WebAssembly
"""

from playwright.sync_api import sync_playwright
import sys

def capture(output="/tmp/godot_screenshot.png", url="http://localhost:8080"):
    with sync_playwright() as p:
        # 必须使用 headed 模式！headless 模式无法渲染 Godot WebAssembly
        browser = p.chromium.launch(headless=False)
        page = browser.new_page(viewport={"width": 1280, "height": 720})
        
        print(f"→ 正在访问 {url}...")
        page.goto(url, wait_until="networkidle", timeout=60000)
        
        print("→ 等待 Godot WebAssembly 加载（15秒）...")
        page.wait_for_timeout(15000)
        
        print(f"→ 保存截图到 {output}...")
        page.screenshot(path=output, full_page=False)
        
        browser.close()
        print(f"✓ 截图已保存：{output}")
        return output

if __name__ == "__main__":
    output = sys.argv[1] if len(sys.argv) > 1 else "/tmp/godot_screenshot.png"
    url = sys.argv[2] if len(sys.argv) > 2 else "http://localhost:8080"
    capture(output, url)
