# at-cli 命令行工具设计文档

## 设计理念

### 1. LIFO 任务栈（核心架构）

所有命令遵循 **"压栈执行，完成后出栈，恢复下层"** 的 LIFO（后进先出）机制：

```bash
# 执行流程示例
[10:00] at-cli build house --at 10,20    # 压栈 [build]，开始执行
[10:01] at-cli scan                       # 压栈 [scan, build]，暂停 build
[10:01.1] scan 完成出栈                   # 栈 [build]，自动恢复 build
[10:05] at-cli look                       # 压栈 [look, build]，暂停 build
[10:05.1] look 完成出栈                   # 栈 [build]，自动恢复 build
```

**核心原则**：
- **完全用户控制**：用户现在想做什么，就立即做什么
- **最近意图优先**：新命令立即执行，旧任务自动暂停
- **自动恢复**：栈顶任务完成后，下层任务自动恢复
- **不干预**：系统不替用户决定优先级，"饿死"旧任务是用户的选择

### 2. 任务类型

| 类型 | 时间 | 示例 |
|------|------|------|
| **瞬间** | ~0.1s - 5s | scan, look, inventory, move（1格） |
| **中等** | 5分钟 - 30分钟 | harvest（大资源）, chop, craft |
| **长时** | 30分钟 - 数小时 | build, repair, demolish, 长途移动 |

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

### 身份与连接

```bash
at-cli join --name <name> --server <url>    # 注册/登录
at-cli agents                                 # 列出本地身份
at-cli switch <agent-id>                      # 切换当前身份
```

### 状态感知（全部压栈）

```bash
at-cli status         # 自身状态
at-cli look           # 观察周围（0.1s）
at-cli scan           # 扫描脚下（0.1s）
at-cli inventory      # 查看背包（0.1s）
at-cli skills         # 查看技能（0.1s）
at-cli balance        # 查看余额（0.1s）
at-cli time           # 世界时间（0.1s）
```

### 栈管理

```bash
at-cli stack                      # 查看当前任务栈
at-cli stack --watch              # 持续监视刷新
at-cli stack focus <depth>        # 跳转到指定深度任务
at-cli stack drop <depth>         # 放弃指定深度任务
at-cli stack clear                # 清空整个栈（危险）
```

**栈输出示例**：
```
DEPTH  TASK             TYPE      STATUS      RUN_TIME    TOTAL_EST
0      scan             instant   RUNNING     0.05s       0.1s
1      look             instant   PAUSED      0.08s       0.1s
2      build house      long      PAUSED      12m         4h
```

### 移动（压栈）

```bash
at-cli move --dir <direction>     # 朝方向移动1格（2s）
at-cli move --to <x>,<y>          # 移动到坐标（距离1=2s，距离>1=长耗时）

# 方向: north, south, east, west, northeast, northwest, southeast, southwest
```

### 资源采集（压栈）

```bash
at-cli harvest                    # 采集脚下小资源（5s）
at-cli harvest --deep             # 深入采集（30m）
at-cli chop                       # 砍伐大树（20m）
at-cli mine                       # 开采矿脉（1h）
```

### 物品使用（压栈）

```bash
at-cli use <item>                 # 使用物品（1s）
at-cli use --slot <n>             # 使用背包第n格（1s）
at-cli equip <item>              # 装备工具（0.5s）
at-cli equip --slot <n>          # 装备背包第n格（0.5s）
at-cli drop <item> [count]       # 丢弃物品（0.1s）
at-cli take                       # 捡起脚下物品（0.5s）
```

### 制作与建造（压栈）

```bash
at-cli craft list                 # 查看可制作列表（0.1s）
at-cli craft <item>               # 制作物品（5m - 30m）
at-cli build --type <type> --at <x>,<y> [--name <name>]   # 建造建筑（30m - 4h）
at-cli repair <building-id>       # 维修建筑（10m - 1h）
at-cli demolish <building-id>     # 拆除建筑（20m - 2h）
```

### 土地（压栈）

```bash
at-cli land                       # 查看我的土地列表（0.1s）
at-cli land <x>,<y>               # 查看特定地块详情（0.1s）
at-cli trade gov buy land --at <x>,<y>    # 向政府购买土地（instant）
at-cli trade offer sell land --at <x>,<y> --price <n>   # 挂卖土地（instant）
```

### 建筑（压栈）

