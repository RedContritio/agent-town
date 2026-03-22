#!/usr/bin/env python3
"""
Web 视角观察工具 - 支持任意视角、分辨率、交互状态截图

用法：
    web-viewer.py [选项]

示例：
    # 默认视角（1280x720）
    web-viewer.py
    
    # 指定分辨率
    web-viewer.py --width 1920 --height 1080
    
    # 移动端视角
    web-viewer.py --device "iPhone 14"
    
    # 全页面截图
    web-viewer.py --full-page
    
    # 等待特定时间后截图
    web-viewer.py --wait 10
    
    # 执行交互后截图（输入 Token 并登录）
    web-viewer.py --action login --token "your-token-here"
    
    # 点击特定坐标后截图
    web-viewer.py --click 640,360 --wait 2
    
    # 连续截图（观察变化）
    web-viewer.py --sequence 5 --interval 3
"""

from playwright.sync_api import sync_playwright
import sys
import argparse
import time
from pathlib import Path


def get_device_descriptors():
    """获取常见设备预设"""
    return {
        "desktop": {"viewport": {"width": 1280, "height": 720}, "user_agent": "desktop"},
        "desktop-hd": {"viewport": {"width": 1920, "height": 1080}, "user_agent": "desktop"},
        "desktop-4k": {"viewport": {"width": 3840, "height": 2160}, "user_agent": "desktop"},
        "ipad": {"viewport": {"width": 810, "height": 1080}, "user_agent": "tablet"},
        "ipad-pro": {"viewport": {"width": 1024, "height": 1366}, "user_agent": "tablet"},
        "iphone-se": {"viewport": {"width": 375, "height": 667}, "user_agent": "mobile"},
        "iphone-14": {"viewport": {"width": 390, "height": 844}, "user_agent": "mobile"},
        "iphone-14-pro": {"viewport": {"width": 393, "height": 852}, "user_agent": "mobile"},
        "pixel-7": {"viewport": {"width": 412, "height": 915}, "user_agent": "mobile"},
    }


def capture(
    output="/tmp/web_viewer.png",
    url="http://localhost:8080",
    width=1280,
    height=720,
    full_page=False,
    wait_seconds=15,
    device=None,
    action=None,
    token=None,
    click_coords=None,
    selector=None,
    selector_visible=None,
):
    """
    捕获 Web 页面截图
    
    参数：
        output: 截图保存路径
        url: 目标网址
        width, height: 视口尺寸
        full_page: 是否截取全页面
        wait_seconds: 初始等待时间
        device: 设备预设名称
        action: 交互动作 (login, click)
        token: 登录 Token
        click_coords: 点击坐标 (x, y)
        selector: 等待选择器出现
        selector_visible: 等待选择器可见
    """
    device_presets = get_device_descriptors()
    
    with sync_playwright() as p:
        # 启动浏览器（headed 模式以支持 Godot WASM）
        browser = p.chromium.launch(headless=False)
        
        # 创建设备上下文
        context_options = {}
        
        if device and device.lower() in device_presets:
            preset = device_presets[device.lower()]
            context_options["viewport"] = preset["viewport"]
            print(f"→ 使用设备预设: {device} ({preset['viewport']['width']}x{preset['viewport']['height']})")
        else:
            context_options["viewport"] = {"width": width, "height": height}
            print(f"→ 设置视口: {width}x{height}")
        
        context = browser.new_context(**context_options)
        page = context.new_page()
        
        # 访问页面
        print(f"→ 正在访问 {url}...")
        page.goto(url, wait_until="networkidle", timeout=60000)
        
        # 等待 Godot WASM 加载
        print(f"→ 等待页面加载（{wait_seconds}秒）...")
        page.wait_for_timeout(wait_seconds * 1000)
        
        # 等待特定选择器（如果指定）
        if selector:
            print(f"→ 等待元素出现: {selector}")
            page.wait_for_selector(selector, timeout=30000)
        
        if selector_visible:
            print(f"→ 等待元素可见: {selector_visible}")
            page.wait_for_selector(selector_visible, state="visible", timeout=30000)
        
        # 执行交互动作
        if action == "login" and token:
            print(f"→ 执行登录操作...")
            # 查找 Token 输入框（假设是 Godot 生成的输入框）
            # 注意：Godot WASM 的输入框可能需要特殊处理
            try:
                # 尝试点击登录面板区域（左下角附近）
                page.click("canvas", position={"x": 100, "y": height - 50})
                page.wait_for_timeout(500)
                # 输入 Token
                page.keyboard.type(token)
                page.wait_for_timeout(500)
                # 按回车
                page.keyboard.press("Enter")
                page.wait_for_timeout(2000)
                print("✓ 登录操作完成")
            except Exception as e:
                print(f"⚠ 登录操作失败: {e}")
        
        elif action == "click" and click_coords:
            x, y = click_coords
            print(f"→ 点击坐标: ({x}, {y})")
            page.click("canvas", position={"x": x, "y": y})
            page.wait_for_timeout(1000)
        
        # 保存截图
        print(f"→ 保存截图到 {output}...")
        page.screenshot(path=output, full_page=full_page)
        
        browser.close()
        print(f"✓ 截图已保存: {output}")
        return output


