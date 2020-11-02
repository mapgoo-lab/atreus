package databusc

import (
	"time"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"github.com/mapgoo-lab/atreus/pkg/log"
)

type ProducerEvent interface {
	//发送消息接口
	SendMessage(data []byte, key string) error

	//关闭生产者
	Close()
}

type ProducerParam struct {
	Address string
	Topic string
	IsAck bool
	KafkaVer string
}

type producerEvent struct {
	isclose bool
	param *ProducerParam
	config kafka.ConfigMap
	producer *kafka.Producer
}

func NewAsyncProducer(param *ProducerParam) (ProducerEvent, error) {
	handle := new(producerEvent)
	handle.isclose = false
	handle.param = param

	handle.config = make(kafka.ConfigMap)
	handle.config["bootstrap.servers"] = param.Address
	handle.config["partitioner"] = "consistent_random"
	handle.config["socket.keepalive.enable"] = true
	handle.config["produce.offset.report"] = false
	
	producer, err := kafka.NewProducer(&handle.config)
	if err != nil {
		log.Error("NewAsyncProducer error(topic:%s,err:%v).", param.Topic, err)
		return nil, err
	}
	
	handle.producer = producer

	go func() {
		for e := range handle.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				m := ev
				if m.TopicPartition.Error != nil {
					log.Error("Delivery failed(err:%v,Key:%s).", m.TopicPartition.Error, string(m.Key))
				}

			default:
				log.Error("Ignored event(err:%v).", ev)
			}
		}
	}()

	return handle, nil
}

func (handle *producerEvent) SendMessage(data []byte, key string) error {
	message := new(kafka.Message)
	message.TopicPartition.Topic = &handle.param.Topic
	message.TopicPartition.Partition = kafka.PartitionAny
	message.Key = []byte(key)
	message.Value = data
	message.Timestamp = time.Now()

	for {
		err := handle.producer.Produce(message, nil)
		if err != nil {
			if err.Error() == kafka.ErrQueueFull.String() {
				log.Error("SendMessage ErrQueueFull(topic:%s,err:%v).", handle.param.Topic, err)
				handle.producer.Flush(100)
				continue
			} else {
				log.Error("SendMessage error(topic:%s,err:%v).", handle.param.Topic, err)
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (handle *producerEvent) Close() {
	for {
		num := handle.producer.Flush(1000)
		if num == 0 {
			break
		}
		log.Info("wait Close(num:%d).", num)
	}
	log.Info("Close producer")
	handle.producer.Close()
}
