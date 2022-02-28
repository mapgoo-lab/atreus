package mqttclient

import (
	"errors"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mapgoo-lab/atreus/pkg/log"
	"strings"
	"sync"
	"time"
)

//使用者必须实现的接口
type MqttConsumerHandle interface {
	//接收消息
	SubMessage(clientId string, topic string, messageId uint16, data []byte) error
}

//使用者必须实现的接口
type MqttEventHandle interface {
	//连接事件
	ConnectEvent(clientId string) error

	//断开连接事件
	DisConnectEvent(clientId string, err error) error

	//重连事件
	ReconnectEvent(clientId string) error
}

//最好不要直接使用这个结构，通过NewMqttParam得到有默认值
type MqttParam struct {
	//连接地址，比如：tcp://foobar.com:1883
	Server string

	//客户端id
	ClientId string

	//用户名
	Username string

	//密码
	Password string

	//协议版本，3：3.1 4：3.1.1 5：5.0
	ProtocolVersion uint32

	//保持活跃的时间
	KeepAlive uint32

	//发送心跳时间
	PingTimeout uint32

	//连接超时
	ConnectTimeout uint32

	//最大重连间隔
	MaxReconnectInterval uint32

	//重连间隔
	ConnectRetryInterval uint32

	//清理会话
	CleanSession bool

	//自动重连
	AutoReconnect bool
}

func NewMqttParam() *MqttParam {
	param := new(MqttParam)
	param.Server = ""
	param.ClientId = ""
	param.Username = ""
	param.Password = ""
	param.ProtocolVersion = 5
	param.KeepAlive = 30
	param.PingTimeout = 10
	param.ConnectTimeout = 30
	param.MaxReconnectInterval = 10
	param.ConnectRetryInterval = 30
	param.CleanSession = false
	param.AutoReconnect = true
	return param
}

//返回的接口
type MqttClientHandle interface {
	//订阅消息
	Subscribe(topic string, qos byte, handle MqttConsumerHandle) error

	//发布消息
	Publish(topic string, qos byte, retained bool, payload interface{}) error

	//取消订阅
	Unsubscribe(topic string) error

	//断开连接
	Disconnect(quiesce uint)
}

type mqttClientHandle struct {
	//消费消息回调函数(key:topic,value:topic对应的回调函数)
	HandleList     map[string]MqttConsumerHandle
	TopicSplitList map[string][]string

	//事件回调函数
	EventHandle MqttEventHandle

	//锁
	Lock *sync.RWMutex

	//mqtt连接客户端
	MqttClient mqtt.Client

	//mqtt连接回调函数
	ConnectFunc mqtt.OnConnectHandler

	//mqtt连接断开回调函数
	ConnectLostFunc mqtt.ConnectionLostHandler

	//mqtt重连函数
	ReConnectFunc mqtt.ReconnectHandler

	//mqtt消息回调函数
	MessageSubFunc mqtt.MessageHandler

	//mqtt连接客户端配置参数
	Opts *mqtt.ClientOptions
}

func NewMqttClient(param *MqttParam, handle MqttEventHandle) (MqttClientHandle, error) {
	client := new(mqttClientHandle)
	client.Lock = new(sync.RWMutex)
	client.HandleList = make(map[string]MqttConsumerHandle)
	client.TopicSplitList = make(map[string][]string)
	client.EventHandle = handle
	client.ConnectFunc = client.ConnectHandler
	client.ConnectLostFunc = client.ConnectLostHandler
	client.ReConnectFunc = client.ReConnectHandler
	client.MessageSubFunc = client.MessageSubHandler
	client.Opts = mqtt.NewClientOptions()
	client.Opts.AddBroker(param.Server)
	client.Opts.SetProtocolVersion(uint(param.ProtocolVersion))
	client.Opts.SetClientID(param.ClientId)
	client.Opts.SetUsername(param.Username)
	client.Opts.SetPassword(param.Password)
	client.Opts.SetKeepAlive(time.Duration(param.KeepAlive) * time.Second)
	client.Opts.SetPingTimeout(time.Duration(param.PingTimeout) * time.Second)
	client.Opts.SetConnectTimeout(time.Duration(param.ConnectTimeout) * time.Second)
	client.Opts.SetMaxReconnectInterval(time.Duration(param.MaxReconnectInterval) * time.Second)
	client.Opts.SetConnectRetryInterval(time.Duration(param.ConnectRetryInterval) * time.Second)
	client.Opts.SetCleanSession(param.CleanSession)
	client.Opts.SetAutoReconnect(param.AutoReconnect)
	client.Opts.SetDefaultPublishHandler(client.MessageSubFunc)
	client.Opts.SetOnConnectHandler(client.ConnectFunc)
	client.Opts.SetConnectionLostHandler(client.ConnectLostFunc)
	client.Opts.SetReconnectingHandler(client.ReConnectFunc)

	client.MqttClient = mqtt.NewClient(client.Opts)
	token := client.MqttClient.Connect()
	if token.Wait() && token.Error() != nil {
		log.Error("NewMqttClient failed(err:%s,opts:%+v).", token.Error(), client.Opts)
		return client, token.Error()
	}

	log.Info("NewMqttClient success(Server:%s,ClientId:%s).", param.Server, param.ClientId)
	return client, nil
}

