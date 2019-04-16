package kemp

import (
	"fmt"
	"github.com/lvqiuxia/zmqserver/msgdefine"
)

type IService interface{
	Type()   string
	CatLog() string
	Init() bool
	//ProcessMsg(msg [][]byte, rep []byte)([][]byte,error)
}

type KService struct {
	KComponent
	Stateless bool
	Worker    IComponent
	//Async     bool
	//SvcState  uint32
}

func(k *KService)Type() string{
	return "KService"
}

func(k *KService)CatLog() string{
	return "Service"
}

//组件初始化，返回true/false
func (k *KService) Init() bool{
	fmt.Printf("Server init(%s)\n", k.Name())
	//检验组件状态
	if k.MState != NDEF {
		fmt.Printf("Server init(%s): invalid state(%d)",k.Name(),k.MState)
		return false
	}

	//组件自身初始化
	//if k.RealOnInit(k) == false{
	if k.OnInit() == false {
		fmt.Printf("Server init(%s): onInit invalid",k.Name())
		return false
	}

	//依次初始化所包含的子组件
	for _,temp := range k.Components{
		if temp.Init() == false{
			fmt.Printf("Server init: Server(%s:%s) init failed",temp.Type(),temp.Name())
			return false
		}
	}

	//更新组件状态
	k.MState = INIT
	fmt.Printf("Server init(%s) ok\n",k.MName)

	return true
}

//打开组件，返回true/false
func (k *KService) Open() bool{
	fmt.Printf("KComponent open(%s)\n", k.Name())
	//检验组件状态
	if k.MState != INIT {
		fmt.Printf("KComponent open(%s): invalid state(%d)",k.Name(),k.MState)
		return false
	}

	//打开组件自身
	if k.OnOpen() == false {
		fmt.Printf("KComponent open(%s): onOpen invalid",k.Name())
		return false
	}

	//依次打开所包含的子组件
	for _,temp := range k.Components{
		if temp.Open() == false{
			fmt.Printf("KComponent open: component(%s:%s) open failed",temp.Type(),temp.Name())
			return false
		}
	}

	//更新组件状态
	k.MState = OPEN
	fmt.Printf("KComponent open(%s) ok\n",k.MName)

	return true
}

//关闭组件
func (k *KService) Close(){
	fmt.Printf("KComponent close(%s)", k.Name())
	//检验组件状态
	if k.MState != OPEN {
		fmt.Printf("KComponent close(%s): invalid state(%d)",k.Name(),k.MState)
		return
	}

	//依次关闭所包含的子组件
	for _,temp := range k.Components{
		temp.Close()
	}

	//关闭组件自身
	k.OnClose()

	//更新组件状态
	k.MState = CLOSE
	fmt.Printf("KComponent close(%s) ok",k.MName)

}

//输出组件信息 //todo:是输出什么？
func (k *KService) Dump(obj *AppConfig){
	obj.Name = k.MName
	obj.Type = k.Type()
	obj.State = k.MState
	obj.Actor = k.IsActor()
	k.OnDump(obj)

	//然后依次所包含的子组件信息
	for _,temp := range k.Components{
		childObj := AppConfig{}
		temp.Dump(&childObj)
		obj.Children = append(obj.Children,childObj)
	}
}

func (k *KService) OnInit() bool{
	if k.Parent() != nil && k.Parent().OnInit() == false{
		fmt.Println("KService onOpen: base KComponent onInit failed")
		return false
	}
	k.Worker = k.GetComponent(k.CatLog())

	return true
}

func (k *KService) OnOpen() bool{
	if k.Parent() != nil && k.Parent().OnOpen() == false{
		fmt.Println("KActor onOpen: base KComponent onOpen failed")
		return false
	}

	//k.Svc()
	return true
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

