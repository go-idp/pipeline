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
}
