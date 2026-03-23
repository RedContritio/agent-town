#!/usr/bin/env python3
"""
Debug Client - 通过 HTTP API 控制 Godot 相机并截图

用法:
    python client.py [命令] [选项]

命令:
    info                    获取相机信息
    set                     设置相机参数
    preset                  使用预设视角
    capture                 截图

示例:
    # 获取相机信息
    python client.py info

    # 设置相机位置
    python client.py set -x 0 -z 0 -d 30 -a 45 -p 60

    # 使用预设视角 (top, side, north, south, east, west)
    python client.py preset --name top

    # 截图
    python client.py capture -o screenshot.png

    # 组合: 设置视角并截图
    python client.py preset --name north && python client.py capture -o north.png
"""

import requests
import argparse
import json
import sys
import time
from pathlib import Path

BASE_URL = "http://localhost:8081"


def wait_for_godot(timeout=30):
    """等待 godot 就绪"""
    print("Waiting for Godot to connect and initialize...")
    start_time = time.time()
    
    while time.time() - start_time < timeout:
        try:
            resp = requests.get(f"{BASE_URL}/health", timeout=2)
            data = resp.json()
            if data.get("godot_ready"):
                print("✓ Godot is ready")
                return True
        except Exception as e:
            pass
        time.sleep(1)
    
    print("✗ Timeout waiting for godot")
    return False


def cmd_info(args):
    """获取相机信息"""
    if not wait_for_godot():
        return
    
    try:
        resp = requests.get(f"{BASE_URL}/api/info", timeout=10)
        if resp.status_code == 200:
            info = resp.json()
            print("Camera Info:")
            print(f"  Target: [{info['target'][0]:.2f}, {info['target'][1]:.2f}, {info['target'][2]:.2f}]")
            print(f"  Distance: {info['distance']:.2f}")
            print(f"  Azimuth: {info['azimuth_deg']:.2f}°")
            print(f"  Polar: {info['polar_deg']:.2f}°")
        else:
            print(f"Error: {resp.status_code} - {resp.text}")
    except Exception as e:
        print(f"Error: {e}")


def cmd_set(args):
    """设置相机参数"""
    if not wait_for_godot():
        return
    
    payload = {
        "target": [args.x, args.y, args.z],
        "distance": args.distance,
        "azimuth": args.azimuth,
        "polar": args.polar
    }
    
    try:
        resp = requests.post(f"{BASE_URL}/api/camera", json=payload, timeout=10)
        if resp.status_code == 200:
            print("✓ Camera set successfully")
        else:
            data = resp.json()
            print(f"Error: {data.get('error', 'unknown')}")
    except Exception as e:
        print(f"Error: {e}")


def cmd_direct(args):
    """直接设置相机位置（绕过轨道控制器，用于平视视角）"""
    if not wait_for_godot():
        return
    
    payload = {
        "position": [args.px, args.py, args.pz],
        "look_at": [args.lx, args.ly, args.lz]
    }
    
    try:
        resp = requests.post(f"{BASE_URL}/api/camera/direct", json=payload, timeout=10)
        if resp.status_code == 200:
            print(f"✓ Camera set directly: position=({args.px}, {args.py}, {args.pz}) -> look_at=({args.lx}, {args.ly}, {args.lz})")
        else:
            data = resp.json()
            print(f"Error: {data.get('error', 'unknown')}")
    except Exception as e:
        print(f"Error: {e}")


def cmd_preset(args):
    """使用预设视角"""
    if not wait_for_godot():
        return
    
    try:
        resp = requests.post(f"{BASE_URL}/api/preset", json={"name": args.name}, timeout=10)
        if resp.status_code == 200:
            print(f"✓ Preset '{args.name}' applied")
        else:
            data = resp.json()
            print(f"Error: {data.get('error', 'unknown')}")
    except Exception as e:
        print(f"Error: {e}")


def cmd_capture(args):
    """截图 - 保存到 /tmp/ 并返回路径"""
    if not wait_for_godot():
        return
    
    payload = {
        "width": args.width,
        "height": args.height
    }
    
    try:
        resp = requests.post(f"{BASE_URL}/api/capture", json=payload, timeout=15)
        if resp.status_code == 200:
            data = resp.json()
            filepath = data.get("filepath", "unknown")
            print(f"✓ Screenshot saved: {filepath}")
            print(f"  Size: {data.get('width')}x{data.get('height')}")
        else:
            data = resp.json()
            print(f"Error: {data.get('error', 'unknown')}")
    except Exception as e:
        print(f"Error: {e}")


def main():
    parser = argparse.ArgumentParser(
        description="Debug Client - 控制 Godot 相机并截图",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例:
  # 获取相机信息
  python client.py info

  # 设置相机位置 (目标 0,1.5,0，距离30米，水平角45°，垂直角60°)
  python client.py set -x 0 -z 0 -d 30 -a 45 -p 60

  # 使用预设视角
  python client.py preset --name top

  # 直接设置相机位置（平视视角）
  python client.py direct --px -12 --py 3 --pz -20 --lx -12 --ly 1.5 --lz -12

  # 截图 (保存到 Godot 主机的 /tmp/)
  python client.py capture
        """
    )
    
    subparsers = parser.add_subparsers(dest="command", help="可用命令")
    
    # info 命令
    subparsers.add_parser("info", help="获取相机信息")
    
    # set 命令
    set_parser = subparsers.add_parser("set", help="设置相机参数")
    set_parser.add_argument("-x", type=float, default=0, help="目标位置 X")
    set_parser.add_argument("-y", type=float, default=1.5, help="目标位置 Y")
    set_parser.add_argument("-z", type=float, default=0, help="目标位置 Z")
    set_parser.add_argument("-d", "--distance", type=float, default=25, help="相机距离")
    set_parser.add_argument("-a", "--azimuth", type=float, default=-45, help="水平角度 (度)")
    set_parser.add_argument("-p", "--polar", type=float, default=60, help="垂直角度 (度)")
    
    # direct 命令（直接设置相机位置）
    direct_parser = subparsers.add_parser("direct", help="直接设置相机位置（用于平视视角）")
    direct_parser.add_argument("--px", type=float, required=True, help="相机位置 X")
    direct_parser.add_argument("--py", type=float, required=True, help="相机位置 Y（高度）")
    direct_parser.add_argument("--pz", type=float, required=True, help="相机位置 Z")
    direct_parser.add_argument("--lx", type=float, default=0, help="看向位置 X")
    direct_parser.add_argument("--ly", type=float, default=1.5, help="看向位置 Y（高度）")
    direct_parser.add_argument("--lz", type=float, default=0, help="看向位置 Z")
    
    # preset 命令
    preset_parser = subparsers.add_parser("preset", help="使用预设视角")
    preset_parser.add_argument("--name", required=True, 
                               choices=["top", "side", "north", "south", "east", "west"],
                               help="预设名称")
    
    # capture 命令
    capture_parser = subparsers.add_parser("capture", help="截图 (保存到 /tmp/)")
    capture_parser.add_argument("-o", "--output", default="screenshot.png", help="(已废弃) 截图保存在 Godot 主机")
    capture_parser.add_argument("-W", "--width", type=int, default=1280, help="宽度")
    capture_parser.add_argument("-H", "--height", type=int, default=720, help="高度")
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        sys.exit(1)
    
    # 运行对应命令
    commands = {
        "info": cmd_info,
        "set": cmd_set,
        "direct": cmd_direct,
        "preset": cmd_preset,
        "capture": cmd_capture,
    }
    
    try:
        commands[args.command](args)
    except KeyboardInterrupt:
        print("\nInterrupted")
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
