package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var HighInt int
var LowInt int

func init() {
	High := os.Getenv("HIGH")
	if High == "" {
		High = "60000"
	}
	HighInt, _ = strconv.Atoi(High)
	Low := os.Getenv("LOW")
	if Low == "" {
		Low = "20000"
	}
	LowInt, _ = strconv.Atoi(Low)
}
func wstotcp(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	randPort := rand.Intn(HighInt-LowInt) + LowInt
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
		defer wg.Done()
		//新开一个buffer存储数据
		//buffer := make([]byte, 1024)
		for {
			_, message, err := conn.ReadMessage()
			if len(message) == 0 {
				break
			}
			if message[len(message)-1] == 13 {
				message[len(message)-1] = 10
			}

			if err != nil {
				fmt.Println(err)
				return
			}
			tcpConn.Write(message)
		}
	}()

	go func() {
		defer wg.Done()
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
	select {}
}

func main() {
	http.HandleFunc("/wstotcp", wstotcp)
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.ListenAndServe(":8081", nil)
}
