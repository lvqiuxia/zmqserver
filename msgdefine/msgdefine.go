package msgdefine

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

//发送的消息
type Message struct {
	MsgHead  []byte //MsgHeader转为比特流
	DataHead []byte //DataHeader转为比特流
	Data     []byte
}

//requestMsg
type RequestMsg struct{
	RawHeader     [][]byte
	MsgHead       ReqMsgHeader
	DataHead      DataHeader
	AttrHead      DataHeader
	Data          []byte
	Attr          []byte
}

//requestMsg
type ResponseMsg struct{
	RawHeader     [][]byte
	MsgHead       RepMsgHeader
	DataHead      DataHeader
	AttrHead      DataHeader
	Data          []byte
	Attr          []byte
}


//request消息头（第一帧）
type ReqMsgHeader struct {
	HeaderType    string  //必须是KMSG
	MsgType       string
	Method        string
	Service       string
	Instance      string
	Domain        string
	Protocol      string  //规约
	SequenceNo    uint64
}

//response消息头（第一帧）
type RepMsgHeader struct {
	HeaderType    string  //必须是KMSG
	MsgType       string
	Method        string
	Service       string
	Status        string
	SequenceNo    uint64
}


//数据头（第二帧）
type DataHeader struct {
	HeaderType    string
	DataLen       uint64
	Protocol      string
}


//消息头转为长度固定可二进制序列化的形式（第一帧）
type MsgHeaderTrans struct {
	HeaderType    uint64  //必须是KMSG
	MsgType       uint64
	Method        uint64
	Service       uint64
	Instance      uint64
	Domain        uint64
	Protocol      uint64  //规约
	SequenceNo    uint64
}

//数据头转为长度固定可二进制序列化的形式（第二帧）
type DataHeaderTrans struct {
	HeaderType    uint64
	DataLen       uint64
	Protocol      uint64
}


/* *******************************************************************
   Function:     GetMessageHeader()
   Description:  创建请求消息
   Date:         2019.4.4
   Auther:       lvqiuxia
   Input/Output:
   Others:       4.8优化，处理err
********************************************************************** */
func CreateRequestMsg(MsgType string, Method string, Service string, Instance string, Domain string,
	Protocol string, SequenceNo uint64, data []byte) (*Message, error){

	if MsgType == "" || Method == "" || Service == "" || Instance == "" || Protocol == "" {
		return nil, errors.New("CreateRequestMsg: invalid request msg")
	}

	seqNo:= strconv.Itoa(int(SequenceNo))
	msgHead := "KMSG"+"|"+MsgType+"|"+Method+"|"+Service+"|"+Instance+"|"+Domain+"|"+Protocol+"|"+seqNo

	dataLen := strconv.Itoa(len(data))
	dataHead := "DATA"+"|"+dataLen+"|"+Protocol

	msg := &Message{}
	msg.MsgHead = []byte(msgHead)
	msg.DataHead = []byte(dataHead)
	msg.Data = data

	return msg, nil
}


/* *******************************************************************
   Function:     CreateResponseMsg()
   Description:  创建response消息
   Date:         2019.4.2
   Auther:       lvqiuxia
   Input/Output:
   Others:       4.8优化，处理err
********************************************************************** */
func CreateResponseMsg(request ReqMsgHeader, data []byte, sts uint64) (*Message, error){
	if data == nil{
		return nil, errors.New("CreateResponseMsg: invalid data")
	}
	if request.Method == "" || request.Service == "" || request.Protocol == ""{
		return nil, errors.New("CreateResponseMsg: invalid request msg")
	}

	status:= strconv.Itoa(int(sts))
	seqNo:= strconv.Itoa(int(request.SequenceNo))
	msgType := "REP"
	msgHead := "KMSG"+"|"+msgType+"|"+request.Method+"|"+request.Service+"|"+status+"|"+seqNo

	dataLen := strconv.Itoa(len(data))
	dataHead := "DATA"+"|"+dataLen+"|"+request.Protocol

	msg := &Message{}
	msg.MsgHead = []byte(msgHead)
	msg.DataHead = []byte(dataHead)
	msg.Data = data
	return msg,nil
}


