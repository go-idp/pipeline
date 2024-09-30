package server

type Server interface {
	Run() error
}

type server struct {
	cfg *Config
}

func New(cfg *Config) Server {
	return &server{
		cfg: cfg,
	}
}
