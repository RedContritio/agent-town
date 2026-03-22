#!/usr/bin/env python3
"""
Debug Client - 连接 debug_server 控制 godot-web 相机并截图

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

import asyncio
import websockets
import json
import argparse
import base64
import sys
from pathlib import Path

SERVER_URL = "ws://localhost:8081/client"


async def send_command(ws, msg_type: str, payload: dict = None):
    """发送命令到 server"""
    msg = {"type": msg_type}
    if payload:
        msg["payload"] = payload
    await ws.send(json.dumps(msg))


async def wait_for_godot_ready(ws, timeout=30):
    """等待 godot 就绪"""
    print("Waiting for godot-web to connect and initialize...")
    start_time = asyncio.get_event_loop().time()
    
    while asyncio.get_event_loop().time() - start_time < timeout:
        try:
            response = await asyncio.wait_for(ws.recv(), timeout=1.0)
            data = json.loads(response)
            
            if data.get("type") == "ready":
                print("✓ Godot is ready")
                return True
            elif data.get("type") == "connected":
                print("✓ Connected to server")
            
        except asyncio.TimeoutError:
            continue
    
    return False


async def wait_for_response(ws, timeout=10):
    """等待响应"""
    try:
        response = await asyncio.wait_for(ws.recv(), timeout=timeout)
        return json.loads(response)
    except asyncio.TimeoutError:
        print("Error: Timeout waiting for response")
        return None


async def cmd_info(args):
    """获取相机信息"""
    async with websockets.connect(SERVER_URL) as ws:
        # 等待 godot 就绪
        if not await wait_for_godot_ready(ws):
            print("Error: Timeout waiting for godot")
            return
        
        await send_command(ws, "get_camera_info")
        response = await wait_for_response(ws)
        
        if response and response.get("type") == "camera_info":
            info = response["payload"]
            print(f"Camera Info:")
            print(f"  Target: [{info['target'][0]:.2f}, {info['target'][1]:.2f}, {info['target'][2]:.2f}]")
            print(f"  Distance: {info['distance']:.2f}")
            print(f"  Azimuth: {info['azimuth_deg']:.2f}°")
            print(f"  Polar: {info['polar_deg']:.2f}°")
        else:
            print(f"Unexpected response: {response}")


async def cmd_set(args):
    """设置相机参数"""
    async with websockets.connect(SERVER_URL) as ws:
        # 等待 godot 就绪
        if not await wait_for_godot_ready(ws):
            print("Error: Timeout waiting for godot")
            return
        
        payload = {
            "target": [args.x, args.y, args.z],
            "distance": args.distance,
            "azimuth": args.azimuth,
            "polar": args.polar
        }
        
        await send_command(ws, "set_camera", payload)
        response = await wait_for_response(ws)
        
        if response and response.get("type") == "ack":
            print(f"Camera set successfully")
        elif response and response.get("type") == "error":
            print(f"Error: {response['payload'].get('message')}")
        else:
            print(f"Unexpected response: {response}")


async def cmd_preset(args):
    """使用预设视角"""
    async with websockets.connect(SERVER_URL) as ws:
        # 等待 godot 就绪
        if not await wait_for_godot_ready(ws):
            print("Error: Timeout waiting for godot")
            return
        
        await send_command(ws, "set_preset", {"preset": args.name})
        response = await wait_for_response(ws)
        
        if response and response.get("type") == "ack":
            print(f"Preset '{args.name}' applied")
        elif response and response.get("type") == "error":
            print(f"Error: {response['payload'].get('message')}")
        else:
            print(f"Unexpected response: {response}")


async def cmd_capture(args):
    """截图 - godot-web 保存到 /tmp/ 并返回路径"""
    async with websockets.connect(SERVER_URL) as ws:
        # 等待 godot 就绪
        if not await wait_for_godot_ready(ws):
            print("Error: Timeout waiting for godot")
            return
        
        payload = {
            "width": args.width,
            "height": args.height
        }
        
        await send_command(ws, "capture_screenshot", payload)
        response = await wait_for_response(ws)
        
        if response and response.get("type") == "screenshot":
            data = response["payload"]
            filepath = data["filepath"]
            print(f"Screenshot saved on godot-web: {filepath}")
            print(f"Size: {data['width']}x{data['height']}")
            
            # 如果指定了本地输出路径，提示用户需要手动复制
            if args.output != "screenshot.png":
                print(f"Note: File is on godot-web host, not saved locally as {args.output}")
        elif response and response.get("type") == "error":
            print(f"Error: {response['payload'].get('message')}")
        else:
            print(f"Unexpected response: {response}")


def main():
    parser = argparse.ArgumentParser(
        description="Debug Client - 控制 godot-web 相机并截图",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例:
  # 获取相机信息
  python client.py info

  # 设置相机位置 (目标 0,1.5,0，距离30米，水平角45°，垂直角60°)
  python client.py set -x 0 -z 0 -d 30 -a 45 -p 60

  # 使用预设视角
  python client.py preset --name top

  # 截图 (保存到 godot-web 主机的 /tmp/)
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
    
    # preset 命令
    preset_parser = subparsers.add_parser("preset", help="使用预设视角")
    preset_parser.add_argument("--name", required=True, 
                               choices=["top", "side", "north", "south", "east", "west"],
                               help="预设名称")
    
    # capture 命令
    capture_parser = subparsers.add_parser("capture", help="截图 (保存到 godot-web 主机的 /tmp/)")
    capture_parser.add_argument("-o", "--output", default="screenshot.png", help="(已废弃) 截图保存在 godot-web 主机")
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
        "preset": cmd_preset,
        "capture": cmd_capture,
    }
    
    try:
        asyncio.run(commands[args.command](args))
    except KeyboardInterrupt:
        print("\nInterrupted")
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
