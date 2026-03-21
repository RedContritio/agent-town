# 事件驱动任务执行引擎设计

## 架构概述

采用**单线程事件循环模型**（类似 epoll），所有任务调度在一个 goroutine 中处理，避免并发复杂性。

```
┌─────────────────────────────────────────────────────────────┐
│                      Event Loop (单线程)                      │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │  Event Queue │───▶│ Event Handler│───▶│ Task Executor│  │
│  │   (Min Heap) │    │              │    │              │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
│         ▲                                    │              │
│         │                                    ▼              │
│         │                           ┌──────────────┐       │
│         │                           │ Push Events  │───────┤
│         │                           │ (Next Steps) │       │
│         │                           └──────────────┘       │
│         │                                                   │
│    Tick Generator (1s)                                      │
└─────────────────────────────────────────────────────────────┘
```

## 核心组件

### 1. Event Queue（事件队列）
- **优先队列（最小堆）**：按时间排序
- **线程安全**：Push 操作加锁，Pop 在主循环单线程执行
- **通知机制**：新事件到达时通知事件循环

```go
type EventQueue struct {
    heap      *EventHeap      // 最小堆
    mu        sync.Mutex      // 保护堆操作
    notifier  chan struct{}   // 通知通道
}
```

### 2. Event Loop（事件循环）
- **单 goroutine**：无锁处理事件
- **时间等待**：计算到下一个事件的时间，避免忙等
- **事件处理**：顺序处理所有到期事件

```go
func (e *Engine) eventLoop() {
    for {
        now := model.NowMillis()
        waitTime := e.eventQueue.WaitTime(now)
        
        select {
        case <-time.After(waitTime):
            // 处理到期事件
        case <-e.eventQueue.GetNotifier():
            // 新事件到达，立即检查
        }
        
        // 处理所有到期事件
        for e.eventQueue.Peek().GetTime() <= now {
            event := e.eventQueue.Pop()
            e.handleEvent(event)
        }
    }
}
```

### 3. Task Executor（任务执行器）
- **无副作用**：执行后返回新事件列表
- **纯函数**：根据任务类型计算结果

```go
func (e *TaskExecutor) ExecuteTaskStart(agentID, taskSeq, taskType int, params) ([]Event, error) {
    switch taskType {
    case TaskTypeMove:
        return e.executeMoveStart(...)
    case TaskTypeHarvest:
        return e.executeHarvestStart(...)
    // ...
    }
}
```

## 事件类型

| 事件 | 类型 | 触发条件 |
|------|------|----------|
| `EventTypeTick` | 系统 | 每秒定时触发 |
| `EventTypeTaskStart` | 任务 | 任务被调度执行 |
| `EventTypeTaskStep` | 任务 | 多步骤任务的中间步骤 |
| `EventTypeTaskComplete` | 任务 | 任务完成 |
| `EventTypeTaskFail` | 任务 | 任务失败 |

## 时间计算（简化版）

| 任务类型 | 耗时计算 |
|----------|----------|
| **移动** | 每格 2 秒，`总时间 = 距离 × 2s` |
| **采集** | 固定 30 秒 |
| **制作** | 固定 10 秒 |
| **建造** | 固定 60 秒 |

## 并发控制

### 单线程保证
- 所有事件处理在主循环单 goroutine
- 无共享状态竞争
- 无需复杂的锁机制

### 线程安全边界
```
┌──────────────┐         ┌──────────────┐
│   HTTP API   │◄───────►│   Database   │
│   (多线程)    │         │  (SQLite/WAL)│
└──────────────┘         └──────────────┘
       │                         ▲
       │    Schedule Task        │
       └─────────────────────────┤
                                 │
┌──────────────┐         ┌──────┴───────┐
│   Engine     │         │   Engine     │
│   (单线程)    │◄───────►│   (单线程)   │
│  事件循环     │         │  定时器生成   │
└──────────────┘         └──────────────┘
```

## 使用示例

### 创建任务并调度
```go
// HTTP Handler 中
taskService.CreateTask(agentID, req)

// TaskService 中自动调度
engine.GetManager().ScheduleTask(agentID, seq)

// Engine 中处理
handleTaskStart() → ExecuteTaskStart() → Push Events
```

### 任务状态流转
```
CREATED → EventTypeTaskStart → RUNNING
  ↓
EventTypeTaskStep (多次)
  ↓
EventTypeTaskComplete → COMPLETED
  ↓
恢复下一个任务
```

## 性能特点

1. **低延迟**：事件到达立即处理（无轮询）
2. **高吞吐**：批量处理到期事件
3. **可扩展**：单线程模型易于水平扩展（多进程）
4. **简单可靠**：无并发 bug 风险

## 后续优化方向

1. **持久化事件**：崩溃后恢复未处理事件
2. **分布式调度**：多实例间协调（Redis/RabbitMQ）
3. **优先级队列**：VIP 任务优先处理
4. **批量执行**：相似任务合并处理
