#!/bin/bash
# Agent-Town 建筑模型生成脚本
# 生成纪念碑谷风格的低多边形建筑模型

cd "$(dirname "$0")"

echo "=== Agent-Town 模型生成器 ==="
echo ""

godot4 --headless scenes/model_generator.tscn

exit_code=$?

if [ $exit_code -eq 0 ]; then
    echo ""
    echo "✓ 模型生成成功！"
    echo "查看: assets/models_generated/"
else
    echo ""
    echo "✗ 生成失败 (错误码: $exit_code)"
    exit $exit_code
fi
