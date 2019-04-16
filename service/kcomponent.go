package service

import (
	"errors"
	"fmt"
	"reflect"
)

/* *******************************************************************
   Name:         IComponent()
   Description:  KComponent的接口及方法
   Date:         2019.4.11
   Auther:       lvqiuxia
   Input/Output:
   Others:
********************************************************************** */
type IComponent interface{
	Name() string
	Type() string
	CatLog() string
	State() ComponentState
	Init() bool
	Open() bool
	Close()
	Dump(obj *AppConfig)
	SetParent(parent IComponent) bool
	Parent() IComponent
	Root() IComponent
	App() IComponent
	IsActor() bool
	AddComponent(component IComponent) bool
	DelComponent(name string) bool
	GetComponent(name string) IComponent
	GetComponents(component []IComponent)
	GetAllComponents(component []IComponent)
		OnInit() bool
		OnOpen() bool
		OnClose()
}

//基类的结构体
type KComponent struct{
	MName           string
	MState          ComponentState
	MParent         IComponent
	Components      []IComponent
}

//组件状态
type ComponentState int
const (
	NDEF ComponentState = iota    ///< 未初始化状态
	INIT                          ///< 初始化状态
	OPEN                          ///< 打开状态
	CLOSE                         ///< 关闭状态
)

//获取组件名称，return组件名称
func (k *KComponent) Name() string{
	return k.MName
}

//获取组件类型，return组件类型
func (k *KComponent) Type() string{
	err := errors.New("invalid Type")
	panic(err)
	return err.Error()
}

//获取组件类别，return组件类别
func (k *KComponent) CatLog() string{
	err := errors.New("invalid CatLog")
	panic(err)
	return err.Error()
}

//获取组件状态，return组件状态
func (k *KComponent) State() ComponentState{
	return k.MState
}

//组件初始化，返回true/false
func (k *KComponent) Init() bool{
	fmt.Printf("KComponent init(%s)\n", k.Name())
	//检验组件状态
	if k.MState != NDEF {
		fmt.Printf("KComponent init(%s): invalid state(%d)",k.Name(),k.MState)
		return false
	}

	//组件自身初始化
	//if k.RealOnInit(k) == false{
	if k.OnInit() == false {
		fmt.Printf("KComponent init(%s): onInit invalid",k.Name())
		return false
	}

	//依次初始化所包含的子组件
	for _,temp := range k.Components{
		if temp.Init() == false{
			fmt.Printf("KComponent init: component(%s:%s) init failed",temp.Type(),temp.Name())
			return false
		}
	}

	//更新组件状态
	k.MState = INIT
	fmt.Printf("KComponent init(%s) ok\n",k.MName)

	return true
}

//打开组件，返回true/false
func (k *KComponent) Open() bool{
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
func (k *KComponent) Close(){
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
func (k *KComponent) Dump(obj *AppConfig){
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

//设置父组件，入参是父组件，return true/false
func (k *KComponent) SetParent(parent IComponent) bool {
	if k.MParent != nil{
		fmt.Printf("KComponent setParent(%s):parent has already been set",k.Name())
	}

	k.MParent = parent
	return true
}

//获取父组件，return 父组件
func (k *KComponent) Parent() IComponent {
	return k.MParent
}

//获取根组件
func (k *KComponent) Root() IComponent {
	var component IComponent = k

	for{
		if component != nil && component.Parent() != nil{
			component = component.Parent()
		}else{
			break
		}
	}
	return component
}

//获取顶级应用组件
func (k *KComponent) App() IComponent {
	return k.Root()
}

//获取包含本组件的活动（actor）组件
func (k *KComponent) Actor() IComponent {
	var component  = IComponent(k)
	for{
		if component != nil && component.IsActor() == false{
			component = component.Parent()
		}else{
			return component
		}
	}
}

////reactor
//func (k *KComponent) Reactor(){
//	return
//}

//是否是活动组件 ，return true/false
func (k *KComponent) IsActor() bool{
	return false
}

//添加子组件，入参是组件对象，return true/false
func (k *KComponent) AddComponent(component IComponent) bool{
	if component == nil || component.Name() == "" {
		fmt.Printf("KComponent AddComponent: invalid component")
		return false
	}

	for _,temp := range k.Components{
		if temp.Name() == component.Name() {
			fmt.Printf("KComponent AddComponent: component already exist")
			return false
		}
	}

	k.Components = append(k.Components,component)
	component.SetParent(k)
	return true
}

//删除指定名称的子组件，入参是名字，return true/false
func (k *KComponent) DelComponent(name string) bool {
	if name == ""{
		fmt.Printf("KComponent delComponent: invalid component name")
		return false
	}

	for i,temp := range k.Components {
		if temp.Name() == name {
			k.Components = append(k.Components[:i],k.Components[i+1:]...)
			break
		}
	}
	return true
}

//获取指定名称的子组件，入参是名称，return 组件对象
func (k *KComponent) GetComponent(name string) IComponent{
	for _,temp := range k.Components {
		if temp.Name() == name {
			return temp
		}
	}
	fmt.Printf("KComponent GetComponent: invalid component name")
	return nil
}

//获取所有子组件，返回组件集合
func (k *KComponent) GetComponents(components []IComponent){
	components = k.Components
	return
}

//递归获取所有子孙组件
func (k *KComponent) GetAllComponents(allComponents []IComponent){
	for _, temp := range k.Components{
		allComponents = append(allComponents, temp)
	}
	for _, temp := range k.Components{
		temp.GetAllComponents(allComponents)
	}

	return
}

//组件初始化 return true/false
func (k *KComponent) OnInit() bool{
	return true
}

//打开组件 return true/false
func (k *KComponent) OnOpen() bool{
	return true
}

//关闭组件
func (k *KComponent) OnClose(){
	return
}

//输出组件详细信息，输出流 //todo:
func (k *KComponent) OnDump(obj *AppConfig){
	return
}

//输出组件详细信息，输出流 //todo:
func (k *KComponent) RealOnInit(child interface{}) bool{
	ref := reflect.ValueOf(child)
	method := ref.MethodByName("OnInit")
	if (method.IsValid()) {
		r := method.Call(make([]reflect.Value, 0))
		return r[0].Bool()
	}
	return false
}