package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
)

type Connection struct {
	ID            string
	Port          int
	History       string
	TCPconnection net.Conn
	TCPlistener   net.Listener
	WSConnection  map[string]websocket.Conn
	//Channel chan int
}

func (c *Connection) createTCPListener() error {
	tcpList, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", c.Port))
	fmt.Printf("已在%d端口监听\n", c.Port)
	if err != nil {
		fmt.Println(err)
		return err
	}
	c.TCPlistener = tcpList
	return nil
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var connectionList = make(map[string]*Connection)
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
func wstotcp(c *gin.Context) {
	id := c.Param("id")
	connection := connectionList[id]
	if connection == nil {
		c.String(404, "连接不存在")
		return
	}
	wsID := getRandID()
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	conn.SetCloseHandler(func(code int, text string) error {
		conn.Close()
		delete(connection.WSConnection, wsID)
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	connection.WSConnection[wsID] = *conn
	if connection.TCPconnection == nil {
		conn.WriteMessage(websocket.TextMessage, []byte("端口"+strconv.Itoa(connection.Port)+"监听中...\n"))
	}
	if connection.History != "" {
		conn.WriteMessage(websocket.TextMessage, []byte("历史消息\n"+connection.History))
	}
	go func() {
		for {
			_, buffer, err := conn.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			connection.History += string(buffer)
			connection.TCPconnection.Write(buffer)
			for wsid, wsConn := range connection.WSConnection {
				if wsid != wsID {
					wsConn.WriteMessage(websocket.TextMessage, []byte("来自"+wsID+"的消息\n"+string(buffer)+"\n"))
				}
			}
		}
	}()
	select {}
}
func getRandID() string {
	randChars := "abcdefghijklmnopqrstuvwxyz1234567890"
	id := ""
	for i := 0; i < 8; i++ {
		id += string(randChars[rand.Intn(len(randChars))])
	}
	return id
}
func createNewConnection(c *gin.Context) {
	newConnection := Connection{
		ID:            getRandID(),
		Port:          rand.Intn(HighInt-LowInt) + LowInt,
		TCPconnection: nil,
		WSConnection:  make(map[string]websocket.Conn),
		History:       "",
	}
	if newConnection.createTCPListener() != nil {
		//等会在写
		c.String(200, "创建监听失败（刷新）")
	}
	connectionList[newConnection.ID] = &newConnection
	c.JSON(200, gin.H{"ID": newConnection.ID, "port": newConnection.Port})
	go func() {
		for {
			conn, err := newConnection.TCPlistener.Accept()
			if err == nil {
				newConnection.TCPconnection = conn
			}
			go func() {
				buffer := make([]byte, 2048)
				for {
					len, err := conn.Read(buffer)
					if err != nil {
						fmt.Println(err)
						return
					}
					newConnection.History += string(buffer[:len])
					for _, wsConn := range newConnection.WSConnection {
						wsConn.WriteMessage(websocket.TextMessage, (buffer[:len]))
					}
				}
			}()
		}
	}()
}

func main() {
	httpServer := gin.Default()
	httpServer.StaticFS("/public", http.Dir("public"))
	httpServer.GET("/create", createNewConnection)
	httpServer.GET("/wstotcp/:id", wstotcp)
	httpServer.Run(":8080")
}
