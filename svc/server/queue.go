package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-idp/pipeline"
	"github.com/go-zoox/logger"
)

// QueueItem 队列项
type QueueItem struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Status    string             `json:"status"` // pending | running | succeeded | failed | cancelled
	CreatedAt time.Time          `json:"created_at"`
	StartedAt *time.Time         `json:"started_at,omitempty"`
	EndedAt   *time.Time         `json:"ended_at,omitempty"`
	Error     string             `json:"error,omitempty"`
	YAML      string             `json:"yaml,omitempty"` // Pipeline YAML 配置
	Pipeline  *pipeline.Pipeline `json:"-"`
	Context   context.Context    `json:"-"`
	Cancel    context.CancelFunc `json:"-"`
}

// Queue 队列接口
type Queue interface {
	// Enqueue 添加 pipeline 到队列
	Enqueue(id, name string, pl *pipeline.Pipeline) error
	// EnqueueWithYAML 添加 pipeline 到队列（带 YAML）
	EnqueueWithYAML(id, name string, pl *pipeline.Pipeline, yaml string) error
	// Dequeue 从队列中取出 pipeline
	Dequeue() (*QueueItem, bool)
	// Get 获取队列项
	Get(id string) (*QueueItem, bool)
	// List 列出所有队列项
	List() []*QueueItem
	// Cancel 取消队列项
	Cancel(id string) bool
	// Stats 获取队列统计信息
	Stats() QueueStats
}

// QueueStats 队列统计信息
type QueueStats struct {
	Total             int `json:"total"`
	Pending           int `json:"pending"`
	Running           int `json:"running"`
	Succeeded         int `json:"succeeded"`
	Failed            int `json:"failed"`
	Cancelled         int `json:"cancelled"`
	MaxConcurrent     int `json:"max_concurrent"`
	CurrentConcurrent int `json:"current_concurrent"`
}

type queue struct {
	mu            sync.RWMutex
	items         map[string]*QueueItem
	pendingItems  []string
	runningItems  map[string]bool
	maxConcurrent int
	store         Store
	workdir       string
	environment   map[string]string
}

// NewQueue 创建队列
func NewQueue(maxConcurrent int, store Store, workdir string, environment map[string]string) Queue {
	q := &queue{
		items:         make(map[string]*QueueItem),
		pendingItems:  make([]string, 0),
		runningItems:  make(map[string]bool),
		maxConcurrent: maxConcurrent,
		store:         store,
		workdir:       workdir,
		environment:   environment,
	}

	// 启动队列处理器
	go q.process()

	return q
}

func (q *queue) Enqueue(id, name string, pl *pipeline.Pipeline) error {
	return q.EnqueueWithYAML(id, name, pl, "")
}

func (q *queue) EnqueueWithYAML(id, name string, pl *pipeline.Pipeline, yaml string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.items[id]; exists {
		return fmt.Errorf("pipeline %s already in queue", id)
	}

	item := &QueueItem{
		ID:        id,
		Name:      name,
		Status:    "pending",
		CreatedAt: time.Now(),
		Pipeline:  pl,
		YAML:      yaml,
	}

	q.items[id] = item
	q.pendingItems = append(q.pendingItems, id)

	// 立即在 store 中创建记录（pending 状态）
	if q.store != nil {
		config := make(map[string]interface{})
		config["name"] = pl.Name
		config["workdir"] = fmt.Sprintf("%s/%s", q.workdir, id)
		config["timeout"] = pl.Timeout
		config["image"] = pl.Image
		q.store.CreateWithYAML(id, name, yaml, config)
	}

	logger.Infof("[queue] enqueued pipeline %s (name: %s)", id, name)

	return nil
}

func (q *queue) Dequeue() (*QueueItem, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 检查是否达到最大并发数
	if len(q.runningItems) >= q.maxConcurrent {
		return nil, false
	}

	// 从待处理队列中取出第一个
	if len(q.pendingItems) == 0 {
		return nil, false
	}

	id := q.pendingItems[0]
	q.pendingItems = q.pendingItems[1:]

	item, exists := q.items[id]
	if !exists {
		return nil, false
	}

	q.runningItems[id] = true
	item.Status = "running"
	now := time.Now()
	item.StartedAt = &now

	return item, true
}

func (q *queue) Get(id string) (*QueueItem, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	item, exists := q.items[id]
	return item, exists
}

func (q *queue) List() []*QueueItem {
	q.mu.RLock()
	defer q.mu.RUnlock()

	items := make([]*QueueItem, 0, len(q.items))
	for _, item := range q.items {
		items = append(items, item)
	}

	return items
}

