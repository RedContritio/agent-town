#!/usr/bin/env python3
"""
Web 观察工具 - 任意位置、任意视角截图

通过在浏览器中执行 JavaScript 控制 Godot Web 的相机位置

用法:
    web-observer.py [选项]

示例:
    # 默认视角
    web-observer.py

    # 俯视视角
    web-observer.py --preset top

    # 聚焦到 Gov Hall
    web-observer.py -x -12 -z -12 -d 15

    # 自定义角度
    web-observer.py -x 0 -z 0 -d 50 -a 0 -p 30
"""

from playwright.sync_api import sync_playwright
import time
import sys
import argparse


def wait_for_debug_controller(page, timeout=30):
    """等待 Godot Web 中的 DebugController 初始化"""
    print(f"→ 等待 DebugController 初始化...")
    
    for i in range(timeout):
        try:
            ready = page.evaluate("() => window.agentTownDebug && window.agentTownDebug.isReady()")
            if ready:
                print(f"✓ DebugController 已就绪")
                return True
        except:
            pass
        
        time.sleep(1)
        if i % 5 == 0 and i > 0:
            print(f"  等待中... ({i}s)")
    
    print("✗ 等待超时，DebugController 未就绪")
    return False


def set_camera_preset(page, preset_name):
    """设置相机预设"""
    result = page.evaluate(f"() => window.agentTownDebug.setPreset('{preset_name}')")
    print(f"  预设 '{preset_name}' 已应用")
    return result


def set_camera(page, target, distance, azimuth, polar):
    """设置相机参数"""
    script = f"""
    () => window.agentTownDebug.setCamera(
        {target[0]}, {target[1]}, {target[2]},
        {distance}, {azimuth}, {polar}
    )
    """
    result = page.evaluate(script)
    print(f"  相机已设置: target={target}, distance={distance}, azimuth={azimuth}, polar={polar}")
    return result


def capture(
    output="/tmp/web_observer.png",
    url="http://localhost:8080",
    width=1280,
    height=720,
    wait_seconds=15,
    camera_preset=None,
    camera_target=None,
    camera_distance=None,
    camera_azimuth=None,
    camera_polar=None
):
    """
    截图主函数
    
    流程:
    1. 打开浏览器
    2. 等待 Godot Web 加载
    3. 等待 DebugController 初始化
    4. 设置相机位置
    5. 等待渲染稳定
    6. 截图
    """
    
    with sync_playwright() as p:
        # 必须使用 headed 模式
        browser = p.chromium.launch(headless=False)
        page = browser.new_page(viewport={"width": width, "height": height})
        
        # 访问页面
        print(f"→ 正在访问 {url}...")
        page.goto(url, wait_until="networkidle", timeout=60000)
        
        # 等待 Godot WASM 初始加载
        print(f"→ 等待页面加载（{wait_seconds}秒）...")
        page.wait_for_timeout(wait_seconds * 1000)
        
        # 等待 DebugController 初始化
        if not wait_for_debug_controller(page, timeout=30):
            browser.close()
            return False
        
        # 设置相机位置
        print("→ 设置相机位置...")
        if camera_preset:
            set_camera_preset(page, camera_preset)
        elif camera_target is not None:
            set_camera(
                page,
                camera_target,
                camera_distance or 25,
                camera_azimuth or -45,
                camera_polar or 60
            )
        
        # 等待相机移动并触发渲染（Godot Web 需要交互才能更新渲染）
        print("→ 等待渲染稳定（3秒）...")
        page.wait_for_timeout(3000)
        
        # 点击画布触发 Godot 渲染更新
        print("→ 触发渲染更新...")
        page.click("canvas", position={"x": width // 2, "y": height // 2})
        page.wait_for_timeout(1000)
        
        # 截图
        print(f"→ 保存截图到 {output}...")
        page.screenshot(path=output, full_page=False)
        
        browser.close()
        print(f"✓ 截图已保存: {output}")
        return True


def main():
    parser = argparse.ArgumentParser(
        description="Web 观察工具 - 任意位置、任意视角截图",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
相机预设:
  top      - 俯视 (distance=30, azimuth=0, polar=10)
  side     - 侧面 (distance=25, azimuth=0, polar=90)
  north    - 从北边看 (distance=20, azimuth=90, polar=45)
  south    - 从南边看 (distance=20, azimuth=-90, polar=45)
  east     - 从东边看 (distance=20, azimuth=180, polar=45)
  west     - 从西边看 (distance=20, azimuth=0, polar=45)

示例:
  # 默认视角
  python web-observer.py

  # 俯视视角
  python web-observer.py --preset top

  # 聚焦到 Gov Hall (-12, -12)
  python web-observer.py -x -12 -z -12 -d 15

  # 自定义视角
  python web-observer.py -x 0 -z 0 -d 50 -a 0 -p 30 -o /tmp/custom.png
        """
    )
    
    parser.add_argument("-o", "--output", default="/tmp/web_observer.png", help="输出文件路径")
    parser.add_argument("-u", "--url", default="http://localhost:8080", help="目标 URL")
    parser.add_argument("-W", "--width", type=int, default=1280, help="视口宽度")
    parser.add_argument("-H", "--height", type=int, default=720, help="视口高度")
    parser.add_argument("-w", "--wait", type=int, default=15, help="初始加载等待时间（秒）")
    
    parser.add_argument("-x", "--target-x", type=float, help="目标位置 X")
    parser.add_argument("-y", "--target-y", type=float, default=1.5, help="目标位置 Y")
    parser.add_argument("-z", "--target-z", type=float, help="目标位置 Z")
    
    parser.add_argument("-d", "--distance", type=float, help="相机距离")
    parser.add_argument("-a", "--azimuth", type=float, help="水平角度（度）")
    parser.add_argument("-p", "--polar", type=float, help="垂直角度（度）")
    
    parser.add_argument("--preset", choices=["top", "side", "north", "south", "east", "west"],
                        help="使用预设视角")
    
    args = parser.parse_args()
    
    # 构建相机参数
    camera_target = None
    if args.target_x is not None or args.target_z is not None:
        camera_target = [
            args.target_x if args.target_x is not None else 0,
            args.target_y,
            args.target_z if args.target_z is not None else 0
        ]
    
    # 执行截图
    success = capture(
        output=args.output,
        url=args.url,
        width=args.width,
        height=args.height,
        wait_seconds=args.wait,
        camera_preset=args.preset,
        camera_target=camera_target,
        camera_distance=args.distance,
        camera_azimuth=args.azimuth,
        camera_polar=args.polar
    )
    
    if not success:
        sys.exit(1)


if __name__ == "__main__":
    main()
