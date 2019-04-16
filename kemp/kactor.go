package kemp

import "fmt"

//type IActor interface{
//	OnInit() bool
//	OnOpen() bool
//	OnClose()
//	Svc()
//	IsActor() bool
//	Type() string
//	CatLog() string
//}

type KActor struct{
	KComponent
	Worker   IComponent
}

func (k *KActor) OnInit() bool{
	if k.KComponent.OnInit() == false{
		fmt.Println("KActor onOpen: base KComponent onInit failed")
		return false
	}
	k.Worker = k.GetComponent(k.CatLog())

	return true
}

func (k *KActor) OnOpen() bool{
	if k.KComponent.OnOpen() == false{
		fmt.Println("KActor onOpen: base KComponent onOpen failed")
		return false
	}

	k.Svc()
	return true
}

func (k *KActor) OnClose() {
	k.KComponent.OnClose()

	return
}

func (k *KActor) Svc(){
	//todo: do work
	return
}

func (k *KActor) IsActor() bool{
	return true
}

func (k *KActor) Type() string{
	return "KActor"
}

func (k *KActor) CatLog() string{
	return "Actor"
}