func (q *queue) Cancel(id string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, exists := q.items[id]
	if !exists {
		return false
	}

	// 如果正在运行，取消执行
	if item.Status == "running" && item.Cancel != nil {
		item.Cancel()
		delete(q.runningItems, id)
		item.Status = "cancelled"
		item.Error = "cancelled by user"
		now := time.Now()
		item.EndedAt = &now

		// 更新 store 状态
		if q.store != nil {
			q.store.UpdateStatus(id, "cancelled", fmt.Errorf("cancelled by user"))
		}

		logger.Infof("[queue] pipeline %s cancelled", id)
		return true
	}

	// 如果是 pending 状态，标记为 cancelled
	if item.Status == "pending" {
		// 从待处理队列中移除
		for i, pendingID := range q.pendingItems {
			if pendingID == id {
				q.pendingItems = append(q.pendingItems[:i], q.pendingItems[i+1:]...)
				break
			}
		}

		item.Status = "cancelled"
		item.Error = "cancelled by user"
		now := time.Now()
		item.EndedAt = &now

		// 更新 store 状态
		if q.store != nil {
			q.store.UpdateStatus(id, "cancelled", fmt.Errorf("cancelled by user"))
		}

		logger.Infof("[queue] pipeline %s cancelled (pending)", id)
		return true
	}

	// 其他状态不能取消
	return false
}

func (q *queue) Stats() QueueStats {
	q.mu.RLock()
	defer q.mu.RUnlock()

	stats := QueueStats{
		Total:             len(q.items),
		MaxConcurrent:     q.maxConcurrent,
		CurrentConcurrent: len(q.runningItems),
	}

	for _, item := range q.items {
		switch item.Status {
		case "pending":
			stats.Pending++
		case "running":
			stats.Running++
		case "succeeded":
			stats.Succeeded++
		case "failed":
			stats.Failed++
		case "cancelled":
			stats.Cancelled++
		}
	}

	return stats
}

// process 处理队列
func (q *queue) process() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		item, ok := q.Dequeue()
		if !ok {
			continue
		}

		// 在 goroutine 中执行 pipeline
		go q.execute(item)
	}
}

// execute 执行 pipeline
func (q *queue) execute(item *QueueItem) {
	logger.Infof("[queue] executing pipeline %s (name: %s)", item.ID, item.Name)

	// 创建 context
	ctx, cancel := context.WithCancel(context.Background())
	item.Context = ctx
	item.Cancel = cancel

	// 创建 pipeline 记录（如果还没有创建）
	config := make(map[string]interface{})
	config["name"] = item.Pipeline.Name
	config["workdir"] = fmt.Sprintf("%s/%s", q.workdir, item.ID)
	config["timeout"] = item.Pipeline.Timeout
	config["image"] = item.Pipeline.Image

	if q.store != nil {
		// 检查记录是否已存在
		if _, exists := q.store.Get(item.ID); !exists {
			q.store.CreateWithYAML(item.ID, item.Pipeline.Name, item.YAML, config)
		}
		q.store.UpdateStatus(item.ID, "running", nil)
	}

	// 设置 pipeline
	item.Pipeline.SetWorkdir(fmt.Sprintf("%s/%s", q.workdir, item.ID))
	item.Pipeline.SetEnvironment(q.environment)

	// 设置输出，将日志记录到 store
	if q.store != nil {
		item.Pipeline.SetStdout(&queueWriter{
			store: q.store,
			id:    item.ID,
			typ:   "stdout",
		})
		item.Pipeline.SetStderr(&queueWriter{
			store: q.store,
			id:    item.ID,
			typ:   "stderr",
		})
	}

	// 执行 pipeline
	err := item.Pipeline.Run(ctx, func(cfg *pipeline.RunConfig) {
		cfg.ID = item.ID
	})

	// 更新状态
	q.mu.Lock()
	defer q.mu.Unlock()

	// 检查是否已经被取消（在锁外可能被取消）
	if item.Status == "cancelled" {
		return
	}

	delete(q.runningItems, item.ID)
	now := time.Now()
	item.EndedAt = &now

	// 检查是否是 context 取消错误
	if err != nil {
		if err == context.Canceled {
			item.Status = "cancelled"
			item.Error = "cancelled by user"
			if q.store != nil {
				q.store.UpdateStatus(item.ID, "cancelled", err)
			}
			logger.Infof("[queue] pipeline %s cancelled", item.ID)
		} else {
			item.Status = "failed"
			item.Error = err.Error()
			if q.store != nil {
				q.store.UpdateStatus(item.ID, "failed", err)
			}
			logger.Errorf("[queue] pipeline %s failed: %s", item.ID, err)
		}
	} else {
		item.Status = "succeeded"
		if q.store != nil {
			q.store.UpdateStatus(item.ID, "succeeded", nil)
		}
		logger.Infof("[queue] pipeline %s succeeded", item.ID)
	}
}

// queueWriter 队列写入器，用于将日志写入 store
type queueWriter struct {
	store Store
	id    string
	typ   string
}

func (w *queueWriter) Write(p []byte) (n int, err error) {
	if w.store != nil {
		w.store.AddLog(w.id, w.typ, string(p))
	}
	return len(p), nil
}
