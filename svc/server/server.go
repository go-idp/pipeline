package server

type Server interface {
	Run() error
}

type server struct {
	cfg         *Config
	store       Store
	queue       Queue
	configStore ConfigStore
}

func New(cfg *Config) Server {
	maxConcurrent := cfg.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 2 // 默认并发数为 2
	}

	store := NewMemoryStore(cfg.Workdir, 1000) // 最多保存1000条记录

	executor := cfg.TaskExecutor
	if executor == "" {
		executor = TaskExecutorInProcess
	}

	queue := NewQueue(maxConcurrent, store, cfg.Workdir, cfg.Environment, executor, cfg.TaskTimeout)
	configStore := NewMemoryConfigStore(cfg.Workdir)

	// 启动时注入模拟数据（内置模板 + 演示运行，均幂等：仅当对应存储为空）
	seed(configStore, store)

	return &server{
		cfg:         cfg,
		store:       store,
		queue:       queue,
		configStore: configStore,
	}
}
