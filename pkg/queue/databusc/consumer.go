package databusc

import (
	"fmt"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"os"
	"time"
	"github.com/mapgoo-lab/atreus/pkg/log"
	"sync"
)

//使用者必须实现的接口
type ConsumerDeal interface {
	//数据处理的实现
	DealMessage(msg *kafka.Message) error
}

type ConsumerEvent interface {
	//启动轮询消费数据
	Start() error

	//关闭消费者，必须调用
	Close()

	//提交offset
	CommitMessage(msg *kafka.Message) error
}

type ConsumerParam struct {
	Address string
	GroupId string
	Topic string
	Dealhanle ConsumerDeal
	ConsumerMode int
	AutoCommit int
}

type consumerEvent struct {
	isclose bool
	param *ConsumerParam
	config kafka.ConfigMap
	consumer *kafka.Consumer
	wg *sync.WaitGroup
}

func NewConsumer(param *ConsumerParam) (ConsumerEvent, error) {
	handle := new(consumerEvent)
	handle.isclose = false
	handle.param = param

	handle.config = make(kafka.ConfigMap)
	handle.config["bootstrap.servers"] = param.Address
	handle.config["broker.address.family"] = "v4"
	handle.config["group.id"] = param.GroupId
	handle.config["session.timeout.ms"] = 6000
	handle.config["client.id"] = fmt.Sprintf("rdkafka-%d-%d", time.Now().Unix(), os.Getpid())
	handle.config["auto.offset.reset"] = "latest"
	handle.config["auto.commit.enable"] = true
	handle.config["enable.auto.offset.store"] = true
	handle.config["socket.keepalive.enable"] = true
	
	consumer, err := kafka.NewConsumer(&handle.config)
	if err != nil {
		log.Error("NewConsumer error(topic:%s,err:%v).", param.Topic, err)
		return nil, err
	}
	
	handle.consumer = consumer
	handle.wg = new(sync.WaitGroup)
	handle.wg.Add(1)
	
	return handle, nil
}

func (handle *consumerEvent) Start() error {
	for {
		err := handle.consumer.SubscribeTopics([]string{handle.param.Topic}, nil)
		if err != nil {
			log.Error("SubscribeTopics error(err:%v,topic:%v).", err, handle.param.Topic)
			time.Sleep(time.Duration(1)*time.Second)
		} else {
			break
		}
	}

	index := 0
	for handle.isclose == false {
		if handle.param.ConsumerMode == 0 {
			ev := handle.consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				handle.param.Dealhanle.DealMessage(e)
				if handle.param.AutoCommit == 1 {
					if index > 1000 {
						handle.consumer.Commit()
						index = 0
					}
					index++
				}
			case kafka.Error:
				log.Error("consumer error(code:%v,e:%v).", e.Code(), e)
				if e.Code() == kafka.ErrAllBrokersDown {
					time.Sleep(time.Duration(1)*time.Second)
				} 
			default:
				log.Error("Ignored consumer error(e:%v).", e)
			}
		} else {
			 ev := <-handle.consumer.Events()
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				log.Error("AssignedPartitions(e:%v).", e)
				handle.consumer.Assign(e.Partitions)
			case kafka.RevokedPartitions:
				log.Error("RevokedPartitions(e:%v).", e)
				handle.consumer.Unassign()
			case *kafka.Message:
				handle.param.Dealhanle.DealMessage(e)
				if handle.param.AutoCommit == 1 {
					if index > 1000 {
						handle.consumer.Commit()
						index = 0
					}
					index++
				}
			case kafka.PartitionEOF:
				log.Error("PartitionEOF(e:%v).", e)
			case kafka.Error:
				log.Error("consumer error(e:%v).", e)
			default:
				log.Error("Ignored consumer error(e:%v).", e)
			}
		}
	}
	handle.wg.Done()
	log.Info("consumerEvent start is exit(topic:%s).", handle.param.Topic)

	return nil
}

func (handle *consumerEvent) Close() {
	handle.isclose = true
	log.Info("wait consumerEvent is close(topic:%s).", handle.param.Topic)
	handle.wg.Wait()
	handle.consumer.Commit()
	handle.consumer.Close()
	log.Info("consumerEvent is closed(topic:%s).", handle.param.Topic)
}

func (handle *consumerEvent) CommitMessage(msg *kafka.Message) error {
	_, err := handle.consumer.CommitMessage(msg)
	return err
}
