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
	queue := NewQueue(maxConcurrent, store, cfg.Workdir, cfg.Environment)
	configStore := NewMemoryConfigStore(cfg.Workdir)

	return &server{
		cfg:         cfg,
		store:       store,
		queue:       queue,
		configStore: configStore,
	}
}
