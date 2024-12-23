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

			go func() {
				// // prepare
				// pl.SetOnChange(func(typ string, status string, payload any) {
				// 	type Status struct {
				// 		Type    string `json:"type"`
				// 		Status  string `json:"status"`
				// 		Payload any    `json:"payload"`
				// 	}

				// 	conn.WriteJSON(&Status{
				// 		Type:    "status",
				// 		Status:  status,
				// 		Payload: payload,
				// 	})
				// })

				// pl.SetOnLog(func(message, typ string, payload any) {
				// 	type Log struct {
				// 		Type    string `json:"type"`
				// 		Message string `json:"message"`
				// 		Context any    `json:"context"`
				// 	}

				// 	conn.WriteJSON(&Log{
				// 		Type:    "log",
				// 		Message: message,
				// 		Context: map[string]any{
				// 			"type":    typ,
				// 			"payload": payload,
				// 		},
				// 	})
				// })

				// pl.Workdir = fmt.Sprintf("%s/%s", s.cfg.Workdir, conn.ID())
				pl.SetWorkdir(fmt.Sprintf("%s/%s", cfg.Workdir, conn.ID()))
				//
				pl.SetEnvironment(cfg.Environment)
				//
				pl.SetStdout(stdout)
				pl.SetStderr(stderr)

				// started
				err := pl.Run(conn.Context(), func(cfg *pipeline.RunConfig) {
					cfg.ID = conn.ID()
				})
				if err != nil {
					sendError(fmt.Errorf("failed to run pipeline: %s", err))
					return
				}

				sendDone()

				// succeeded
			}()
		default:
			sendError(fmt.Errorf("unsupported action type: %s", act.Type))
			return nil
		}

		return nil
	})

	return nil
}
