package server

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-idp/pipeline"
	"github.com/go-idp/pipeline/svc/action"
	"github.com/go-zoox/core-utils/io"
	"github.com/go-zoox/debug"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/websocket/conn"
	"github.com/go-zoox/zoox"
)

type MountConfig struct {
	Path string
	//
	Workdir string
	//
	Environment map[string]string
	//
	Store Store
	//
	Queue Queue
}

type MountOption func(cfg *MountConfig)

func Mount(app *zoox.Application, opts ...MountOption) error {
	cfg := MountConfig{
		Path: "/",
	}
	for _, o := range opts {
		o(&cfg)
	}

	server, err := app.WebSocket(cfg.Path)
	if err != nil {
		return err
	}

	server.OnConnect(func(conn conn.Conn) error {
		logger.Infof("[ws][connected] client %s", conn.ID())
		return nil
	})

	server.OnClose(func(conn conn.Conn, code int, message string) error {
		logger.Infof("[ws][closed] client %s", conn.ID())
		return nil
	})

	server.OnTextMessage(func(conn conn.Conn, msg []byte) error {
		sendError := func(err error) {
			logger.Errorf("error: %s", err)

			msg, errx := action.Error.Encode(err)
			if errx != nil {
				panic(fmt.Errorf("failed to encode error: %s", errx))
			}

			conn.WriteTextMessage(msg)
		}

		sendDone := func() {
			msg, err := action.Done.Encode(nil)
			if err != nil {
				panic(fmt.Errorf("failed to encode error: %s", err))
			}

			conn.WriteTextMessage(msg)
		}

		stdout := io.WriterWrapFunc(func(b []byte) (n int, err error) {
			if debug.IsDebugMode() {
				os.Stdout.Write(b)
			}

			msg, err := action.Stdout.Encode(b)
			if err != nil {
				panic(fmt.Errorf("failed to encode error: %s", err))
			}

			conn.WriteTextMessage(msg)

			// 记录日志到存储
			if cfg.Store != nil {
				cfg.Store.AddLog(conn.ID(), "stdout", string(b))
			}

			return len(b), nil
		})

		stderr := io.WriterWrapFunc(func(b []byte) (n int, err error) {
			if debug.IsDebugMode() {
				os.Stderr.Write(b)
			}

			msg, err := action.Stderr.Encode(b)
			if err != nil {
				panic(fmt.Errorf("failed to encode error: %s", err))
			}

			conn.WriteTextMessage(msg)

			// 记录日志到存储
			if cfg.Store != nil {
				cfg.Store.AddLog(conn.ID(), "stderr", string(b))
			}

			return len(b), nil
		})

		var act action.Action
		if err := json.Unmarshal(msg, &act); err != nil {
			sendError(err)
			return nil
		}

		switch act.Type {
		case action.Run.Name():
			pl, err := action.Run.Decode([]byte(act.Payload))
			if err != nil {
				sendError(err)
				return nil
			}

			// 保存原始 YAML
			yamlPayload := act.Payload

			// 设置输出
			pl.SetStdout(stdout)
			pl.SetStderr(stderr)

			// 添加到队列
			if cfg.Queue != nil {
				// 传递 YAML 到队列
				if err := cfg.Queue.EnqueueWithYAML(conn.ID(), pl.Name, pl, yamlPayload); err != nil {
					sendError(fmt.Errorf("failed to enqueue pipeline: %s", err))
					return nil
				}

				// 发送确认消息
				sendDone()
			} else {
				// 如果没有队列，直接执行（向后兼容）
				go func() {
					// 创建 pipeline 记录
					config := make(map[string]interface{})
					config["name"] = pl.Name
					config["workdir"] = fmt.Sprintf("%s/%s", cfg.Workdir, conn.ID())
					config["timeout"] = pl.Timeout
					config["image"] = pl.Image

					if cfg.Store != nil {
						cfg.Store.CreateWithYAML(conn.ID(), pl.Name, yamlPayload, config)
						cfg.Store.UpdateStatus(conn.ID(), "running", nil)
					}

					pl.SetWorkdir(fmt.Sprintf("%s/%s", cfg.Workdir, conn.ID()))
					pl.SetEnvironment(cfg.Environment)

					err := pl.Run(conn.Context(), func(cfg *pipeline.RunConfig) {
						cfg.ID = conn.ID()
					})

					if cfg.Store != nil {
						if err != nil {
							cfg.Store.UpdateStatus(conn.ID(), "failed", err)
						} else {
							cfg.Store.UpdateStatus(conn.ID(), "succeeded", nil)
						}
					}

					if err != nil {
						sendError(fmt.Errorf("failed to run pipeline: %s", err))
						return
					}

					sendDone()
				}()
			}
		default:
			sendError(fmt.Errorf("unsupported action type: %s", act.Type))
			return nil
		}

		return nil
	})

	return nil
}
