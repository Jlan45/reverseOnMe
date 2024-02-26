package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func init() {
	os.Getenv("HOST")
}
func wstotcp(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	randPort := rand.Intn(34000) + 20000
	// 假设你想连接的TCP服务器在 localhost:8080 上
	tcpList, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", randPort))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("已在%d端口监听\n", randPort)
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("已在%d端口监听\n", randPort)))
	defer tcpList.Close()
	tcpConn, err := tcpList.Accept()
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("已成功建立连接\n")))
	defer func() {
		conn.WriteMessage(websocket.TextMessage, []byte("连接已关闭"))
	}()
	// 启动两个goroutine，分别用于从WebSocket读取数据并写入TCP连接，以及从TCP连接读取数据并写入WebSocket
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			tcpConn.Write(message)
		}
	}()

	go func() {
		for {
			buffer := make([]byte, 1024)
			n, err := tcpConn.Read(buffer)
			if err != nil {
				fmt.Println(err)
				return
			}
			conn.WriteMessage(websocket.TextMessage, buffer[:n])
		}
	}()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigCh:
		// 收到信号，关闭连接
		fmt.Println("Received signal. Closing connection.")
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "已退出"))
	case <-time.After(time.Second * 10):
		// 10秒超时，关闭连接
		fmt.Println("Timeout. Closing connection.")
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "十秒超时"))
	}
}

func main() {
	http.HandleFunc("/wstotcp", wstotcp)
	http.ListenAndServe(":8081", nil)
}