/* *******************************************************************
   Function:     GetMessageHeader()
   Description:  获取消息头
   Date:         2019.4.4
   Auther:       lvqiuxia
   Input/Output:
   Others:       4.8优化，处理err
********************************************************************** */
func GetMessageHeader(msg [][]byte) (*ReqMsgHeader, error){
	for temp := range msg{
		if strings.HasPrefix(string(msg[temp]),"KMSG"){

			mhs := strings.Split(string(msg[temp]),"|")

			if len(mhs) < 7{
				//fmt.Println("GetMessageHeader: invalid message header!")
				return nil, errors.New("GetMessageHeader: invalid message header len")
			}

			for _,mh := range mhs{
				if mh == ""{
					//fmt.Println("GetMessageHeader: invalid message header!")
					return nil, errors.New("GetMessageHeader: invalid message header who is nil")
				}
			}
			if mhs[0] != "KMSG" || strings.Contains("REQNTFEVT",mhs[1]) == false{
				//fmt.Println("GetMessageHeader: invalid message header!")
				return nil, errors.New("GetMessageHeader: invalid message header REQ")
			}

			seqNo,_ := strconv.Atoi(mhs[7])

			header := ReqMsgHeader{
				HeaderType:mhs[0],
				MsgType:mhs[1],
				Method:mhs[2],
				Service:mhs[3],
				Instance:mhs[4],
				Domain:mhs[5],
				Protocol:mhs[6],
				SequenceNo:uint64(seqNo),
			}
			return &header,nil
		}else{
			//fmt.Println("GetMessageHeader: invalid message header!")
			return nil, errors.New("GetMessageHeader: invalid message header with KMSG")
		}
	}
	return nil, errors.New("GetMessageHeader: invalid message header")
}


//msgHead编码
func EncodeMsgHeadToByte(msg ReqMsgHeader)[]byte{

	headType,_ := strconv.Atoi(msg.HeaderType)
	msgType,_ := strconv.Atoi(msg.MsgType)
	method,_ := strconv.Atoi(msg.Method)
	service,_ := strconv.Atoi(msg.Service)
	instance,_ := strconv.Atoi(msg.Instance)
	domain,_ := strconv.Atoi(msg.Domain)
	protocol,_ := strconv.Atoi(msg.Protocol)
	m := MsgHeaderTrans{
		HeaderType:uint64(headType),
		MsgType:uint64(msgType),
		Method:uint64(method),
		Service:uint64(service),
		Instance:uint64(instance),
		Domain:uint64(domain),
		Protocol:uint64(protocol),
		SequenceNo:msg.SequenceNo,
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, m); err != nil {
		log.Fatal("binary write error:", err)
	}
	fmt.Println(buf.Bytes())
	return buf.Bytes()
}


//dataHead编码
func EncodeDataHeadToByte(msg DataHeader)[]byte{

	headType,_ := strconv.Atoi(msg.HeaderType)
	protocol,_ := strconv.Atoi(msg.Protocol)
	m := DataHeaderTrans{
		HeaderType:uint64(headType),
		DataLen:msg.DataLen,
		Protocol:uint64(protocol),
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, m); err != nil {
		log.Fatal("binary write error:", err)
	}
	fmt.Println(buf.Bytes())
	return buf.Bytes()
}


/* *******************************************************************
   Function:     ParseRequestMsg()
   Description:  解析请求消息，并保存消息头部，
   Date:         2019.4.4
   Auther:       lvqiuxia
   Input/Output:
   Others:       4.8优化，处理err
********************************************************************** */
func ParseRequestMsg(message [][]byte) (*RequestMsg,error) {
	request := RequestMsg{}

	if message == nil {
		//fmt.Println("ParseRequestMsg: null message!")
		return nil, errors.New("ParseRequestMsg: null message")
	}
	frames := len(message)
	idx := 0

	//process router header
	for _,temp := range message{
		if temp != nil && strings.HasPrefix(string(temp),"KMSG") == true{
			break
		}else{
			request.RawHeader = append(request.RawHeader,temp)
		}
		idx ++
	}

	//process message header
	mhs := strings.Split(string(message[idx]),"|")

	if len(mhs) < 7{
		//fmt.Println("ParseRequestMsg: invalid message header!")
		return nil, errors.New("ParseRequestMsg: null message header")
	}

	for _,mh := range mhs{
		if mh == ""{
			//fmt.Println("ParseRequestMsg: invalid message header!")
			return nil, errors.New("ParseRequestMsg: null message header")
		}
	}
	if mhs[0] != "KMSG" || strings.Contains("REQNTFEVT",mhs[1]) == false{
		//fmt.Println("ParseRequestMsg: invalid message header!")
		return nil, errors.New("ParseRequestMsg: null message header")
	}

	seqNo,_ := strconv.Atoi(mhs[7])
	request.MsgHead.HeaderType = mhs[0]
	request.MsgHead.MsgType = mhs[1]
	request.MsgHead.Method = mhs[2]
	request.MsgHead.Service = mhs[3]
	request.MsgHead.Instance = mhs[4]
	request.MsgHead.Domain = mhs[5]
	request.MsgHead.Protocol = mhs[6]
	request.MsgHead.SequenceNo = uint64(seqNo)

	//process attr or data header
	idx ++
	if idx < frames{
		adhs := strings.Split(string(message[idx]),"|")
		if len(adhs) < 2{
			//fmt.Println("ParseRequestMsg: invalid header!")
			return nil, errors.New("ParseRequestMsg: null message data")
		}
		if strings.Contains("ATTRDATA",string(adhs[0])) == false {
			//fmt.Println("ParseRequestMsg: invalid header!")
			return nil, errors.New("ParseRequestMsg: null message data")
		}

		dLen,_ := strconv.Atoi(string(adhs[1]))
		if adhs[0] == "ATTR"{
			request.AttrHead.DataLen = uint64(dLen)
		}
		if adhs[0] == "DATA"{
			request.DataHead.DataLen = uint64(dLen)
		}
		if dLen > 0 {
			idx ++
			if idx >= frames || len(message[idx]) != dLen {
				//fmt.Println("ParseRequestMsg: invalid message data!")
				return nil, errors.New("ParseRequestMsg: null message data")
			}
			if adhs[0] == "ATTR"{
				request.Attr = message[idx]
			}else{
				request.Data = message[idx]
			}
		}
	}

	idx ++

	//todo: verify message
	if request.DataHead.DataLen == 0 {
		//fmt.Println("GetMessageHeader: invalid message!")
		return nil, errors.New("ParseRequestMsg: null message")
	}
	return &request, nil

}


