package databusc

import (
	"fmt"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"os"
	"strconv"
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
	//0:poll 1:channel
	ConsumerMode int
	//0:timer commitmsg 1:commitmsg 2:auto
	CommitMode int
	ThreadNum int
	QueueLen int
	SessionMs int
	PollMs int
}

type consumerEvent struct {
	isclose bool
	param *ConsumerParam
	config kafka.ConfigMap
	consumer *kafka.Consumer
	wg *sync.WaitGroup
	exit chan int
	queuelist []chan *kafka.Message
	sis *Consistent
}

func NewConsumer(param *ConsumerParam, appname string, Id int) (ConsumerEvent, error) {
	handle, err := NewConsumerHandle(param, appname, Id)
	return handle, err
}

func NewConsumerHandle(param *ConsumerParam, appname string, Id int) (*consumerEvent, error) {
	if appname == "" {
		appname = "rdkafka"
	}

	handle := new(consumerEvent)
	handle.isclose = false
	handle.param = param

	handle.config = make(kafka.ConfigMap)
	handle.config["bootstrap.servers"] = param.Address
	handle.config["broker.address.family"] = "v4"
	handle.config["group.id"] = param.GroupId
	handle.config["session.timeout.ms"] = 120000
	handle.config["max.poll.interval.ms"] = 600000
	handle.config["client.id"] = fmt.Sprintf("%s-%d-%d-%d", appname, time.Now().Unix(), os.Getpid(), Id)
	handle.config["auto.offset.reset"] = "latest"
	handle.config["enable.auto.commit"] = true
	handle.config["enable.auto.offset.store"] = true
	handle.config["socket.keepalive.enable"] = true
	if handle.param.ConsumerMode == 1 {
		handle.config["go.events.channel.enable"] = true
	}
	if handle.param.CommitMode == 2 {
		handle.config["enable.auto.commit"] = false
	}
	if handle.param.SessionMs > 0 {
		handle.config["session.timeout.ms"] = handle.param.SessionMs
	}
	if handle.param.PollMs > 0 {
		handle.config["max.poll.interval.ms"] = handle.param.PollMs
	}
	handle.config["enable.partition.eof"] = true

	consumer, err := kafka.NewConsumer(&handle.config)
	if err != nil {
		log.Error("NewConsumer error(topic:%s,err:%v).", param.Topic, err)
		return nil, err
	}

	handle.consumer = consumer
	handle.wg = new(sync.WaitGroup)
	handle.wg.Add(1)
	if handle.param.ConsumerMode == 1 {
		handle.exit = make(chan int, 1)
	}

	if handle.param.ThreadNum <= 0 {
		handle.param.ThreadNum = 51
	}

	if handle.param.QueueLen <= 0 {
		handle.param.QueueLen = 5
	}

	handle.sis = New()
	for i := 0; i < handle.param.ThreadNum; i++ {
		elt := fmt.Sprintf("%d", i)
		handle.sis.Add(elt)
	}

	handle.queuelist = make([]chan *kafka.Message, handle.param.ThreadNum)
	for i := 0; i < handle.param.ThreadNum; i++ {
		handle.queuelist[i] = make(chan *kafka.Message, handle.param.QueueLen)
		go func(index int) {
			for {
				msg, ok := <-handle.queuelist[index]
				if ok {
					handle.DealMessage(msg)
				} else {
					log.Error("deal chan is close(index:%d).", index)
					break
				}
			}
		}(i)
	}

	return handle, nil
}

func (handle *consumerEvent) SendToChannel(msg *kafka.Message, index int) {
	var mod int32 = 0
	iseffective := false
	key := string(msg.Key)
	if key != ""{
		modstr, err := handle.sis.Get(key)
		if err == nil {
			convstr, converr := strconv.Atoi(modstr)
			if converr == nil {
				iseffective = true
				mod = int32(convstr)
			}
		}
	}

	if iseffective == false {
		mod = msg.TopicPartition.Partition % int32(handle.param.ThreadNum)
	}

	handle.queuelist[mod] <- msg
	if handle.param.CommitMode == 0 {
		if index > 1000 {
			handle.consumer.CommitMessage(msg)
			index = 0
		}
		index++
	} else if handle.param.CommitMode == 1 {
		handle.consumer.CommitMessage(msg)
	}
}

func (handle *consumerEvent) Start() error {
	go func () {
		defer func() {
			if r := recover(); r != nil {
				log.Error("Start exception(r:%+v)", r)
			}
		}()

		for handle.isclose == false {
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
					handle.SendToChannel(e, index)
				case kafka.AssignedPartitions:
					log.Error("AssignedPartitions(e:%v,%+v).", e, e.Partitions)
				case kafka.RevokedPartitions:
					log.Error("RevokedPartitions(e:%v).", e)
				case kafka.PartitionEOF:
					log.Error("PartitionEOF(e:%v).", e)
				case kafka.Error:
					log.Error("consumer error(code:%v,e:%v).", e.Code(), e)
					if e.Code() == kafka.ErrAllBrokersDown {
						time.Sleep(time.Duration(1)*time.Second)
					}
				default:
					log.Error("Ignored consumer error(e:%v).", e)
				}
			} else {
				select {
				case <- handle.exit:
					break

				case ev := <-handle.consumer.Events():
					switch e := ev.(type) {
					case kafka.AssignedPartitions:
						log.Error("AssignedPartitions(e:%v,%+v).", e, e.Partitions)
						handle.consumer.Assign(e.Partitions)
					case kafka.RevokedPartitions:
						log.Error("RevokedPartitions(e:%v).", e)
						handle.consumer.Unassign()
					case *kafka.Message:
						handle.SendToChannel(e, index)
					case kafka.PartitionEOF:
						log.Error("PartitionEOF(e:%v).", e)
					case kafka.Error:
						log.Error("consumer error(e:%v).", e)
					default:
						log.Error("Ignored consumer error(e:%v).", e)
					}
				}
			}
		}

		if handle.param.CommitMode == 0 || handle.param.CommitMode == 1 {
			handle.consumer.Commit()
		}

		handle.wg.Done()
		log.Info("consumerEvent start is exit(topic:%s).", handle.param.Topic)
	}()

	return nil
}

func (handle *consumerEvent) Close() {
	handle.isclose = true
	if handle.param.ConsumerMode == 1 {
		handle.exit <- 1
	}
	log.Info("wait consumerEvent is close(topic:%s).", handle.param.Topic)
	handle.wg.Wait()
	handle.consumer.Close()
	for i := 0; i < handle.param.ThreadNum; i++ {
		for len(handle.queuelist[i]) > 0 {
			time.Sleep(10*time.Millisecond)
			log.Info("wait queuelist deal(topic:%s,i:%d,len:%d).", handle.param.Topic, i, len(handle.queuelist[i]))
		}
		close(handle.queuelist[i])
	}
	log.Info("consumerEvent is closed(topic:%s).", handle.param.Topic)
}

func (handle *consumerEvent) CommitMessage(msg *kafka.Message) error {
	_, err := handle.consumer.CommitMessage(msg)
	return err
}

func (handle *consumerEvent) DealMessage(msg *kafka.Message) error {
	if msg == nil {
		log.Error("DealMessage msg is nil")
		return nil
	}
	
	defer func() {
		if r := recover(); r != nil {
			log.Error("DealMessage exception(r:%+v,Partition:%d)", r, msg.TopicPartition.Partition)
		}
	}()

	err := handle.param.Dealhanle.DealMessage(msg)
	if err != nil {
		log.Error("DealMessage failed(partition:%d,err:%v)", msg.TopicPartition.Partition, err)
	}
	
	return err
}