//订阅消息
func (d *mqttClientHandle) Subscribe(topic string, qos byte, handle MqttConsumerHandle) error {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Subscribe exception(r:%+v)", r)
		}
	}()
		
	if topic == "" {
		log.Error("Subscribe topic is empty.")
		return errors.New("topic is empty.")
	}

	if qos <= 0 || qos > 2 {
		log.Error("Subscribe qos is invaild(qos:%d).", qos)
		return errors.New("qos is invaild.")
	}

	if handle == nil {
		log.Error("Subscribe handle is empty.")
		return errors.New("handle is empty.")
	}

	topickey := ""
	topiclist := strings.Split(topic, "/")
	topiclen := len(topiclist)
	if topiclen > 2 && topiclist[0] == "$share" {
		topiclist = topiclist[2:]
		topickey = "/" + strings.Join(topiclist, "/")
	} else {
		topickey = topic
	}

	d.Lock.Lock()
	defer d.Lock.Unlock()
	if _, ok := d.HandleList[topickey]; !ok {
		d.HandleList[topickey] = handle
		d.TopicSplitList[topickey] = topiclist
	}

	token := d.MqttClient.Subscribe(topic, qos, nil)
	if token.Wait() && token.Error() != nil {
		log.Error("Subscribe failed(err:%s).", token.Error())
		return token.Error()
	}

	log.Info("Subscribe success(topic:%s).", topic)

	return nil
}

//发布消息
func (d *mqttClientHandle) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Publish exception(r:%+v)", r)
		}
	}()
	
	if topic == "" {
		log.Error("Publish topic is empty(payload:%v).", payload)
		return errors.New("topic is empty.")
	}

	if qos <= 0 || qos > 2 {
		log.Error("Publish qos is invaild(qos:%d,payload:%v).", qos, payload)
		return errors.New("qos is invaild.")
	}

	token := d.MqttClient.Publish(topic, qos, retained, payload)
	if token.Wait() && token.Error() != nil {
		log.Error("Publish failed(err:%s,payload:%v).", token.Error(), payload)
		return token.Error()
	}

	return nil
}

//取消订阅
func (d *mqttClientHandle) Unsubscribe(topic string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Unsubscribe exception(r:%+v)", r)
		}
	}()
	
	if topic == "" {
		log.Error("Unsubscribe topic is empty.")
		return errors.New("topic is empty.")
	}

	token := d.MqttClient.Unsubscribe(topic)
	if token.Wait() && token.Error() != nil {
		log.Error("Unsubscribe failed(err:%s).", token.Error())
		return token.Error()
	}

	d.Lock.Lock()
	defer d.Lock.Unlock()
	if _, ok := d.HandleList[topic]; !ok {
		delete(d.HandleList, topic)
	}

	log.Info("Unsubscribe success(topic:%s).", topic)
	return nil
}

//断开连接
func (d *mqttClientHandle) Disconnect(quiesce uint) {
	d.MqttClient.Disconnect(quiesce)
}

func (d *mqttClientHandle) ConnectHandler(client mqtt.Client) {
	reader := client.OptionsReader()

	err := d.EventHandle.ConnectEvent(reader.ClientID())
	if err != nil {
		log.Error("ConnectEvent return failed(ClientID:%+v,err:%v).", reader.ClientID(), err)
		return
	}

	return
}

func (d *mqttClientHandle) ConnectLostHandler(client mqtt.Client, err error) {
	reader := client.OptionsReader()

	callerr := d.EventHandle.DisConnectEvent(reader.ClientID(), err)
	if callerr != nil {
		log.Error("DisConnectEvent return failed(ClientID:%+v,callerr:%v).", reader.ClientID(), callerr)
		return
	}

	return
}

func (d *mqttClientHandle) ReConnectHandler(client mqtt.Client, opt *mqtt.ClientOptions) {
	reader := client.OptionsReader()

	err := d.EventHandle.ReconnectEvent(reader.ClientID())
	if err != nil {
		log.Error("ReconnectEvent return failed(ClientID:%+v,err:%v).", reader.ClientID(), err)
		return
	}

	return
}

func (d *mqttClientHandle) isSameTopic(topic string) (bool, string) {
	topiclist := strings.Split(topic, "/")
	topiclen := len(topiclist)

	for key, value := range d.TopicSplitList {
		subtopiclen := len(value)
		for i, topicsplit := range value {
			if i < topiclen {
				if topicsplit != topiclist[i] && topicsplit == "#" && topicsplit == "+" {
					break
				} else if topicsplit == "#" {
					return true, key
				}

				if (i == (topiclen - 1)) && (i == (subtopiclen - 1)) {
					return true, key
				}
			}
		}
	}

	return false, ""
}

func (d *mqttClientHandle) getMsgHandle(topic string) (MqttConsumerHandle, error) {
	d.Lock.RLock()
	defer d.Lock.RUnlock()
	isexist, topickey := d.isSameTopic(topic)
	if isexist == false {
		log.Error("getMsgHandle return failed(topic:%s).", topic)
		return nil, errors.New("订阅已取消")
	}

	return d.HandleList[topickey], nil
}

func (d *mqttClientHandle) MessageSubHandler(client mqtt.Client, msg mqtt.Message) {
	reader := client.OptionsReader()

	topic := msg.Topic()
	handle, err := d.getMsgHandle(topic)
	if err != nil {
		log.Error("SubMessageEvent return failed(ClientID:%+v,Payload:%s,Topic:%s,Duplicate:%v,Qos:%v,Retained:%v,MessageID:%d).", reader.ClientID(), msg.Payload(), msg.Topic(), msg.Duplicate(), msg.Qos(), msg.Retained(), msg.MessageID())
		return
	}

	err = handle.SubMessage(reader.ClientID(), msg.Topic(), msg.MessageID(), msg.Payload())
	if err != nil {
		log.Error("SubMessage return failed(ClientID:%+v,Payload:%s,Topic:%s,Duplicate:%v,Qos:%v,Retained:%v,MessageID:%d,err:%v).", client.OptionsReader(), msg.Payload(), msg.Topic(), msg.Duplicate(), msg.Qos(), msg.Retained(), msg.MessageID(), err)
		return
	}

	msg.Ack()
	return
}
