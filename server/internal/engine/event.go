// Package engine 事件驱动任务执行引擎
// 采用单线程事件循环模型（类似 epoll）
package engine

import (
	"container/heap"
	"sync"
	"time"
)

// EventType 事件类型
type EventType int

const (
	EventTypeTick      EventType = iota // 定时 tick
	EventTypeTaskStart                  // 任务开始
	EventTypeTaskStep                   // 任务步骤执行
	EventTypeTaskComplete               // 任务完成
	EventTypeTaskFail                   // 任务失败
)

// Event 事件接口
type Event interface {
	GetType() EventType
	GetTime() int64 // 毫秒时间戳
	GetAgentID() int64
}

// BaseEvent 基础事件
type BaseEvent struct {
	Type      EventType
	Time      int64 // 执行时间（毫秒）
	AgentID   int64
	TaskSeq   int
}

func (e *BaseEvent) GetType() EventType  { return e.Type }
func (e *BaseEvent) GetTime() int64      { return e.Time }
func (e *BaseEvent) GetAgentID() int64   { return e.AgentID }

// TaskStartEvent 任务开始事件
type TaskStartEvent struct {
	BaseEvent
	TaskType int
	Params   map[string]interface{}
}

// TaskStepEvent 任务步骤事件
type TaskStepEvent struct {
	BaseEvent
	Step      int    // 当前步骤
	TotalStep int    // 总步骤
}

// TaskCompleteEvent 任务完成事件
type TaskCompleteEvent struct {
	BaseEvent
	Result map[string]interface{}
}

// TaskFailEvent 任务失败事件
type TaskFailEvent struct {
	BaseEvent
	ErrorCode int
	Reason    string
}

// TickEvent 定时 Tick 事件
type TickEvent struct {
	BaseEvent
}

// EventHeap 事件优先队列（最小堆）
type EventHeap []Event

func (h EventHeap) Len() int           { return len(h) }
func (h EventHeap) Less(i, j int) bool { return h[i].GetTime() < h[j].GetTime() }
func (h EventHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *EventHeap) Push(x interface{}) {
	*h = append(*h, x.(Event))
}

func (h *EventHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// Peek 查看堆顶元素（不移除）
func (h *EventHeap) Peek() Event {
	if h.Len() == 0 {
		return nil
	}
	return (*h)[0]
}

// EventQueue 事件队列
type EventQueue struct {
	heap      *EventHeap
	mu        sync.Mutex
	notifier  chan struct{} // 通知有新事件
}

// NewEventQueue 创建事件队列
func NewEventQueue() *EventQueue {
	h := &EventHeap{}
	heap.Init(h)
	return &EventQueue{
		heap:     h,
		notifier: make(chan struct{}, 1),
	}
}

// Push 添加事件
func (q *EventQueue) Push(event Event) {
	q.mu.Lock()
	heap.Push(q.heap, event)
	isNewEarliest := q.heap.Len() == 1 || event.GetTime() < q.heap.Peek().GetTime()
	q.mu.Unlock()
	
	// 如果新事件是最早的，通知事件循环
	if isNewEarliest {
		select {
		case q.notifier <- struct{}{}:
		default:
		}
	}
}

// Pop 取出最早的事件
func (q *EventQueue) Pop() Event {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	if q.heap.Len() == 0 {
		return nil
	}
	return heap.Pop(q.heap).(Event)
}

// Peek 查看最早的事件
func (q *EventQueue) Peek() Event {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	if q.heap.Len() == 0 {
		return nil
	}
	return q.heap.Peek()
}

// WaitTime 计算到下一个事件的时间
func (q *EventQueue) WaitTime(now int64) time.Duration {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	if q.heap.Len() == 0 {
		return time.Hour // 没有事件，长时间等待
	}
	
	nextTime := q.heap.Peek().GetTime()
	if nextTime <= now {
		return 0 // 立即执行
	}
	return time.Duration(nextTime-now) * time.Millisecond
}

// GetNotifier 获取通知通道
func (q *EventQueue) GetNotifier() <-chan struct{} {
	return q.notifier
}
