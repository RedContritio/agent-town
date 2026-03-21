# at-cli 命令行工具设计文档

## 设计理念

### 1. 执行模型

**所有操作都压入 LIFO 任务栈**，CLI 行为根据操作类型而异：

- **同步操作**：CLI 等待任务完成并出栈，然后返回结果
- **异步操作**：CLI 立即返回任务 id，不等待完成

```bash
# 同步操作 - 压栈执行，CLI 等待完成后返回结果
$ at-cli --agent Alice look
[look] running...
[look] done: [tree-001] 橡树 (10,20), [rock-003] 铁矿 (12,21)

# 异步操作 - 压栈执行，CLI 立即返回任务 id
$ at-cli --agent Alice move 3,0
Alice-001: move 3,0 (est: 6s)

# 新异步命令压栈，旧任务暂停
$ at-cli --agent Alice harvest tree-001
Alice-002: harvest tree-001 (est: 5m)  # Alice-001 暂停

$ at-cli --agent Alice stack
DEPTH  TASK              STATUS      EST
0      harvest tree-001  RUNNING     5m
1      move 3,0          PAUSED      6s
```

**任务 ID 规则**：
- 与 Agent 绑定，由 Server 分配和维护
- 格式：`<agent-name>-<sequence>`（如 `Alice-001`）
- 每个 Agent 有自己的 ID 池，在 Agent 维度唯一

**核心原则**：
- **完全用户控制**：用户现在想做什么，就立即做什么
- **最近意图优先**：新命令立即压栈执行，旧任务自动暂停
- **自动恢复**：栈顶任务完成后，下层任务自动恢复
- **不干预**：系统不替用户决定优先级，"饿死"旧任务是用户的选择

### 2. 任务类型

| 类型 | CLI 行为 | 示例 |
|------|----------|------|
| **同步** | CLI 等待任务完成并返回结果 | look, scan, inventory, stack, time |
| **异步** | 立即返回任务 id，不等待 | move, harvest, craft, build, combat |

### 3. 双层结构

| 系统 | 内容 | 机制 |
|------|------|------|
| **任务栈（Doing）** | 我正在做的事（改变世界）| LIFO 压栈执行 |
| **收件箱（Inbox）** | 别人在等我的事（需响应）| 独立列表，响应动作进栈 |

### 4. 交易统一模型

所有经济交互统一为 **"挂单-响应"** 模型：

| 对象 | 撮合方式 | 说明 |
|------|----------|------|
| **market** | 自动撮合 | 公开挂单，价格匹配即成交 |
| **gov** | 固定价格 | 向政府买地、卖物资 |
| **<agent>** | 显式响应 | 定向交易，需对方 inbox 响应 |

### 5. 土地-建筑分离

- **土地**：地块所有权，可单独交易
- **建筑**：依附于土地但产权独立，可单独交易
- **经济循环**：建造付土地费 → 建筑损耗 → 维修再付土地费给**当前**土地主

---

## 命令系统

每个命令必须显式指定 `--agent` 参数（Agent 名称），CLI 从默认位置加载密钥文件，无会话状态。

### Agent 管理（本地操作，无 Server 交互）

```bash
at-cli agent create --name <name>                    # 生成本地密钥对，存储于 ~/.at-cli/agents/<name>.pem
at-cli agent list                                     # 列出本地 Agent 名称
at-cli agent export <name> [--output <path>]          # 导出 Agent 密钥文件
at-cli agent delete <name>                            # 删除本地 Agent
```

### 状态感知（全部同步）

```bash
at-cli --agent <name> status         # 自身状态（同步）
at-cli --agent <name> look           # 观察周围（同步）
at-cli --agent <name> scan           # 扫描脚下（同步）
at-cli --agent <name> inventory      # 查看背包（同步）
at-cli --agent <name> skills         # 查看技能（同步）
at-cli --agent <name> balance        # 查看余额（同步）
at-cli --agent <name> time           # 世界时间（同步）
```

### 栈管理（全部同步）

```bash
at-cli --agent <name> stack                      # 查看当前任务栈（同步）
at-cli --agent <name> stack focus <depth>        # 跳转到指定深度任务（同步）
at-cli --agent <name> stack drop <depth>         # 放弃指定深度任务（同步）
at-cli --agent <name> stack clear                # 清空整个栈（危险，同步）
```

**栈输出示例**：
```
DEPTH  TASK              STATUS      EST
0      harvest tree-001  RUNNING     5m
1      move 3,0          PAUSED      6s
2      build house       PAUSED      4h
```

### 移动（异步）

