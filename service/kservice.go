package service

import "github.com/lvqiuxia/zmqserver/msgdefine"

type IService interface{
	Type()   string
	CatLog() string
	//ProcessMsg(msg [][]byte, rep []byte)([][]byte,error)
}

type KService struct {
	KComponent
	Stateless bool
	//Workers   uint64
	//Async     bool
	//SvcState  uint32
}

func(k *KService)Type() string{
	return "KService"
}

func(k *KService)CatLog() string{
	return "Service"
}

func(s *KService)ProcessMsg(msg [][]byte, rep []byte)([][]byte,error){

	//解析请求消息，并保存消息头
	request, err := msgdefine.ParseRequestMsg(msg)
	if err != nil{
		return nil, err
	}

	response := make([][]byte, 0)

	for _,temp := range request.RawHeader{
		response = append(response,temp)
	}

	repData := rep
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

//func(k *KService) Name() string {
//	return ""
//}
//
//func(k *KService) State() ComponentState{}
//func(k *KService) Init() bool{}
//func(k *KService) Open() bool{}
//func(k *KService) Close(){}
//func(k *KService) Dump(obj *AppConfig){}
//func(k *KService) SetParent(parent IComponent) bool{}
//func(k *KService) Parent() IComponent{}
//func(k *KService) Root() IComponent{}
//func(k *KService) App() IComponent{}
//func(k *KService) IsActor() bool{}
//func(k *KService) AddComponent(component IComponent) bool{}
//func(k *KService) DelComponent(name string) bool{}
//func(k *KService) GetComponent(name string) IComponent{}
//func(k *KService) GetComponents(component []IComponent){}
//func(k *KService) GetAllComponents(component []IComponent){}