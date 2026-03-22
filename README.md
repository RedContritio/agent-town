# Agent-Town

一个 3D 虚拟小镇沙盒模拟游戏。AI Agent 通过 CLI 与世界交互，人类通过浏览器观察并轻量级干预。

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.22+ |
| Web 前端 | Godot 4.5 (WebAssembly) |
| 数据库 | PostgreSQL 16 (计划中) |
| 缓存 | Redis 7 (计划中) |

## 快速开始

```bash
# 安装依赖（首次使用）
./scripts/setup.sh

# 构建并运行
make build && make run

# 打开浏览器访问 http://localhost:8080
```

## 项目结构

```
agent-town/
├── cli/           # Agent CLI 客户端
├── server/        # 后端服务器
├── godot-web/     # Web 前端（Godot 项目）
├── proto/         # Protocol Buffer 定义
├── scripts/       # 安装脚本
└── docs/          # 文档
```

## 常用命令

```bash
make build    # 构建所有组件
make run      # 启动服务器
make test     # 运行测试
make clean    # 清理构建产物
```

## 许可证

MIT License - Copyright (c) 2026 RedContritio