def capture_sequence(
    output_template="/tmp/web_viewer_{:03d}.png",
    url="http://localhost:8080",
    count=5,
    interval=3,
    **kwargs
):
    """连续截图序列"""
    outputs = []
    for i in range(count):
        output = output_template.format(i)
        print(f"\n=== 截图 {i+1}/{count} ===")
        capture(output=output, url=url, **kwargs)
        outputs.append(output)
        if i < count - 1:
            print(f"→ 等待 {interval} 秒...")
            time.sleep(interval)
    return outputs


def main():
    parser = argparse.ArgumentParser(
        description="Web 视角观察工具 - 支持任意视角截图",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
设备预设:
  desktop, desktop-hd, desktop-4k
  ipad, ipad-pro
  iphone-se, iphone-14, iphone-14-pro
  pixel-7

示例:
  # 默认 1280x720
  python web-viewer.py
  
  # 4K 分辨率
  python web-viewer.py --device desktop-4k
  
  # 移动端视角
  python web-viewer.py --device iphone-14 --output /tmp/mobile.png
  
  # 全页面截图
  python web-viewer.py --full-page --output /tmp/full.png
  
  # 等待 30 秒后截图（等待 Godot 完全加载）
  python web-viewer.py --wait 30
  
  # 连续截图 10 张，间隔 5 秒
  python web-viewer.py --sequence 10 --interval 5
        """
    )
    
    parser.add_argument("-o", "--output", default="/tmp/web_viewer.png", help="输出文件路径")
    parser.add_argument("-u", "--url", default="http://localhost:8080", help="目标 URL")
    parser.add_argument("-W", "--width", type=int, default=1280, help="视口宽度")
    parser.add_argument("-H", "--height", type=int, default=720, help="视口高度")
    parser.add_argument("-f", "--full-page", action="store_true", help="全页面截图")
    parser.add_argument("-w", "--wait", type=int, default=15, help="加载等待时间（秒）")
    parser.add_argument("-d", "--device", help="设备预设名称")
    
    parser.add_argument("-a", "--action", choices=["login", "click"], help="交互动作")
    parser.add_argument("-t", "--token", help="登录 Token（配合 --action login）")
    parser.add_argument("-c", "--click", help="点击坐标 x,y（配合 --action click）")
    
    parser.add_argument("-s", "--selector", help="等待选择器出现")
    parser.add_argument("-v", "--selector-visible", help="等待选择器可见")
    
    parser.add_argument("-S", "--sequence", type=int, help="连续截图数量")
    parser.add_argument("-i", "--interval", type=int, default=3, help="截图间隔（秒）")
    
    parser.add_argument("--list-devices", action="store_true", help="列出所有设备预设")
    
    args = parser.parse_args()
    
    if args.list_devices:
        print("可用设备预设:")
        for name, preset in get_device_descriptors().items():
            vp = preset["viewport"]
            print(f"  {name:20s} {vp['width']:4d}x{vp['height']}")
        return
    
    # 解析点击坐标
    click_coords = None
    if args.click:
        try:
            x, y = map(int, args.click.split(","))
            click_coords = (x, y)
        except ValueError:
            print("错误: 点击坐标格式应为 x,y（例如: 640,360）")
            sys.exit(1)
    
    # 执行截图
    kwargs = {
        "url": args.url,
        "width": args.width,
        "height": args.height,
        "full_page": args.full_page,
        "wait_seconds": args.wait,
        "device": args.device,
        "action": args.action,
        "token": args.token,
        "click_coords": click_coords,
        "selector": args.selector,
        "selector_visible": args.selector_visible,
    }
    
    if args.sequence:
        template = args.output.replace(".png", "_{:03d}.png")
        outputs = capture_sequence(
            output_template=template,
            count=args.sequence,
            interval=args.interval,
            **kwargs
        )
        print(f"\n✓ 序列截图完成，共 {len(outputs)} 张")
        for o in outputs:
            print(f"  - {o}")
    else:
        capture(output=args.output, **kwargs)


if __name__ == "__main__":
    main()
