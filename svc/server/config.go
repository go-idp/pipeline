package server

type Config struct {
	Port int
	//
	Path string
	//
	Workdir string
	//
	Environment map[string]string
	//
	Username string
	Password string
	//
	MaxConcurrent int // 最大并发数，默认 2
}