/* *******************************************************************
   Function:     ParseRequestMsg()
   Description:  解析response消息，并保存消息头部，返回response的data
   Date:         2019.4.4
   Auther:       lvqiuxia
   Input/Output:
   Others:       4.8优化，处理err
********************************************************************** */
func ParseResponseMsg(message [][]byte) (*ResponseMsg, error) {
	response := ResponseMsg{}

	if message == nil {
		//fmt.Println("ParseResponseMsg: null message!")
		return nil, errors.New("ParseResponseMsg: null message")
	}
	frames := len(message)
	idx := 0

	//skip router header
	for _,temp := range message{
		if temp != nil && strings.HasPrefix(string(temp),"KMSG") == true{
			break
		}
		idx ++
	}

	//process message header
	mhs := strings.Split(string(message[idx]),"|")

	if len(mhs) <= 5{
		//fmt.Println("ParseResponseMsg: invalid message header!")
		return nil, errors.New("ParseResponseMsg: invalid message header len")
	}

	for _,mh := range mhs{
		if mh == ""{
			//fmt.Println("ParseResponseMsg: invalid message header!")
			return nil, errors.New("ParseResponseMsg: invalid message header who is nil")
		}
	}
	if mhs[0] != "KMSG" || mhs[1] != "REP"{
		//fmt.Println("ParseResponseMsg: invalid message header!")
		return nil, errors.New("ParseResponseMsg: invalid message header REP")
	}

	seqNo,_ := strconv.Atoi(mhs[5])
	response.MsgHead.HeaderType = mhs[0]
	response.MsgHead.MsgType = mhs[1]
	response.MsgHead.Method = mhs[2]
	response.MsgHead.Service = mhs[3]
	response.MsgHead.Status = mhs[4]
	response.MsgHead.SequenceNo = uint64(seqNo)

	//process attr or data header
	idx ++
	if idx < frames{
		adhs := strings.Split(string(message[idx]),"|")
		if len(adhs) < 2{
			//fmt.Println("ParseResponseMsg: invalid header!")
			return nil, errors.New("ParseResponseMsg: invalid message header DATA len")
		}
		if strings.Contains("ATTRDATA",string(adhs[0])) == false {
			//fmt.Println("ParseResponseMsg: invalid header!")
			return nil, errors.New("ParseResponseMsg: invalid message header DATA")
		}

		dLen,_ := strconv.Atoi(string(adhs[1]))
		if adhs[0] == "ATTR"{
			response.AttrHead.DataLen = uint64(dLen)
		}
		if adhs[0] == "DATA"{
			response.DataHead.DataLen = uint64(dLen)
		}
		if dLen > 0 {
			idx ++
			if idx >= frames || len(message[idx]) != dLen {
				//fmt.Println("ParseResponseMsg: invalid message data!")
				return nil, errors.New("ParseResponseMsg: invalid message data")
			}
			if adhs[0] == "ATTR"{
				response.Attr = message[idx]
			}else{
				response.Data = message[idx]
			}
		}
	}

	idx ++

	//todo: verify message
	if response.DataHead.DataLen == 0 {
		//fmt.Println("GetMessageHeader: invalid message!")
		return nil, errors.New("ParseResponseMsg: invalid response")
	}
	return &response, nil

}

