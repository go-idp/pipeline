package server

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-idp/pipeline"
	"github.com/go-zoox/encoding/yaml"
	"github.com/go-zoox/fs"
)

// PipelineConfigTemplate 配置模板
type PipelineConfigTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	YAML        string                 `json:"yaml"`        // YAML 格式配置
	Visual      map[string]interface{} `json:"visual"`      // 可视化配置（JSON 格式）
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ConfigStore 配置模板存储接口
type ConfigStore interface {
	// Create 创建配置模板
	Create(name, description, yamlConfig string) (*PipelineConfigTemplate, error)
	// Get 获取配置模板
	Get(id string) (*PipelineConfigTemplate, bool)
	// List 列出所有配置模板
	List() []*PipelineConfigTemplate
	// Update 更新配置模板
	Update(id, name, description, yamlConfig string) (*PipelineConfigTemplate, error)
	// Delete 删除配置模板
	Delete(id string) bool
	// ConvertYAMLToVisual 将 YAML 转换为可视化配置
	ConvertYAMLToVisual(yamlConfig string) (map[string]interface{}, error)
	// ConvertVisualToYAML 将可视化配置转换为 YAML
	ConvertVisualToYAML(visual map[string]interface{}) (string, error)
}

type memoryConfigStore struct {
	mu       sync.RWMutex
	configs  map[string]*PipelineConfigTemplate
	workdir string
}

// NewMemoryConfigStore 创建内存配置存储
func NewMemoryConfigStore(workdir string) ConfigStore {
	store := &memoryConfigStore{
		configs:  make(map[string]*PipelineConfigTemplate),
		workdir:  workdir,
	}
	
	// 从文件加载配置
	store.loadFromFiles()
	
	return store
}

func (s *memoryConfigStore) Create(name, description, yamlConfig string) (*PipelineConfigTemplate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证 YAML 格式
	var pl pipeline.Pipeline
	if err := yaml.Decode([]byte(yamlConfig), &pl); err != nil {
		return nil, fmt.Errorf("invalid YAML config: %s", err)
	}

	// 转换为可视化配置
	visual, err := s.convertYAMLToVisual(yamlConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert YAML to visual: %s", err)
	}

	// 生成 ID
	id := fmt.Sprintf("config_%d", time.Now().UnixNano())

	template := &PipelineConfigTemplate{
		ID:          id,
		Name:        name,
		Description: description,
		YAML:        yamlConfig,
		Visual:      visual,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.configs[id] = template
	s.saveToFile(id, template)

	return template, nil
}

func (s *memoryConfigStore) Get(id string) (*PipelineConfigTemplate, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	template, ok := s.configs[id]
	if !ok {
		// 尝试从文件加载
		return s.loadFromFile(id)
	}

	return template, true
}

func (s *memoryConfigStore) List() []*PipelineConfigTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	templates := make([]*PipelineConfigTemplate, 0, len(s.configs))
	for _, template := range s.configs {
		templates = append(templates, template)
	}

	// 按更新时间倒序排序
	for i := 0; i < len(templates)-1; i++ {
		for j := i + 1; j < len(templates); j++ {
			if templates[i].UpdatedAt.Before(templates[j].UpdatedAt) {
				templates[i], templates[j] = templates[j], templates[i]
			}
		}
	}

	return templates
}

func (s *memoryConfigStore) Update(id, name, description, yamlConfig string) (*PipelineConfigTemplate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	template, ok := s.configs[id]
	if !ok {
		return nil, fmt.Errorf("config not found")
	}

	// 验证 YAML 格式
	var pl pipeline.Pipeline
	if err := yaml.Decode([]byte(yamlConfig), &pl); err != nil {
		return nil, fmt.Errorf("invalid YAML config: %s", err)
	}

	// 转换为可视化配置
	visual, err := s.convertYAMLToVisual(yamlConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert YAML to visual: %s", err)
	}

	template.Name = name
	template.Description = description
	template.YAML = yamlConfig
	template.Visual = visual
	template.UpdatedAt = time.Now()

	s.saveToFile(id, template)

	return template, nil
}

