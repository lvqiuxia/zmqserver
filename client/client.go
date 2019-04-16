package client

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"github.com/lvqiuxia/zmqserver/msgdefine"
	"log"
	"net/http"
	"time"
)

//func init(){
//	//初始化httpserver
//	InitHttpServer()
//}


//③处理请求，返回结果
type handleFunc struct {
	ZmqMsg   *zmq.Socket
}

// ServeHTTP
func (f handleFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//fmt.Print("hello world, i am lqx\n")

	//2、起zmq连接服务
	sendContsxt,sendSock, err := InitZmqClient()
	if err != nil{
		fmt.Println("InitZmqClient failed",err)
		return
	}
	f.ZmqMsg = sendSock

	//3、处理消息并返回给http客户端
	err = f.ReceiveHttpMsg(w,r)
	if err != nil{
		fmt.Println("ReceiveClientMsg failed",err)
		return
	}

	sendSock.Close()
	sendContsxt.Close()
}

//处理消息并返回给http客户端
func (f handleFunc) ReceiveHttpMsg(w http.ResponseWriter, r *http.Request) error{
	switch r.URL.Path {
	case "/serviceone":
		fmt.Print("serviceone msg come in\n")

		reply, err := f.StartZmqForReqMsg()
		if err != nil{
			return err
		}
		fmt.Fprint(w,reply)

	}
	return nil
}

/* *******************************************************************
   Function:     InitHttpServer()
   Description:  自定义server, 然后使用自定义的server的ListenAndServe()。
   Date:         2019.4.2
   Auther:       lvqiuxia
   Input/Output:
   Others:       4.8优化，处理err
********************************************************************** */
func InitHttpServer() {
	mux := http.NewServeMux()

	handler := handleFunc{}

	//①路由注册
	mux.Handle("/", handler)

	s := &http.Server{
		Addr: ":8085",
		Handler: mux, //指定路由或处理器，不指定时为nil，表示使用默认的路由DefaultServeMux
		ReadTimeout: 20 * time.Second,
		WriteTimeout: 20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	//②服务监听
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}


//使用zmq进行消息请求
func (f handleFunc) StartZmqForReqMsg() ([]byte, error){

	// createRequestMsg
	MsgType := "REQ"
	Method := "get"
	Service := "testService"
	Instance := "*"
	Domain := "local"
	Protocol := "msgp"
	SequenceNo := uint64(1)
	data := []byte{1,2,3,4,5}
	msg,err := msgdefine.CreateRequestMsg(MsgType, Method, Service, Instance, Domain,
		Protocol, SequenceNo, data)
	if err != nil{
		return nil, err
	}

	req := make([][]byte,0)
	req = append(req, msg.MsgHead)
	req = append(req, msg.DataHead)
	req = append(req, msg.Data)

	err = f.ZmqMsg.SendMultipart(req, 0)
	if err != nil{
		return nil,nil
	}

	fmt.Println("Sending ", msg.Data)

	// Wait for reply:
	reply, err := f.ZmqMsg.RecvMultipart(0)
	if err != nil{
		return nil, nil
	}

	repData,err := msgdefine.ParseResponseMsg(reply)
	if err != nil{
		return nil, err
	}

	fmt.Println("Received ", repData.Data)

	if repData.Data != nil{
		return repData.Data, nil
	}else{
		return repData.Attr, nil
	}
}


/* *******************************************************************
   Function:     InitZmqClient()
   Description:  使用zmq起一个客户端，连接到一个端口并进行消息发送
   Date:         2019.2.23
   Auther:       lvqiuxia
   Input/Output:
   Others:       4.8优化，处理err
********************************************************************** */
func InitZmqClient() (*zmq.Context,*zmq.Socket, error) {
	context, err := zmq.NewContext()
	if err != nil{
		return nil, nil, err
	}
	socket, err := context.NewSocket(zmq.REQ)
	if err != nil{
		return nil, nil, err
	}


	fmt.Printf("Connecting to zmq server...\n")
	err = socket.Connect("tcp://localhost:8003")//todo: 实验发现填写localhost才可以通讯成功
	if err != nil{
		return nil, nil, err
	}
	return context, socket, nil
}
