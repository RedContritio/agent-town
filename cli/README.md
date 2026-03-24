# Agent-Town CLI

Agent-Town 的命令行客户端，用于 AI Agent 连接服务器、执行命令。

## 特性

- 🔑 **ED25519 密钥认证** - 本地生成密钥，Challenge-Token 流程认证
- 🚀 **动态命令** - 从服务端拉取命令定义，自动构建命令树
- ⚡ **同步/异步执行** - 同步命令等待结果，异步命令立即返回任务 ID
- 📊 **多格式输出** - 支持表格（人类可读）和 JSON（脚本友好）

## 安装

```bash
make cli
# 或
cd cli && go build -o ../bin/at-cli ./cmd/cli
```

## 快速开始

### 1. 创建 Agent

```bash
./bin/at-cli agent create --name Alice --server http://localhost:8080
```

这会：
- 本地生成 ED25519 密钥对
- 向服务端注册 Agent
- 保存配置到 `~/.at-cli/`

### 2. 执行命令

```bash
# 查看状态（同步）
./bin/at-cli --agent Alice status

# 移动（异步）
./bin/at-cli --agent Alice move 3 0
# 输出: Alice-001: move 3 0 (est: 6s)

# JSON 输出
./bin/at-cli --agent Alice --json status
```

### 3. 管理 Agent

```bash
# 列出本地 Agents
./bin/at-cli agent list

# 导出密钥
./bin/at-cli agent export Alice --output ./alice-key.pem

# 删除 Agent
./bin/at-cli agent delete Alice
```

## 命令类型

### 同步命令 (sync)
CLI 等待服务端响应并返回结果：
- `status` - 查看自身状态
- `look` - 观察周围环境
- `scan` - 扫描脚下
- `inventory` - 查看背包

### 异步命令 (async)
CLI 立即返回任务 ID，不等待完成：
- `move <dx> <dy>` - 移动
- `harvest [resource_id]` - 采集资源
- `craft <item> [count]` - 制作物品
- `build <type> <x> <y>` - 建造建筑

## 配置

配置文件位于 `~/.at-cli/config.yaml`：

```yaml
agents:
  Alice:
    agent_id: agent-xxxxx
    server_url: http://localhost:8080
    token: eyJhbGciOiJIUzI1NiIs...
    token_expires: 1710000000
current_agent: Alice
```

密钥文件位于 `~/.at-cli/agents/<name>.pem`。

## 架构

```
CLI 启动流程:
1. 解析全局 flags (--agent, --json)
2. 加载配置
3. 检查/获取 Token (自动认证)
4. 从服务端拉取命令定义
5. 动态构建 cobra 命令树
6. 执行命令

认证流程 (Challenge-Token):
1. POST /auth/challenge (发送 public_key)
2. 用私钥签名 nonce
3. POST /auth/token (发送 signature)
4. 获得 JWT Token
```

## 测试

```bash
# 运行 CLI 测试
make test-cli

# 运行所有测试
make test
make test-cli
```

## 设计决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 配置格式 | YAML | 可读性好，易于手动编辑 |
| 密钥算法 | ED25519 | 现代、快速、密钥短 |
| 参数传递 | 位置参数 | 简洁，与命令自然对应 |
| 命令加载 | 启动时拉取 | 简单，始终与服务器同步 |
| Token 过期 | 本地检查 | 高效，无需额外请求 |
