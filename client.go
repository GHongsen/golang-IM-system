package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func CreateClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil
	}
	client.conn = conn

	return client
}

// 接收服务器回复的消息
func (client *Client) ReadResponse() {
	io.Copy(os.Stdout, client.conn)
}

var serverIp string
var serverPort int

// 初始化设置 ./client -ip *** -port ***
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口")
}

func main() {
	// 命令行解析
	flag.Parse()

	client := CreateClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>  连接（", serverIp, serverPort, "）服务器失败！")
		return
	}
	fmt.Println(">>>>>>  连接（", serverIp, ":", serverPort, "）服务器成功！")

	// 开启一个goroutine监听server输出
	go client.ReadResponse()

	client.Run()
}

func (client *Client) Run() {
	for client.flag != 0 {
		for !client.menu() {
		}

		switch client.flag {
		case 1:
			//公聊模式
			client.PublicChat()
		case 2:
			// 私聊模式
			client.PrivateChat()
		case 3:
			// 更改用户名
			client.UpdateName()
		}
	}
}

// 显示菜单
func (client *Client) menu() bool {

	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scan(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("请输入正确的数字编号")
		return false
	}
}

// 公聊
func (client *Client) PublicChat() {
	var sendMsg string
	for {
		sendMsg = ""
		fmt.Println("请输入聊天内容，exit退出")
		fmt.Scanln(&sendMsg)
		if sendMsg == "exit" {
			return
		}

		_, err := client.conn.Write([]byte(sendMsg + "\n"))
		if err != nil {
			fmt.Println("client conn.Write err:", err)
			break
		}
	}
}

// 私聊
func (client *Client) PrivateChat() {
	var sendMsg string
	var sendName string

	fmt.Println("选输入你要私聊用户的用户名，exit退出")
	client.GetOnlineList()

	fmt.Scanln(&sendName)
	for sendName != "exit" {
		sendMsg = ""
		fmt.Println("选输入你对他私聊的话，exit退出")
		fmt.Scanln(&sendMsg)
		if sendMsg == "exit" {
			return
		}
		_, err := client.conn.Write([]byte("to " + sendName + ":" + sendMsg + "\n"))
		if err != nil {
			fmt.Println("client conn.Write err:", err)
			break
		}
	}
}

// 更改用户名
func (client *Client) UpdateName() bool {
	fmt.Println("请输入你想要修改的用户名：")
	fmt.Scanln(&client.Name)

	sendMsg := "rename " + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("client conn.Write err:", err)
		return false
	}
	return true
}

// 获取所有在线用户列表
func (client *Client) GetOnlineList() {
	_, err := client.conn.Write([]byte("list\n"))
	if err != nil {
		fmt.Println("client conn.Write err:", err)
	}
}