```bash
at-cli building                   # 查看我的建筑列表（0.1s）
at-cli building <id>              # 查看建筑详情（0.1s）
at-cli trade offer sell building <id> --price <n>   # 挂卖建筑（instant）
```

### 交易（压栈执行或进 inbox）

```bash
# 查看市场（0.1s）
at-cli trade market
at-cli trade market <item>

# 挂单（instant）
at-cli trade offer buy <item> <count> --price <n> [--to <agent>]
at-cli trade offer sell <item> <count> --price <n> [--to <agent>]

# 查看我的挂单（0.1s）
at-cli trade offers

# 取消挂单（instant）
at-cli trade cancel <order-id>
```

**定向交易响应**（从 inbox 发起）：
```bash
at-cli inbox                      # 查看收件箱（0.1s）
at-cli inbox do <id>              # 接受/执行交易（1s）
at-cli inbox skip <id>            # 拒绝交易（instant）
at-cli inbox info <id>            # 查看详情（0.1s）
```

### 社交（压栈）

```bash
at-cli contact list               # 视野内 Agent 列表（0.1s）
at-cli contact <agent>            # 查看 Agent 信息（0.1s）
at-cli contact <agent> --say <msg> # 附近广播（0.1s）
at-cli contact <agent> --msg <msg> # 私聊留言（0.1s）
at-cli contact <agent> --rate <0-10> # 评分（0.1s）
```

### 战斗（压栈）

```bash
# 发起挑战（instant，进入对方 inbox）
at-cli combat offer <agent> --bet <amount>

# 响应挑战（从 inbox 发起，进栈执行）
at-cli inbox do <challenge-id>    # 接受挑战（instant）
at-cli inbox skip <challenge-id>   # 拒绝挑战（instant）

# 战斗中动作（每个动作压栈）
at-cli combat attack              # 攻击（instant）
at-cli combat defend              # 防御（instant）
at-cli combat flee                # 逃跑（instant）
at-cli combat use --slot <n>      # 使用物品（instant）
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
at-cli trade gov buy land --at 10,20

# 2. 建房（付土地费给自己 = 免）
at-cli build --type apartment --at 10,20 --floors 3

# 3. 卖建筑单元（保留土地）
at-cli trade offer sell building unit-10-20-2F --price 1000

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
6. **持久化**：用户下线后，栈状态是否保存？下次登录是否恢复
7. **服务器重启**：任务栈如何恢复
8. **提醒机制**：任务被暂停多久后提醒用户

---

## 使用示例

### 新玩家入门

```bash
# 注册并进入世界
at-cli join --name Alice --server https://town1.com

# 查看状态
at-cli status

# 观察周围
at-cli look
at-cli scan

# 采集资源
at-cli harvest

# 查看背包
at-cli inventory
```

### 建造房屋

```bash
# 买地
at-cli trade gov buy land --at 10,20

# 查看可建造
at-cli craft list

# 建造（压栈，长耗时）
at-cli build --type house --at 10,20 --name "Home"

# 期间查询（自动中断-恢复）
at-cli look
at-cli scan
at-cli inventory

# 查看建造进度
at-cli building

# 最终完成
# Welcome back. "Home" completed at 14:30
```

### 房地产商操作

```bash
# 批量买地
at-cli trade gov buy land --at 10,20
at-cli trade gov buy land --at 10,21

# 建造公寓
at-cli build --type apartment --at 10,20 --floors 3

# 出售建筑单元（保留土地）
at-cli trade offer sell building unit-10-20-2F --price 1000

# 等待未来业主维修时收取土地费
```

### 紧急响应

```bash
# 正在建造
at-cli build --type house --at 10,20

# 突然受到攻击
at-cli inbox
# inb-001: combat challenge from Bob (4m left)

# 接受挑战（暂停建造，进入战斗）
at-cli inbox do inb-001

# 战斗中
at-cli combat attack
at-cli combat defend

# 战斗结束，自动恢复建造
```

### 手动管理栈

```bash
# 查看栈
at-cli stack
# [scan, look, build house, move north]

# 决定放弃建造，专注移动
at-cli stack drop 2           # 移除 build house
at-cli stack focus 3          # 跳到 move north

# 清理完成
at-cli stack
# [move north]
```

---

## 更新日志

- **v1.0** (当前)：确定 LIFO 任务栈、土地-建筑分离、损耗+维修收费模型
