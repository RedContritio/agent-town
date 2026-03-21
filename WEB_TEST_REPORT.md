# Godot Web 功能测试报告

## 测试环境
- **日期**: 2026-03-22
- **服务器**: http://localhost:8080
- **Godot 版本**: 4.5.stable.official.876b29033

## 实现功能

### 1. API 客户端扩展 (`scripts/api_client.gd`)
- ✅ 认证状态管理 (`auth_token`, `current_agent_id`)
- ✅ 认证 API (`/agents/me`, `/agents/me/status`)
- ✅ 任务栈 API (`/agents/me/tasks`)
- ✅ 任务操作 (`create_task`, `pause_task`, `resume_task`, `drop_task`)
- ✅ 自动刷新机制 (每 5 秒)

### 2. Agent 面板 (`scenes/ui/agent_panel.tscn`)
- ✅ Agent 信息显示 (名称、位置、HP、体力、余额)
- ✅ 任务栈显示 (LIFO 列表)
- ✅ 任务操作按钮 (暂停/恢复/放弃)
- ✅ 创建任务对话框
- ✅ 点击 Agent 实体显示面板

### 3. 登录面板 (`scenes/ui/login_panel.tscn`)
- ✅ Token 输入框
- ✅ 登录/登出功能
- ✅ 状态显示

### 4. 主场景更新 (`scenes/main.tscn`)
- ✅ 集成 AgentPanel (右上角)
- ✅ 集成 LoginPanel (左下角)
- ✅ 保持现有 HUD

## API 端点测试

### 公开端点
```bash
# 世界信息
GET /api/v1/world/info
# {"id":"world-1774117329","name":"Agent Town","seed":"123456789",...}

# 命令列表 (用于 CLI 发现)
GET /api/v1/commands
# {"commands":[{"name":"status","type":"sync",...},...]}
```

### 需要认证的端点
```bash
# Agent 信息
GET /api/v1/agents/me

# Agent 状态
GET /api/v1/agents/me/status

# 任务栈
GET /api/v1/agents/me/tasks

# 创建任务
POST /api/v1/agents/me/tasks
BODY: {"type": 0, "params": {"dx": 1, "dy": 0}}

# 暂停任务
POST /api/v1/agents/me/tasks/{task_id}/pause

# 恢复任务
POST /api/v1/agents/me/tasks/{task_id}/resume

# 放弃任务
DELETE /api/v1/agents/me/tasks/{task_id}
```

## 使用流程

### 1. 启动服务器
```bash
make run
# Server running at http://localhost:8080
```

### 2. 打开 Web 界面
```
http://localhost:8080
```

### 3. CLI 登录获取 Token
```bash
# 创建 Agent
./bin/cli agent create --name Alice

# 注册并获取 Token
./bin/cli --agent Alice --server http://localhost:8080 status
# 首次需要完成挑战-响应流程...
```

### 4. Web 端操作
1. 在 LoginPanel 输入 Token
2. 点击 Agent 查看详细信息
3. 在 AgentPanel 查看任务栈
4. 创建/暂停/恢复/放弃任务

## Godot Web 导出

```bash
export PATH=$PATH:/snap/bin
make web
# → Exporting web frontend...
# ✓ Web exported
```

导出文件:
- `server/cmd/server/web/index.html`
- `server/cmd/server/web/index.js`
- `server/cmd/server/web/index.wasm`
- `server/cmd/server/web/index.pck`

## 已知限制

1. **认证流程**: Web 端需要 CLI 生成的 Token，暂不支持直接在 Web 端注册
2. **任务执行**: 任务仅记录状态，实际执行逻辑需后续实现
3. **实时更新**: 依赖 5 秒轮询，非 WebSocket 实时推送

## 后续建议

1. 实现 WebSocket 实时推送
2. 添加任务进度可视化
3. 集成地图点击交互
4. 添加 Agent 间社交功能
