package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播
	Message chan string
}

// CreateServer 创建一个服务器
func CreateServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// ListenMessage 监听消息
func (server *Server) ListenMessage() {
	for {
		msg := <-server.Message

		// 给每一个用户发送消息
		server.mapLock.Lock()
		for _, user := range server.OnlineMap {
			user.C <- msg
		}
		server.mapLock.Unlock()
	}
}

// Broadcast 广播通知
func (server *Server) Broadcast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	server.Message <- sendMsg
}

func (server *Server) Handler(conn net.Conn) {
	// 创建用户
	user := CreateUser(conn, server)
	// 用户存活状态
	isLive := make(chan bool)
	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			// 超时离线关闭会最后发一个回车 可以触发下线逻辑
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("server conn read err:", err)
				return
			}
			// 发出消息
			msg := string(buf[:n-1])
			user.DoMessage(msg)

			isLive <- true
		}
	}()

	// 设置超时下线时间
	for {
		select {
		case <-isLive:
		case <-time.After(time.Minute * 10):
			user.WriteMsg("超时下线")

			close(user.C)
			conn.Close()

			return
		}
	}
}

// Start 启动服务器
func (server *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	// 关闭 Listen
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("net.Listen Close err:", err)
			return
		}
	}(listener)

	// 启动消息监听
	go server.ListenMessage()

	for {
		// 获取连接
		accept, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		go server.Handler(accept)
	}
}
