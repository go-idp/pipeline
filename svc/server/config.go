package server

import "io/fs"

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
	//
	// TaskTimeout 是未显式设置 timeout 的 pipeline 的默认执行超时（秒），
	// 0 表示不限制。默认 3600（见 server 命令 --task-timeout）。
	TaskTimeout int64
	//
	// TaskExecutor 是任务执行方式：in-process（默认，与旧行为一致）或
	// subprocess（独立 pipeline run 子进程，进程级隔离）。
	TaskExecutor string
	//
	// WebFS 是嵌入的前端资源（web/ui/dist）。非 nil 时在根路径挂载 SPA
	//（pipeline web 命令），nil 时保持轻量 API-only（pipeline server 命令）。
	WebFS fs.FS
}
