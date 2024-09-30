package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-idp/pipeline/svc/action"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/websocket"
	"github.com/go-zoox/websocket/conn"
)

func (c *client) Connect() error {
	u, err := url.Parse(c.cfg.Server)
	if err != nil {
		return fmt.Errorf("invalid caas server address: %s", err)
	}
	logger.Debugf("connecting to %s", u.String())

	if u.User != nil {
		c.cfg.Username = u.User.Username()
		c.cfg.Password, _ = u.User.Password()

		// @TODO fix malformed ws or wss URL
		u.User = nil
	}

	headers := http.Header{}
	if c.cfg.Username != "" || c.cfg.Password != "" {
		headers.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(c.cfg.Username+":"+c.cfg.Password))))
	}

	wc, err := websocket.NewClient(func(opt *websocket.ClientOption) {
		opt.Context = context.Background()
		opt.Addr = u.String()
		opt.Headers = headers
		opt.ConnectTimeout = 10 * time.Second
	})
	if err != nil {
		return err
	}

	c.core = wc

	connectCh := make(chan struct{})

	wc.OnConnect(func(conn conn.Conn) error {
		connectCh <- struct{}{}
		return nil
	})

	wc.OnClose(func(conn conn.Conn, code int, message string) error {
		c.done <- fmt.Errorf("terminal connection closed (code: %d)", code)
		return nil
	})

	wc.OnTextMessage(func(conn websocket.Conn, msg []byte) error {
		var act action.Action
		if err := json.Unmarshal(msg, &act); err != nil {
			c.done <- fmt.Errorf("failed to unmarshal message: %s", err)
			return nil
		}

		switch act.Type {
		case action.Error.Name():
			err, errx := action.Error.Decode([]byte(act.Payload))
			if errx != nil {
				c.done <- fmt.Errorf("failed to decode error message: %s", errx)
				return nil
			}

			c.done <- err
		case action.Done.Name():
			done, err := action.Done.Decode([]byte(act.Payload))
			if err != nil {
				c.done <- fmt.Errorf("failed to decode done message: %s", err)
				return nil
			}

			c.done <- done
		case action.Log.Name():
			log, err := action.Log.Decode([]byte(act.Payload))
			if err != nil {
				c.done <- fmt.Errorf("failed to decode log message: %s", err)
				return nil
			}

			os.Stdout.Write(log)
		default:
			os.Stderr.Write([]byte(fmt.Sprintf("unknown message type: %v\n", act.Type)))
		}

		return nil
	})

	if err := wc.Connect(); err != nil {
		return err
	}

	// wait for connect
	<-connectCh

	logger.Debugf("connected to %s", u.String())

	return nil
}
