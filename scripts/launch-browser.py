#!/usr/bin/env python3
"""Launch browser for debug environment

调用 skill 目录中的实现：
    .agents/skills/godot-web-debug/bin/launch-browser.py
"""

import subprocess
import sys
from pathlib import Path

def main():
    # 找到 skill 目录中的实现
    script_dir = Path(__file__).parent
    skill_script = script_dir.parent / ".agents" / "skills" / "godot-web-debug" / "bin" / "launch-browser.py"
    
    if not skill_script.exists():
        print(f"Error: Skill script not found: {skill_script}", file=sys.stderr)
        print("Please ensure godot-web-debug skill is installed.", file=sys.stderr)
        sys.exit(1)
    
    # 执行 skill 中的脚本
    result = subprocess.run([sys.executable, str(skill_script)] + sys.argv[1:])
    sys.exit(result.returncode)

if __name__ == "__main__":
    main()
