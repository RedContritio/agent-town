#!/usr/bin/env python3
"""Launch browser for debug environment"""

from playwright.sync_api import sync_playwright
import time
import sys

def main():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=False)
        page = browser.new_page(viewport={'width': 1280, 'height': 720})
        
        # 捕获所有 console 日志（包括 Godot 的 print）
        def handle_console(msg):
            print(f"[BROWSER][{msg.type}] {msg.text}", flush=True)
        page.on("console", handle_console)
        
        # 捕获页面错误
        def handle_page_error(err):
            print(f"[BROWSER][PAGE_ERROR] {err}", flush=True)
        page.on("pageerror", handle_page_error)
        
        page.goto('http://localhost:8080', wait_until='networkidle')
        print('[LAUNCHER] Browser launched. Waiting for godot-web to initialize...', flush=True)
        page.wait_for_timeout(15000)  # 15秒，保持事件循环运行
        print('[LAUNCHER] Godot-web should be ready now.', flush=True)
        print('[LAUNCHER] Browser will stay open. Press Ctrl+C to close.', flush=True)
        try:
            while True:
                page.wait_for_timeout(1000)  # 保持事件循环运行
        except KeyboardInterrupt:
            pass
        browser.close()

if __name__ == "__main__":
    main()
