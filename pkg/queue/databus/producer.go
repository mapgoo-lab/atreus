package databus

import (
	"time"
	"errors"
	"strconv"
	"github.com/Shopify/sarama"
	"github.com/mapgoo-lab/atreus/pkg/log"
)

type ProducerEvent interface {
	//发送消息接口
	SendMessage(data []byte, key string) error

	//关闭生产者
	Close()
}

const (
	//返回一个手动选择分区的分割器,也就是获取msg中指定的`partition`
    KafkaManual uint32 = 1
	
	//通过随机函数随机获取一个分区号
    KafkaRandom uint32 = 2
	
	//环形选择,也就是在所有分区中循环选择一个
    KafkaRoundRobin uint32 = 3
	
	//通过msg中的key生成hash值,选择分区
    KafkaHash uint32 = 4
)

type producerEvent struct {
	address []string
	topic string
	isack bool
	producer sarama.AsyncProducer
	partlen int
	partitioner uint32
}

type ProducerParam struct {
	Address []string
	Topic string
	IsAck bool
	KafkaVer string
	Partitioner uint32
}

func NewAsyncProducer(param ProducerParam) (ProducerEvent, error) {
	config := sarama.NewConfig()
	config.Net.MaxOpenRequests = 10
	config.Net.DialTimeout = 30 * time.Second
	config.Net.ReadTimeout = 30 * time.Second
	config.Net.WriteTimeout = 30 * time.Second

    //等待服务器所有副本都保存成功后的响应
	config.Producer.RequiredAcks = sarama.WaitForAll
	
	//分区选择算法
	if param.Partitioner == KafkaManual {
		config.Producer.Partitioner = sarama.NewManualPartitioner
	} else if param.Partitioner == KafkaRandom {
		config.Producer.Partitioner = sarama.NewRandomPartitioner
	} else if param.Partitioner == KafkaRandom {
		config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	} else {
		config.Producer.Partitioner = sarama.NewHashPartitioner
	}
    
	
    //是否等待成功和失败后的响应,只有上面的RequireAcks设置不是NoReponse这里才有用.
    if param.IsAck == true {
		config.Producer.Return.Successes = true
		config.Producer.Return.Errors = true
	}

	config.Producer.Timeout = 10 * time.Second

	//重试次数
	config.Producer.Retry.Max = 3 
	config.Producer.Retry.Backoff = 100 * time.Millisecond
	
    //设置使用的kafka版本,如果低于V0_10_0_0版本,消息中的timestrap没有作用.需要消费和生产同时配置
    //注意，版本设置不对的话，kafka会返回很奇怪的错误，并且无法成功发送消息
    version, err := sarama.ParseKafkaVersion(param.KafkaVer)
	if err != nil {
		log.Error("Error parsing Kafka version: %v", err)
		return nil, err
	}
	config.Version = version

	client, err := sarama.NewClient(param.Address, config)
	if err != nil {
		log.Error("Error sarama.NewClient: %v", err)
        return nil, err
	}
	
	partitions, err := client.Partitions(param.Topic)
	if err != nil {
		log.Error("Error client.Partitions: %v", err)
        return nil, err
	}
	partlen := len(partitions)
	client.Close()
	
    producer, err := sarama.NewAsyncProducer(param.Address, config)
    if err != nil {
		log.Error("Error sarama.NewAsyncProducer: %v", err)
        return nil, err
	}

	return &producerEvent{
		address: param.Address,
		topic: param.Topic,
		isack: param.IsAck,
		producer: producer,
		partlen: partlen,
		partitioner: param.Partitioner,
	}, nil
}

func (handle *producerEvent) SendMessage(data []byte, key string) error {
	var partindex int32
	partindex = 0
	if handle.partitioner == KafkaManual {
		index, err := strconv.Atoi(key)
		if err != nil {
			index = 0
		}
		partindex = int32(index % handle.partlen)
	}
	
	// 注意：这里的msg必须得是新构建的变量，不然你会发现发送过去的消息内容都是一样的，因为批次发送消息的关系。
	msg := &sarama.ProducerMessage{
		Topic: handle.topic,
		Key:sarama.ByteEncoder(key),
		Value:sarama.ByteEncoder(data),
		Partition:partindex,
	}

	//使用通道发送
	handle.producer.Input() <- msg

	if handle.isack == true {
		select {
		//case suc := <-handle.producer.Successes():
			//log.Info("offset: ", suc.Offset, "timestamp: ", suc.Timestamp.String(), "partitions: ", suc.Partition)
		case <-handle.producer.Successes():
			return nil
		case fail := <-handle.producer.Errors():
			log.Error("SendMessage fail: %v", fail.Err)
			return fail.Err
		case timeout := <-time.After(time.Second*10):
			log.Error("ack msg error %p.", timeout)
			return errors.New("ack msg timeout.")
		}
	}

	return nil
}

func (handle *producerEvent) Close() {
	log.Info("Close producer")
	handle.producer.AsyncClose()
}
