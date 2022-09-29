package databusc

import (
	"github.com/mapgoo-lab/atreus/pkg/log"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"time"
)

type ProducerEvent interface {
	//发送消息接口
	SendMessage(data []byte, key string) error

	//发送消息到指定分区接口
	SendMessagePartition(data []byte, partition uint32) error

	//发送消息到分区取模接口
	SendMessageByMod(data []byte, key uint32) error

	//关闭生产者
	Close()
}

type ProducerParam struct {
	Address  string
	Topic    string
	IsAck    bool
	KafkaVer string
	//0:channel 1:sync
	ConsumerMode int
}

type producerEvent struct {
	isclose      bool
	param        *ProducerParam
	config       kafka.ConfigMap
	producer     *kafka.Producer
	maxpartition uint32
}

func NewAsyncProducer(param *ProducerParam) (ProducerEvent, error) {
	handle := new(producerEvent)
	handle.isclose = false
	handle.param = param

	handle.config = make(kafka.ConfigMap)
	handle.config["bootstrap.servers"] = param.Address
	handle.config["partitioner"] = "consistent_random"
	handle.config["socket.keepalive.enable"] = true
	handle.config["go.delivery.reports"] = false
	handle.config["request.required.acks"] = 1
	handle.config["acks"] = 1

	producer, err := kafka.NewProducer(&handle.config)
	if err != nil {
		log.Error("NewAsyncProducer error(topic:%s,err:%v).", param.Topic, err)
		return nil, err
	}

	handle.producer = producer

	medaresp, medaerr := handle.producer.GetMetadata(&param.Topic, false, 300)
	if medaerr != nil {
		log.Error("NewAsyncProducer error(topic:%s,medaerr:%v).", param.Topic, medaerr)
		return nil, medaerr
	}

	handle.maxpartition = uint32(len(medaresp.Topics[param.Topic].Partitions))

	log.Info("NewAsyncProducer(topic:%s,maxpartition:%d).", param.Topic, handle.maxpartition)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("NewAsyncProducer exception(r:%+v)", r)
			}
		}()

		for e := range handle.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				m := ev
				if m.TopicPartition.Error != nil {
					log.Error("Delivery failed(err:%v,Key:%s).", m.TopicPartition.Error, string(m.Key))
				} else {
					log.Error("Delivery report(err:%v,Key:%s).", m.TopicPartition.Error, string(m.Key))
				}

			default:
				log.Error("Ignored event(err:%v).", ev)
			}
		}
	}()

	return handle, nil
}

func (handle *producerEvent) SendMessage(data []byte, key string) error {
	return handle.transMessage(data, key, -1)
}

func (handle *producerEvent) SendMessagePartition(data []byte, partition uint32) error {
	return handle.transMessage(data, "", int32(partition))
}

func (handle *producerEvent) SendMessageByMod(data []byte, key uint32) error {
	var partition int32 = int32(key % handle.maxpartition)
	return handle.transMessage(data, "", partition)
}

func (handle *producerEvent) transMessage(data []byte, key string, partition int32) error {
	message := new(kafka.Message)
	message.TopicPartition.Topic = &handle.param.Topic
	message.TopicPartition.Partition = kafka.PartitionAny
	if partition >= 0 {
		message.TopicPartition.Partition = partition
	}
	message.Key = []byte(key)
	message.Value = data
	message.Timestamp = time.Now()

	go func(msg *kafka.Message) {
		if handle.param.ConsumerMode == 0 {
			handle.producer.ProduceChannel() <- msg
		} else {
			i := 0
			for i < 3 {
				err := handle.producer.Produce(msg, nil)
				if err != nil {
					if err.Error() == kafka.ErrQueueFull.String() {
						log.Error("transMessage ErrQueueFull(topic:%s,err:%v).", handle.param.Topic, err)
						handle.producer.Flush(100)
						continue
					} else {
						log.Error("transMessage error(topic:%s,err:%v).", handle.param.Topic, err)
						i++
						continue
					}
				} else {
					break
				}
			}
		}
	}(message)

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