```bash
at-cli --agent <name> move <dx>,<dy>             # 向相对坐标移动（异步）

# 示例:
# move 1,0   向东1格
# move 0,-1  向北1格
# move 3,0   向东3格
# move -2,2  向西北2格
```

### 资源采集（异步）

```bash
at-cli --agent <name> harvest                    # 采集脚下最近资源（异步）
at-cli --agent <name> harvest <resource-id>      # 采集指定资源（异步）
at-cli --agent <name> harvest --deep             # 深入采集（异步）
```

### 物品使用

```bash
at-cli --agent <name> use <item>                 # 使用物品（异步）
at-cli --agent <name> use --slot <n>             # 使用背包第n格（异步）
at-cli --agent <name> equip <item>              # 装备工具（同步）
at-cli --agent <name> equip --slot <n>          # 装备背包第n格（同步）
at-cli --agent <name> drop <item> [count]       # 丢弃物品（同步）
at-cli --agent <name> take                       # 捡起脚下物品（异步）
```

### 制作与建造

```bash
at-cli --agent <name> craft list                 # 查看可制作列表（同步）
at-cli --agent <name> craft <item>               # 制作物品（异步）
at-cli --agent <name> build --type <type> --at <x>,<y> [--name <name>]   # 建造建筑（异步）
at-cli --agent <name> repair <building-id>       # 维修建筑（异步）
at-cli --agent <name> demolish <building-id>     # 拆除建筑（异步）
```

### 土地

```bash
at-cli --agent <name> land                       # 查看我的土地列表（同步）
at-cli --agent <name> land <x>,<y>               # 查看特定地块详情（同步）
at-cli --agent <name> trade gov buy land --at <x>,<y>    # 向政府购买土地（同步）
at-cli --agent <name> trade offer sell land --at <x>,<y> --price <n>   # 挂卖土地（同步）
```

### 建筑

```bash
at-cli --agent <name> building                   # 查看我的建筑列表（同步）
at-cli --agent <name> building <id>              # 查看建筑详情（同步）
at-cli --agent <name> trade offer sell building <id> --price <n>   # 挂卖建筑（同步）
```

### 交易

```bash
# 查看市场（同步）
at-cli --agent <name> trade market
at-cli --agent <name> trade market <item>

# 挂单（同步）
at-cli --agent <name> trade offer buy <item> <count> --price <n> [--to <agent>]
at-cli --agent <name> trade offer sell <item> <count> --price <n> [--to <agent>]

# 查看我的挂单（同步）
at-cli --agent <name> trade offers

# 取消挂单（同步）
at-cli --agent <name> trade cancel <order-id>
```

**定向交易响应**（从 inbox 发起）：
```bash
at-cli --agent <name> inbox                      # 查看收件箱（同步）
at-cli --agent <name> inbox do <id>              # 接受/执行交易（异步）
at-cli --agent <name> inbox skip <id>            # 拒绝交易（同步）
at-cli --agent <name> inbox info <id>            # 查看详情（同步）
```

### 社交（全部同步）

```bash
at-cli --agent <name> contact list               # 视野内 Agent 列表（同步）
at-cli --agent <name> contact <agent>            # 查看 Agent 信息（同步）
at-cli --agent <name> contact <agent> --say <msg> # 附近广播（同步）
at-cli --agent <name> contact <agent> --msg <msg> # 私聊留言（同步）
at-cli --agent <name> contact <agent> --rate <0-10> # 评分（同步）
```

### 战斗

```bash
# 发起挑战（同步，进入对方 inbox）
at-cli --agent <name> combat offer <agent> --bet <amount>

# 响应挑战（从 inbox 发起）
at-cli --agent <name> inbox do <challenge-id>    # 接受挑战（异步，进入战斗）
at-cli --agent <name> inbox skip <challenge-id>   # 拒绝挑战（同步）

# 战斗中动作（异步，提交后等待对方响应或超时）
at-cli --agent <name> combat attack              # 攻击（异步）
at-cli --agent <name> combat defend              # 防御（异步）
at-cli --agent <name> combat flee                # 逃跑（异步）
at-cli --agent <name> combat use --slot <n>      # 使用物品（异步）
```

---

## 土地-建筑经济模型

### 核心循环

```
1. 买地（从 gov 或市场）
2. 建房（付一次性土地费给土地主）
3. 使用（建筑自然损耗）
4. 维修（付维修土地费给当前土地主）
```

### 土地转手的影响

```
原始: Alice 拥有土地，Bob 在土地上建房（付土地费给 Alice）
变化: Alice 把地卖给 Carol
结果: Carol 成为新土地主，Bob 维修时付费给 Carol
```

### 房地产商模式

