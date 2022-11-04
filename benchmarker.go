package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
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

func WebsocketBenchmarkerOptionConnectInterval(duration time.Duration) WebsocketBenchmarkerOption {
	return func(b *WebsocketBenchmarker) { b.connectInterval = duration }
}

func WebsocketBenchmarkerConnectionSleep(duration time.Duration) WebsocketBenchmarkerOption {
	return func(b *WebsocketBenchmarker) { b.connectionSleep = duration }
}

func WebsocketBenchmarkerOptionMessageInterval(duration time.Duration) WebsocketBenchmarkerOption {
	return func(b *WebsocketBenchmarker) { b.messageInterval = duration }
}

func WebsocketBenchmarkerOptionMessageTimes(times int) WebsocketBenchmarkerOption {
	return func(b *WebsocketBenchmarker) { b.messageTimes = times }
}

func WebsocketBenchmarkerOptionUserNum(userNum int) WebsocketBenchmarkerOption {
	return func(b *WebsocketBenchmarker) { b.userNum = userNum }
}

type WebsocketBenchmarker struct {
	endpoint string
	path     string

	userNum         int
	message         string
	messageInterval time.Duration
	connectInterval time.Duration
	connectionSleep time.Duration
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
		messageInterval: 30 * time.Second,
		connectInterval: 10 * time.Millisecond,
		dialer:          &websocket.Dialer{Proxy: http.ProxyFromEnvironment, HandshakeTimeout: 60 * time.Second},
		conns:           make(map[int]*websocket.Conn),
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

func (b *WebsocketBenchmarker) Test() error {
	uri := url.URL{Scheme: "ws", Host: b.endpoint, Path: b.path}
	c, _, err := b.dialer.Dial(uri.String(), nil)
	if err != nil {
		return err
	}
	_ = c.Close()
	return nil
}

func (b *WebsocketBenchmarker) StartConnBenchmark() error {
	return b.initConnection()
}

func (b *WebsocketBenchmarker) initConnection() error {
	signalChan := make(chan struct{}, b.userNum)

	go func() {
		ticker := time.NewTicker(b.connectInterval)
		defer ticker.Stop()

		for i := 0; i < b.userNum; i++ {
			select {
			case <-ticker.C:
				signalChan <- struct{}{}
			}
		}
	}()

	for i := 0; i < b.userNum; i++ {
		if atomic.LoadInt64(&b.closed) == 1 {
			break
		}

		connId := i + 1
		<-signalChan
		conn, err := b.connect()
		if err != nil {
			log.Printf("conn: %d. connect err: %s\n", connId, err.Error())
			continue
		}
		b.addConn(connId, conn)
		log.Println("add a conn ", connId)
	}
	close(signalChan)
	return nil
}

func (b *WebsocketBenchmarker) StartMessageBenchmark() error {
	if err := b.initConnection(); err != nil {
		return err
	}
	messageContent := []byte(b.message)

	log.Println(b.MessageInfo())

	wg := &sync.WaitGroup{}
	wg.Add(len(b.conns))

	for connId, conn := range b.conns {
		conn := conn
		connId := connId

		go func() {
			defer wg.Done()
			times := b.messageTimes
			curTimes := 0

			ticker := time.NewTicker(b.messageInterval)
			defer ticker.Stop()

			for {
				if atomic.LoadInt64(&b.closed) == 1 {
					return
				}

				select {
				case <-ticker.C:
					if err := conn.WriteMessage(websocket.TextMessage, messageContent); err != nil {
						log.Printf("conn: %d, write message err: %s \n", connId, err.Error())
					}
					curTimes++
					if curTimes >= times {
						return
					}
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
	defer b.mutex.Unlock()
	if atomic.LoadInt64(&b.closed) == 1 {
		return
	}
	b.conns[connId] = conn
}

func (b *WebsocketBenchmarker) Stop() {
	if !atomic.CompareAndSwapInt64(&b.closed, 0, 1) {
		return
	}
	b.mutex.Lock()
	for _, conn := range b.conns {
		if err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closed")); err != nil {
			//log.Println("conn close err: " + err.Error())
		}
	}
	b.mutex.Unlock()
}

func (b *WebsocketBenchmarker) ConnInfo() string {
	return fmt.Sprintf("\nendpoint: %s \npath: %s \nuser: %d \nconnectIntv: %s \nconnectionSleep: %s\n",
		b.endpoint, b.path, b.userNum, b.connectInterval.String(), b.connectionSleep.String())
}

func (b *WebsocketBenchmarker) MessageInfo() string {
	return fmt.Sprintf("\nendpoint: %s \npath: %s \nuser: %d \nconnectIntv: %s \nmessageContent: %s \nmessageInterval: %s \nmessageTimes: %d\n",
		b.endpoint, b.path, b.userNum, b.connectInterval.String(), b.message, b.messageInterval, b.messageTimes)
}