func (s *memoryConfigStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.configs[id]; !ok {
		return false
	}

	delete(s.configs, id)
	s.deleteFile(id)

	return true
}

func (s *memoryConfigStore) ConvertYAMLToVisual(yamlConfig string) (map[string]interface{}, error) {
	return s.convertYAMLToVisual(yamlConfig)
}

func (s *memoryConfigStore) ConvertVisualToYAML(visual map[string]interface{}) (string, error) {
	return s.convertVisualToYAML(visual)
}

// convertYAMLToVisual 将 YAML 转换为可视化配置
func (s *memoryConfigStore) convertYAMLToVisual(yamlConfig string) (map[string]interface{}, error) {
	var pl pipeline.Pipeline
	if err := yaml.Decode([]byte(yamlConfig), &pl); err != nil {
		return nil, err
	}

	// 转换为 JSON 格式（可视化配置）
	data, err := json.Marshal(pl)
	if err != nil {
		return nil, err
	}

	var visual map[string]interface{}
	if err := json.Unmarshal(data, &visual); err != nil {
		return nil, err
	}

	return visual, nil
}

// convertVisualToYAML 将可视化配置转换为 YAML
func (s *memoryConfigStore) convertVisualToYAML(visual map[string]interface{}) (string, error) {
	// 先转换为 Pipeline 结构验证
	data, err := json.Marshal(visual)
	if err != nil {
		return "", err
	}

	var pl pipeline.Pipeline
	if err := json.Unmarshal(data, &pl); err != nil {
		return "", fmt.Errorf("invalid visual config: %s", err)
	}

	// 转换为 YAML
	yamlData, err := yaml.Encode(pl)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}

func (s *memoryConfigStore) saveToFile(id string, template *PipelineConfigTemplate) {
	if s.workdir == "" {
		return
	}

	filepath := fmt.Sprintf("%s/.pipeline_configs/%s.json", s.workdir, id)
	dir := fmt.Sprintf("%s/.pipeline_configs", s.workdir)

	if !fs.IsExist(dir) {
		fs.Mkdirp(dir)
	}

	data, err := json.Marshal(template)
	if err != nil {
		return
	}

	fs.WriteFile(filepath, data)
}

func (s *memoryConfigStore) loadFromFile(id string) (*PipelineConfigTemplate, bool) {
	if s.workdir == "" {
		return nil, false
	}

	filepath := fmt.Sprintf("%s/.pipeline_configs/%s.json", s.workdir, id)
	if !fs.IsExist(filepath) {
		return nil, false
	}

	data, err := fs.ReadFile(filepath)
	if err != nil {
		return nil, false
	}

	var template PipelineConfigTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, false
	}

	// 加载到内存
	s.mu.Lock()
	s.configs[id] = &template
	s.mu.Unlock()

	return &template, true
}

func (s *memoryConfigStore) loadFromFiles() {
	if s.workdir == "" {
		return
	}

	dir := fmt.Sprintf("%s/.pipeline_configs", s.workdir)
	if !fs.IsExist(dir) {
		return
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filepath := fmt.Sprintf("%s/%s", dir, file.Name())
		data, err := fs.ReadFile(filepath)
		if err != nil {
			continue
		}

		var template PipelineConfigTemplate
		if err := json.Unmarshal(data, &template); err != nil {
			continue
		}

		s.configs[template.ID] = &template
	}
}

func (s *memoryConfigStore) deleteFile(id string) {
	if s.workdir == "" {
		return
	}

	filepath := fmt.Sprintf("%s/.pipeline_configs/%s.json", s.workdir, id)
	if fs.IsExist(filepath) {
		fs.RemoveFile(filepath)
	}
}
