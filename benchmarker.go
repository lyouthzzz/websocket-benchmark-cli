package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type WebsocketBenchmarkerOption func(*WebsocketBenchmarker)

func WebsocketBenchmarkerOptionEndpoint(endpoint string) WebsocketBenchmarkerOption {
	return func(b *WebsocketBenchmarker) { b.endpoint = endpoint }
}

func WebsocketBenchmarkerOptionPath(path string) WebsocketBenchmarkerOption {
	return func(b *WebsocketBenchmarker) { b.path = path }
}

func WebsocketBenchmarkerOptionMessage(message string) WebsocketBenchmarkerOption {
	return func(b *WebsocketBenchmarker) { b.message = message }
}

func WebsocketBenchmarkerOptionMessageInterval(duration time.Duration) WebsocketBenchmarkerOption {
	return func(b *WebsocketBenchmarker) { b.messageInterval = duration }
}

func WebsocketBenchmarkerOptionMessageTimes(times int) WebsocketBenchmarkerOption {
	return func(b *WebsocketBenchmarker) { b.messageTimes = times }
}

func WebsocketBenchmarkerOptionMessageFilePath(messageFilePath string) WebsocketBenchmarkerOption {
	return func(b *WebsocketBenchmarker) { b.messageFilePath = messageFilePath }
}

func WebsocketBenchmarkerUserNum(userNum int) WebsocketBenchmarkerOption {
	return func(b *WebsocketBenchmarker) { b.userNum = userNum }
}

type WebsocketBenchmarker struct {
	endpoint string
	path     string

	userNum         int
	message         string
	messageFilePath string
	messageInterval time.Duration
	messageTimes    int

	dialer *websocket.Dialer
	conns  map[int]*websocket.Conn
	mutex  sync.Mutex

	closed int64
}

func NewWebsocketBenchmarker(opts ...WebsocketBenchmarkerOption) *WebsocketBenchmarker {
	b := &WebsocketBenchmarker{
		endpoint:        "http://localhost:8080",
		path:            "/ws",
		userNum:         500,
		message:         "ping",
		messageInterval: time.Second,
		dialer:          &websocket.Dialer{Proxy: http.ProxyFromEnvironment, HandshakeTimeout: 5 * time.Second},
		conns:           make(map[int]*websocket.Conn),
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

func (b *WebsocketBenchmarker) Test() error {
	if b.messageFilePath != "" {
		_, err := os.Stat(b.messageFilePath)
		if err != nil {
			return err
		}
	}

	uri := url.URL{Scheme: "ws", Host: b.endpoint, Path: b.path}
	c, _, err := b.dialer.Dial(uri.String(), nil)
	if err != nil {
		return err
	}
	_ = c.Close()
	return nil
}

func (b *WebsocketBenchmarker) Start() error {
	if b.messageFilePath != "" {
		content, err := os.ReadFile(b.messageFilePath)
		if err != nil {
			return err
		}
		b.message = string(content)
	}

	messageContent := []byte(b.message)

	log.Printf("websocket message: %s, message interval: %s \n", b.message, b.messageInterval.String())

	wg := &sync.WaitGroup{}
	wg.Add(b.userNum)

	for i := 0; i < b.userNum; i++ {
		connId := i
		go func() {
			defer wg.Done()

			conn, err := b.connect()
			if err != nil {
				log.Println("connect err: ", err)
				return
			}
			b.addConn(connId, conn)

			times := b.messageTimes
			curTimes := 0

			ticker := time.NewTicker(b.messageInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if atomic.LoadInt64(&b.closed) == 1 {
						b.closeConn(connId)
						b.delConn(connId)
						return
					}
					if err := conn.WriteMessage(websocket.TextMessage, messageContent); err != nil {
						log.Printf("conn: %d, write message err: %s", connId, err.Error())
					}
					curTimes++
					if curTimes >= times {
						return
					}
					log.Printf("conn id: %d, cur time: %d times: %d \n", connId, curTimes, times)
				}
			}
		}()
	}

	wg.Wait()

	return nil
}

func (b *WebsocketBenchmarker) connect() (*websocket.Conn, error) {
	uri := url.URL{Scheme: "ws", Host: b.endpoint, Path: b.path}
	c, _, err := b.dialer.Dial(uri.String(), nil)
	return c, err
}

func (b *WebsocketBenchmarker) closeConn(connId int) {
	conn, ok := b.conns[connId]
	if !ok {
		return
	}
	_ = conn.Close()
}

func (b *WebsocketBenchmarker) delConn(connId int) {
	delete(b.conns, connId)
}

func (b *WebsocketBenchmarker) addConn(connId int, conn *websocket.Conn) {
	b.mutex.Lock()
	b.conns[connId] = conn
	b.mutex.Unlock()
}

func (b *WebsocketBenchmarker) Stop() {
	b.closed = 1
	b.mutex.Lock()
	for _, conn := range b.conns {
		_ = conn.Close()
	}
	b.mutex.Unlock()
}
