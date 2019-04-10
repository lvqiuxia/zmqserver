package service

import "github.com/lvqiuxia/zmqserver/msgdefine"


func init(){
	//init service
	serviceOne := &ServiceOne{
		ServiceName: "one",
		ServiceID:    "1",
	}

	serviceTwo := &ServiceTwo{
		ServiceName: "two",
		ServiceID:    "2",
	}

	BindServiceNameMap["one"] = serviceOne
	BindServiceNameMap["two"] = serviceTwo
}

//servicename与对象的绑定信息
var BindServiceNameMap = make(map[string] ServiceInterface,0)
//var BindServiceNameMap map[string] ServiceInterface

//定义service的方法
type ServiceInterface interface {
	ProcessMsg(msg [][]byte)([][]byte,error)
}

//service对象
type ServiceOne struct {
	ServiceName  string
	ServiceID    string
}

func(s *ServiceOne)ProcessMsg(msg [][]byte)([][]byte,error){

	//解析请求消息，并保存消息头
	request, err := msgdefine.ParseRequestMsg(msg)
	if err != nil{
		return nil, err
	}

	response := make([][]byte, 0)

	for _,temp := range request.RawHeader{
		response = append(response,temp)
	}

	repData := []byte{9,8,7,6,5,4}
	sts := uint64(0)   //todo: 返回状态
	repHead, err := msgdefine.CreateResponseMsg(request.MsgHead,repData,sts)
	if err != nil{
		return nil, err
	}
	response = append(response,repHead.MsgHead)
	response = append(response,repHead.DataHead)
	response = append(response,repData)

	return response, nil
}

//service对象
type ServiceTwo struct {
	ServiceName  string
	ServiceID    string
}

func(s *ServiceTwo)ProcessMsg(msg [][]byte)([][]byte,error){

	return nil,nil
}


