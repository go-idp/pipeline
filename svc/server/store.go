package server

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-zoox/fs"
)

// PipelineRecord 记录 pipeline 执行信息
type PipelineRecord struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Status      string                 `json:"status"` // pending | running | succeeded | failed | cancelled
	StartedAt   time.Time              `json:"started_at"`
	SucceedAt   *time.Time             `json:"succeed_at,omitempty"`
	FailedAt    *time.Time             `json:"failed_at,omitempty"`
	CancelledAt *time.Time             `json:"cancelled_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	YAML        string                 `json:"yaml,omitempty"` // 完整的 pipeline YAML 配置
	Logs        []LogEntry             `json:"logs,omitempty"`
}

// LogEntry 日志条目
type LogEntry struct {
	Type      string    `json:"type"` // stdout | stderr | status
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Store pipeline 执行历史存储
type Store interface {
	// Create 创建新的 pipeline 记录
	Create(id, name string, config map[string]interface{}) *PipelineRecord
	// CreateWithYAML 创建新的 pipeline 记录（带 YAML）
	CreateWithYAML(id, name, yaml string, config map[string]interface{}) *PipelineRecord
	// Get 获取 pipeline 记录
	Get(id string) (*PipelineRecord, bool)
	// List 列出所有 pipeline 记录
	List(limit int) []*PipelineRecord
	// UpdateStatus 更新 pipeline 状态
	UpdateStatus(id, status string, err error)
	// AddLog 添加日志
	AddLog(id string, logType, message string)
	// Delete 删除 pipeline 记录
	Delete(id string) bool
}

type memoryStore struct {
	mu      sync.RWMutex
	records map[string]*PipelineRecord
	maxSize int
	workdir string
}

// NewMemoryStore 创建内存存储
func NewMemoryStore(workdir string, maxSize int) Store {
	return &memoryStore{
		records: make(map[string]*PipelineRecord),
		maxSize: maxSize,
		workdir: workdir,
	}
}

func (s *memoryStore) Create(id, name string, config map[string]interface{}) *PipelineRecord {
	return s.CreateWithYAML(id, name, "", config)
}

func (s *memoryStore) CreateWithYAML(id, name, yaml string, config map[string]interface{}) *PipelineRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果超过最大数量，删除最旧的记录
	if len(s.records) >= s.maxSize {
		s.evictOldest()
	}

	record := &PipelineRecord{
		ID:        id,
		Name:      name,
		Status:    "pending",
		StartedAt: time.Now(),
		Config:    config,
		YAML:      yaml,
		Logs:      make([]LogEntry, 0),
	}

	s.records[id] = record

	// 保存到文件（可选）
	s.saveToFile(id, record)

	return record
}

func (s *memoryStore) Get(id string) (*PipelineRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.records[id]
	if !ok {
		// 尝试从文件加载
		return s.loadFromFile(id)
	}

	return record, true
}

func (s *memoryStore) List(limit int) []*PipelineRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := make([]*PipelineRecord, 0, len(s.records))
	for _, record := range s.records {
		records = append(records, record)
	}

	// 按时间倒序排序
	for i := 0; i < len(records)-1; i++ {
		for j := i + 1; j < len(records); j++ {
			if records[i].StartedAt.Before(records[j].StartedAt) {
				records[i], records[j] = records[j], records[i]
			}
		}
	}

	if limit > 0 && limit < len(records) {
		return records[:limit]
	}

	return records
}

func (s *memoryStore) UpdateStatus(id, status string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[id]
	if !ok {
		return
	}

	record.Status = status
	now := time.Now()

	switch status {
	case "succeeded":
		record.SucceedAt = &now
	case "failed":
		record.FailedAt = &now
		if err != nil {
			record.Error = err.Error()
		}
	case "cancelled":
		record.CancelledAt = &now
		if err != nil {
			record.Error = err.Error()
		}
	}

	s.saveToFile(id, record)
}

func (s *memoryStore) AddLog(id string, logType, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[id]
	if !ok {
		return
	}

	entry := LogEntry{
		Type:      logType,
		Message:   message,
		Timestamp: time.Now(),
	}

	record.Logs = append(record.Logs, entry)

	// 限制日志数量，避免内存溢出
	if len(record.Logs) > 10000 {
		record.Logs = record.Logs[len(record.Logs)-10000:]
	}

	s.saveToFile(id, record)
}

func (s *memoryStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.records[id]; !ok {
		return false
	}

	delete(s.records, id)

	// 删除文件
	s.deleteFile(id)

	return true
}

func (s *memoryStore) evictOldest() {
	var oldestID string
	var oldestTime time.Time
	first := true

	for id, record := range s.records {
		if first || record.StartedAt.Before(oldestTime) {
			oldestID = id
			oldestTime = record.StartedAt
			first = false
		}
	}

	if oldestID != "" {
		delete(s.records, oldestID)
		s.deleteFile(oldestID)
	}
}

func (s *memoryStore) saveToFile(id string, record *PipelineRecord) {
	if s.workdir == "" {
		return
	}

	filepath := fmt.Sprintf("%s/.pipeline_records/%s.json", s.workdir, id)
	dir := fmt.Sprintf("%s/.pipeline_records", s.workdir)

	if !fs.IsExist(dir) {
		fs.Mkdirp(dir)
	}

	data, err := json.Marshal(record)
	if err != nil {
		return
	}

	fs.WriteFile(filepath, data)
}

func (s *memoryStore) loadFromFile(id string) (*PipelineRecord, bool) {
	if s.workdir == "" {
		return nil, false
	}

	filepath := fmt.Sprintf("%s/.pipeline_records/%s.json", s.workdir, id)
	if !fs.IsExist(filepath) {
		return nil, false
	}

	data, err := fs.ReadFile(filepath)
	if err != nil {
		return nil, false
	}

	var record PipelineRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, false
	}

	// 加载到内存
	s.mu.Lock()
	s.records[id] = &record
	s.mu.Unlock()

	return &record, true
}

func (s *memoryStore) deleteFile(id string) {
	if s.workdir == "" {
		return
	}

	filepath := fmt.Sprintf("%s/.pipeline_records/%s.json", s.workdir, id)
	if fs.IsExist(filepath) {
		fs.RemoveFile(filepath)
	}
}
