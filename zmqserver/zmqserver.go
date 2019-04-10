package zmqserver

import (
	"errors"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"github.com/golang/protobuf/proto"
	"github.com/lvqiuxia/zmqserver/msgdefine"
	"github.com/lvqiuxia/zmqserver/pbmsg"
	"github.com/lvqiuxia/zmqserver/service"
	"os"
	"strconv"
	"strings"
	"time"
)

//func init() {
//	//起一个zmq监听服务
//	go StartZmqServer()
//}


/* *******************************************************************
   Function:     StartZmqServer()
   Description:  使用zmq起一个监听服务
   Date:         2019.4.2
   Auther:       lvqiuxia
   Input/Output:
   Others:       4.8优化，处理err
********************************************************************** */
func InitZmqServer() {

	context, _ := zmq.NewContext()

	socket, _ := context.NewSocket(zmq.ROUTER)
	socket.SetRcvTimeout(time.Second * 2)//todo:设置消息接收超时时间，保证通道有消息
	defer context.Close()
	defer socket.Close()
	//err := socket.Bind("tcp://*:5555")
	err := socket.Bind("ipc:///tmp/test")
	if err != nil{
		fmt.Println(err)
		return
	}

	//todo ********************************创建client发心跳信息*****************************
	go DoHeartBeat(context)

	//TODO ********************************************************************************

	responseCh := make(chan [][]byte, 1)

	//todo: Wait for messages
	for {
		fmt.Println("InitZmqServer:message come in")
		//msg, _ := socket.RecvMultipart(0)
		select{
		case rep := <-responseCh:
			// send reply back to client
			fmt.Println("come into case")
			//rep := <-responseCh
			err := socket.SendMultipart(rep, 0)
			if err != nil{
				fmt.Println("InitZmqServer: RecvMultipart msg failed",err)
				return
			}
			fmt.Println("InitZmqServer: RecvMultipart msg success")
		default:
			fmt.Println("come into default")
			msg,err := socket.RecvMultipart(0)
			if err != nil{
				fmt.Println("InitZmqServer: RecvMultipart msg failed",err)
				break
			}
			//起一个协程去处理
			go HandlerRecvMsgInGoRoutine(responseCh,msg)
		}
		//go HandlerMsgInGoRoutine(socket,msg)
	}
}

//在gc里面处理接收消息
func HandlerRecvMsgInGoRoutine(responseCh chan [][]byte,msg [][]byte){

	if responseCh == nil || msg == nil{
		fmt.Println("HandlerRecvMsgInGoRoutine: invalid msg")
		return
	}
	//do work
	reply,err := HandlerMsgProc(msg)
	if err != nil{
		fmt.Println("HandlerRecvMsgInGoRoutine: HandlerMsgProc failed", err)
	}

	responseCh <- reply
}

////在gc里面处理消息
//func HandlerMsgInGoRoutine(socket *zmq.Socket,msg [][]byte){
//	//do work
//	reply := HandlerMsgProc(msg)
//
//	socket.SendMultipart(reply, 0)
//}


/* *******************************************************************
   Function:     HandlerMsgProc()
   Description:  处理zmq接收到的消息,解析消息头发送给service，并return
                 service返回的处理消息
   Date:         2019.4.2
   Auther:       lvqiuxia
   Input/Output:
   Others:       4.8优化，处理err
********************************************************************** */
func HandlerMsgProc(msg [][]byte) ([][]byte,error) {

	var inMsg [][]byte
	for i,temp := range msg{
		if temp != nil && strings.HasPrefix(string(temp),"KMSG") == true{
			inMsg = msg[i:]
			break
		}
	}


	msgHeader, err := msgdefine.GetMessageHeader(inMsg)
	if err != nil || msgHeader == nil{
		return nil,err
	}

	topic := msgHeader.Service
	response, err := SendMsgToService(topic,msg)
	if err != nil{
		return nil, err
	}

	return response,nil
}