```bash
# 1. 买地
at-cli --agent alice.pem trade gov buy land --at 10,20

# 2. 建房（付土地费给自己 = 免）
at-cli --agent alice.pem build --type apartment --at 10,20 --floors 3

# 3. 卖建筑单元（保留土地）
at-cli --agent alice.pem trade offer sell building unit-10-20-2F --price 1000

# 4. 未来收维修土地费（被动收入）
```

---

## 损耗与维修

### 损耗速度（数值待调）

- 建筑耐久度随时间自然下降
- 具体损耗速度待平衡测试后确定
- 方向：慢损耗，给 casual 玩家充足缓冲

### 耐久度影响

| 耐久度 | 状态 | 影响 |
|--------|------|------|
| 100-80 | 优秀 | 正常功能 |
| 80-50 | 良好 | 正常功能 |
| 50-20 | 老化 | 效率下降（生产建筑产出减少） |
| 20-0 | 危险 | 功能受限，可能损坏物品 |
| 0 | 倒塌 | 建筑消失，内部物品丢失 |

### 维修

- 消耗材料（根据建筑类型）
- 支付土地维修费（给当前土地主）
- 耐久度恢复至 100

---

## 错误处理

1. **明确失败原因**："Insufficient wood: need 50, have 15"
2. **提供下一步提示**："Run 'at-cli craft list' to see requirements"
3. **原子性**：命令要么完全成功，要么完全失败
4. **资源不退**：取消/放弃任务，已消耗资源不退回

---

## 待决策事项

1. **方向简写**：是否支持 `at-cli move n` 代替 `north`
2. **损耗速度**：具体数值需平衡测试
3. **维修材料比例**：维修所需材料 vs 新建材料的比率
4. **建筑倒塌**：耐久度归零时是否保留部分材料
5. **栈深度限制**：是否限制最大深度（如 100）防止内存爆炸
6. **持久化**：CLI 进程结束 = 任务栈消失（P3 直接结果，无需决策）
7. **服务器重启**：任务栈如何恢复
8. **提醒机制**：任务被暂停多久后提醒用户

---

## 使用示例

### 新玩家入门

```bash
# 创建 Agent
at-cli agent create --name Alice

# 注册并进入世界
at-cli --agent Alice --server https://town1.com status

# 观察周围
at-cli --agent Alice look
at-cli --agent Alice scan

# 采集脚下资源
at-cli --agent Alice harvest

# 采集指定资源（通过 look/scan 获取 resource-id）
at-cli --agent Alice harvest res-001

# 查看背包
at-cli --agent Alice inventory
```

### 建造房屋

```bash
# 买地
at-cli --agent alice.pem trade gov buy land --at 10,20

# 查看可建造
at-cli --agent alice.pem craft list

# 建造（压栈，长耗时）
at-cli --agent alice.pem build --type house --at 10,20 --name "Home"

# 期间查询（自动中断-恢复）
at-cli --agent alice.pem look
at-cli --agent alice.pem scan
at-cli --agent alice.pem inventory

# 查看建造进度
at-cli --agent alice.pem building

# 最终完成
# Welcome back. "Home" completed at 14:30
```

### 房地产商操作

```bash
# 批量买地
at-cli --agent alice.pem trade gov buy land --at 10,20
at-cli --agent alice.pem trade gov buy land --at 10,21

# 建造公寓
at-cli --agent alice.pem build --type apartment --at 10,20 --floors 3

# 出售建筑单元（保留土地）
at-cli --agent alice.pem trade offer sell building unit-10-20-2F --price 1000

# 等待未来业主维修时收取土地费
```

### 紧急响应

```bash
# 正在建造
at-cli --agent alice.pem build --type house --at 10,20

# 突然受到攻击
at-cli --agent alice.pem inbox
# inb-001: combat challenge from Bob (4m left)

# 接受挑战（暂停建造，进入战斗）
at-cli --agent alice.pem inbox do inb-001

# 战斗中
at-cli --agent alice.pem combat <battle-id> attack
at-cli --agent alice.pem combat <battle-id> defend

# 战斗结束，自动恢复建造
```

### 手动管理栈

```bash
# 查看栈
at-cli --agent alice.pem stack
# [scan, look, build house, move north]

# 决定放弃建造，专注移动
at-cli --agent alice.pem stack drop 2           # 移除 build house
at-cli --agent alice.pem stack focus 3          # 跳到 move north

# 清理完成
at-cli --agent alice.pem stack
# [move north]
```

---

## 更新日志

- **v1.0** (当前)：确定 LIFO 任务栈、土地-建筑分离、损耗+维修收费模型
- **v1.1**：移除所有会话状态命令（join/logout/switch），改为 `--agent` 参数模式（符合 P3 显式完整）
