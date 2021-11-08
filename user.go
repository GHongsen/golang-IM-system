package main

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// Online 用户上线
func (user *User) Online() {
	// 用户上线操作加入到在线用户表中
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Addr] = user
	user.server.mapLock.Unlock()

	// 广播通知用户上线
	user.server.Broadcast(user, "已上线")
}

// Offline 用户下线
func (user *User) Offline() {
	// 用户下线将用户从Online列表中删除
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Addr)
	user.server.mapLock.Unlock()

	// 广播通知用户下线
	user.server.Broadcast(user, "下线")
}

// WriteMsg 将信息输出到面板
func (user *User) WriteMsg(msg string) {
	_, err := user.conn.Write([]byte(msg + "\n"))
	if err != nil {
		fmt.Println("user conn read err:", err)
		return
	}
}

// DoMessage 用户发送消息
func (user *User) DoMessage(msg string) {
	// 指令判断
	if msg == "list" {
		user.server.mapLock.Lock()
		var userList string
		for _, info := range user.server.OnlineMap {
			userList += "[" + info.Addr + "]" + info.Name + ":在线\n"
		}
		// 将拼接出来的信息写到面板上
		user.WriteMsg(userList[:len(userList)-1])
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename " {
		newName := msg[7:]
		var b = false
		// 判断重名
		for _, info := range user.server.OnlineMap {
			if info.Name == newName {
				b = true
				break
			}
		}
		if b {
			user.WriteMsg("名字已存在，名称修改失败")
		} else {
			user.WriteMsg("您的用户名以改名为：" + newName)
			user.Name = newName
		}
	} else if regexp.MustCompile("^to .+:.*$").MatchString(msg) {
		// 找到 ：的位置
		charIndex := strings.Index(msg, ":")
		// 截取出接收消息的用户名
		callName := msg[3:charIndex]
		// 判断是否存在并发送
		for _, info := range user.server.OnlineMap {
			if info.Name == callName {
				info.WriteMsg("[" + user.Addr + "]" + user.Name + " 私聊您说" + msg[charIndex:])
				return
			}
		}

		user.WriteMsg("未找到用户：[" + msg[3:charIndex] + "]")

	} else {
		user.server.Broadcast(user, msg)
	}
}

// CreateUser 创建用户对象
func CreateUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	user.Online()

	// 启动监听
	go user.ListenMessage()

	return user
}

// ListenMessage 监听用户收到消息
func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		user.WriteMsg(msg)
	}
}