/* *******************************************************************
   Function:     SendMsgToService()
   Description:  发送消息给特定的service
   Date:         2019.4.2
   Auther:       lvqiuxia
   Input/Output:
   Others:       4.8优化，处理err
********************************************************************** */
func SendMsgToService(srName string, msg [][]byte) ([][]byte, error){
	var inst service.ServiceInterface
	if value, ok := service.BindServiceNameMap[srName]; ok{
		inst = value
	}else{
		return nil, errors.New("SendMsgToService: service is not exit")
	}

	response, err := inst.ProcessMsg(msg)
	if err != nil || response == nil{
		return nil, err
	}

	return response, nil
}


func DoHeartBeat(context *zmq.Context){
	heart,_ := context.NewSocket(zmq.DEALER)
	defer heart.Close()
	err := heart.Connect("tcp://localhost:8003")
	if err != nil{
		fmt.Println(err)
		return
	}

	heartMsg, err :=CreatHeartBeatRequest()
	if err != nil{
		fmt.Println("CreatHeartBeatRequest failed",err)
		return
	}

	time := time.NewTicker(time.Second * 2)
	for{
		fmt.Println("come into for")
		select {
		case <-time.C:
			fmt.Println("come into case")
			go SendHeartBeatMsg(heart,heartMsg)
		}
	}
}

/* *******************************************************************
   Function:     HeartBeat()
   Description:  构造心跳信息
   Date:         2019.4.9
   Auther:       lvqiuxia
   Input/Output:
   Others:
********************************************************************** */
func CreatHeartBeatRequest() ([][]byte, error){
	//发送的心跳data
	hostName,err := os.Hostname()
	if err != nil{
		return nil, err
	}

	processId := strconv.Itoa(os.Getpid())

	pid := hostName + ":" + processId

	service   := "one"
	state     := uint32(1)
	object    := pid
	domain    := "cloud"
	protocal  := "msgp"
	priority  := uint32(1)
	endpoint  := "ipc:///tmp/test" //"tcp://127.0.0.1:5555"
	stateless := true

	msgInfo := &knaming.KNamingInfo{
		Service:  &service,
		State:    &state,
		Object:   &object,
		Domain:   &domain,
		Protocol: &protocal,
		Priority: &priority,
		Endpoint: &endpoint,
		Stateless:&stateless,
	}

	dataMsg := &knaming.KNamingNotify{}
	sign := uint32(0x444d424e)
	dataMsg.Sign =  &sign
	dataMsg.NameList = append(dataMsg.NameList,msgInfo)


	data,err := proto.Marshal(dataMsg)
	if err != nil{
		//fmt.Println("HeartBeat: KNamingInfo marshal failed")
		return nil, errors.New("HeartBeat: KNamingInfo marshal failed")
	}

	//构造msgHeader
	MsgType := "NTF"
	Method := "register"
	Service := "_naming_"
	Instance := "*"
	Domain := "cloud"
	Protocol := "pbuf"
	SequenceNo := uint64(1)
	msg,err := msgdefine.CreateRequestMsg(MsgType, Method, Service, Instance, Domain,
		Protocol, SequenceNo, data)
	if err != nil{
		//fmt.Println("HeartBeat: CreateRequestMsg failed",err)
		return nil,err
	}

	req := make([][]byte,0)
	req = append(req, msg.MsgHead)
	req = append(req, msg.DataHead)
	req = append(req, msg.Data)

	//调用socket发送接口
	//time := time.NewTimer(time.Second * 2)
	return req,nil
}


/* *******************************************************************
   Function:     HeartBeat()
   Description:  server发送心跳
   Date:         2019.4.9
   Auther:       lvqiuxia
   Input/Output:
   Others:
********************************************************************** */
func SendHeartBeatMsg(socket *zmq.Socket, msg [][]byte){

	err := socket.SendMultipart(msg,0)
	if err != nil{
		fmt.Println("SendHeartBeatMsg failed",err)
		return
	}
	fmt.Println("SendHeartBeatMsg success")
}