package service

import (
	"errors"
	"fmt"
	"github.com/alecthomas/gozmq"
	"github.com/golang/protobuf/proto"
	"github.com/lvqiuxia/zmqserver/msgdefine"
	"github.com/lvqiuxia/zmqserver/pbmsg"
	"os"
	"strconv"
	"strings"
	"time"
)

//type ServerInterface interface{
//	OnInit() bool
//	OnOpen() bool
//	OnClose()
//	AddComponent(component KService)
//	AddService(service KService)
//	OnService()
//	OnForward()
//	OnDump()
//	GetServices() []KService
//	GetServiceID() string
//	GetServiceAddress() string
//	GetDomain() string
//	Type() string
//	CatLog() string
//}

type Server struct{
	KActor
	Domain        string
	RegisterType  string
	AddressPath   string
	Address       string
	MyID          string
	TotalRequests uint64
	Services      map[string]KService
	Backends      []byte
	SvcSock       interface{}
}

func (k *Server) OnInit() bool{
	fmt.Println("aaaaaaaaaaa")
	if k.KActor.OnInit() == false{
		fmt.Println("KServer onInit: base failed")
		return false
	}
	if k.Address == ""{
		hostName,_ := os.Hostname()
		processId := strconv.Itoa(os.Getpid())
		pid := hostName + ":" + processId
		k.MyID = pid
		k.AddressPath = "/tmp/kserver"+":"+ k.App().Name() +":"+ k.MyID
		k.Address = "ipc://" + k.AddressPath
	}
	if k.Services == nil{
		fmt.Printf("KServer(%s): no service available",k.Name())
		return false
	}

	//todo:
	register := k.GetComponent("register")
	if register == nil{
	}

	return true
}

func (k *Server) OnOpen() bool{
	fmt.Printf("come into server %s onOpen",k.Name())
	if k.KActor.OnOpen() == false {
		fmt.Println("KServer OnOpen: base failed")
		return false
	}
	k.ZmqServerProc()
	return true
}

func (k *Server) OnClose(){
	return
}

func (k *Server) AddComponent(component KService){
	if component.CatLog() == "Service"{
		k.AddService(component)
	}
	k.KActor.KComponent.AddComponent(&component)
	return
}

func (k *Server) AddService(service KService){
	if _,ok := k.Services[service.Name()];ok{
		return
	}
	if k.Services == nil{
		k.Services = make(map[string]KService,0)
	}
	k.Services[service.Name()] = service

	//todo: 添加work
	return
}

func (k *Server) OnService(){
	return
}

func (k *Server) OnForward(){
	return
}

func (k *Server) OnDump(){
	return
}

func (k *Server) GetServices() []KService{
	services := make([]KService,0)
	for _,temp := range k.Services{
		services = append(services,temp)
	}
	return services
}

func (k *Server) GetServiceID() string{
	return k.MyID
}

func (k *Server) GetServiceAddress() string{
	return k.Address
}

func (k *Server) GetDomain() string{
	return k.Domain
}

func (k *Server) Type() string{
	return "KServer"
}

func (k *Server) CatLog() string{
	return "Server"
}

func (k *Server) Run() {
	if k.Init() == false{
		fmt.Println("Server Init failed")
	}
	if k.Open() == false{
		fmt.Println("Server open failed")
	}
	return
}

//server端的处理
func (k *Server) ZmqServerProc(){
	context, _ := gozmq.NewContext()

	socket, _ := context.NewSocket(gozmq.ROUTER)
	socket.SetRcvTimeout(time.Second * 2)//todo:设置消息接收超时时间，保证通道有消息
	defer context.Close()
	defer socket.Close()
	err := socket.Bind(k.Address)
	if err != nil{
		fmt.Println(err)
		return
	}

	//todo ********************************创建client发心跳信息*****************************
	go k.DoHeartBeat(context)

	//TODO ********************************************************************************

	responseCh := make(chan [][]byte, 1)

	//todo: Wait for messages
	for {
		fmt.Println("ZmqServerProc:message come in")
		//msg, _ := socket.RecvMultipart(0)
		select{
		case rep := <-responseCh:
			// send reply back to client
			fmt.Println("come into case")
			//rep := <-responseCh
			err := socket.SendMultipart(rep, 0)
			if err != nil{
				fmt.Println("ZmqServerProc: RecvMultipart msg failed",err)
				return
			}
			fmt.Println("ZmqServerProc: RecvMultipart msg success")
		default:
			fmt.Println("come into default")
			msg,err := socket.RecvMultipart(0)
			if err != nil{
				fmt.Println("ZmqServerProc: RecvMultipart msg failed",err)
				break
			}
			//起一个协程去处理
			go k.HandlerRecvMsgInGoRoutine(responseCh,msg)
		}
	}
}

//在gc里面处理接收消息
func (k *Server) HandlerRecvMsgInGoRoutine(responseCh chan [][]byte,msg [][]byte){

	if responseCh == nil || msg == nil{
		fmt.Println("HandlerRecvMsgInGoRoutine: invalid msg")
		return
	}
	//do work
	reply,err := k.HandlerMsgProc(msg)
	if err != nil{
		fmt.Println("HandlerRecvMsgInGoRoutine: HandlerMsgProc failed", err)
	}

	responseCh <- reply
}



func (k *Server) HandlerMsgProc(msg [][]byte) ([][]byte,error) {

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
	response, err := k.SendMsgToService(topic,msg)
	if err != nil{
		return nil, err
	}

	return response,nil
}


func (k *Server) SendMsgToService(srName string, msg [][]byte) ([][]byte, error){
	var inst KService
	if value, ok := k.Services[srName]; ok{
		inst = value
	}else{
		return nil, errors.New("SendMsgToService: service is not exit")
	}

	response, err := inst.ProcessMsg(msg,k.Backends)
	if err != nil || response == nil{
		return nil, err
	}

	return response, nil
}


func (k *Server) DoHeartBeat(context *gozmq.Context){

	for _, temp := range k.Services{
		k.HeartBeatProc(context,temp.Name())
	}
}


func (k *Server) HeartBeatProc(context *gozmq.Context, name string){
	heart,_ := context.NewSocket(gozmq.DEALER)
	defer heart.Close()
	err := heart.Connect("tcp://localhost:8003")
	if err != nil{
		fmt.Println(err)
		return
	}

	heartMsg, err :=k.CreatHeartBeatRequest(name)
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
			go k.SendHeartBeatMsg(heart,heartMsg)
		}
	}
}


func (k *Server) CreatHeartBeatRequest(name string) ([][]byte, error){

	service   := name
	state     := uint32(1)
	object    := k.MyID
	domain    := k.Domain
	protocal  := "msgp"
	priority  := uint32(1)
	endpoint  := k.Address
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


func (k *Server) SendHeartBeatMsg(socket *gozmq.Socket, msg [][]byte){

	err := socket.SendMultipart(msg,0)
	if err != nil{
		fmt.Println("SendHeartBeatMsg failed",err)
		return
	}
	fmt.Println("SendHeartBeatMsg success")
}