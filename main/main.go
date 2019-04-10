package main

import (
	"github.com/lvqiuxia/zmqserver/client"
	_ "github.com/lvqiuxia/zmqserver/service"
	"github.com/lvqiuxia/zmqserver/zmqserver"

)


func main(){
		//初始化httpserver
	go client.InitHttpServer()

	//起一个zmq监听服务
	zmqserver.InitZmqServer()


}