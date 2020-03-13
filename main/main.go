package main

import (
	"github.com/lvqiuxia/zmqserver/client"
	"github.com/lvqiuxia/zmqserver/kemp"
	_ "github.com/lvqiuxia/zmqserver/kemp"
	"os"
	"os/signal"
)


func main(){
		//初始化httpserver
	go client.InitHttpServer()

	//起一个zmq监听服务
	//zmqserver.InitZmqServer()

	server := kemp.Server{}
	server.MName = "zmqServer"
	server.MState = kemp.NDEF
	server.Domain = "cloud"

	service1 := kemp.KService{}
	service1.MName = "testService"
	service1.MState = kemp.NDEF

	server.AddComponent(service1)
	server.Run()
	//server.OnInit()


	c := make(chan os.Signal)
    signal.Notify(c)
	select{
	case <-c:
		return
	}
}