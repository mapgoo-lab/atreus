package databus

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"os"
	"sync"
	"time"
	"errors"
	"github.com/mapgoo-lab/atreus/pkg/log"
)

//使用者必须实现的接口
type ConsumerDeal interface {
	//数据处理的实现
	DealMessage(data []byte, topic string, partition int32, offset int64, groupid string) error

	//消费组增加消费者的消息通知
	Setup(topicAndPartitions map[string][]int32, memberId string, generationId int32)

	//消费组中消费者退出的消息通知
	Cleanup(topicAndPartitions map[string][]int32, memberId string, generationId int32)
}

type ConsumerEvent interface {
	//启动轮询消费数据
	Start() error

	//关闭消费者，必须调用
	Close()
}

type consumerEvent struct {
	address []string
	groupid string
	topic string
	dealhanle ConsumerDeal
	config *sarama.Config
	isclose bool
	wg *sync.WaitGroup
	consumer sarama.ConsumerGroup
}

type ConsumerParam struct {
	Address []string
	GroupId string
	Topic string
	KafkaVer string
	Dealhanle ConsumerDeal
}

func NewConsumer(param ConsumerParam) (ConsumerEvent, error) {
	config := sarama.NewConfig()

	//是否接收返回的错误消息,当发生错误时会放到Error这个通道中.从它里面获取错误消息
	config.Consumer.Return.Errors = true

	//指定Offset,也就是从哪里获取消息,默认时从主题的开始获取
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	//config.Consumer.Offsets.Initial = sarama.OffsetOldest

	//消费组内的消费者消费的算法（BalanceStrategySticky、BalanceStrategyRange、BalanceStrategyRoundRobin）
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin;

	config.ClientID = fmt.Sprintf("Consumer-%d", os.Getpid())

	//设置使用的kafka版本,如果低于V0_10_0_0版本,消息中的timestrap没有作用.需要消费和生产同时配置
	//注意，版本设置不对的话，kafka会返回很奇怪的错误，并且无法成功发送消息
	version, err := sarama.ParseKafkaVersion(param.KafkaVer)
	if err != nil {
		log.Error("Error parsing Kafka version(topic:%s,err:%v).", param.Topic, err)
		return nil, err
	}
	log.Info("parsing Kafka version(topic:%s,version:%v).", param.Topic, version)

	config.Version = version

	consumer, err := sarama.NewConsumerGroup(param.Address, param.GroupId, config)
	if err != nil {
		log.Error("call NewConsumerGroup failed(topic:%s,err:%v).", param.Topic, err)
		return nil, err
	}

	go func() {
		for err := range consumer.Errors() {
			log.Error("consumer error(topic:%s,err:%v).", param.Topic, err)
		}
		log.Error("consumer error exit(topic:%s).", param.Topic)
	}()

	wg := new(sync.WaitGroup)
	wg.Add(1)
	return &consumerEvent {
		address: param.Address,
		groupid: param.GroupId,
		topic: param.Topic,
		config: config,
		isclose: false,
		dealhanle: param.Dealhanle,
		consumer: consumer,
		wg: wg,
	}, nil
}

func (handle *consumerEvent) Start() error {
	ctx := context.Background()
	topics := []string{handle.topic};
	handler := consumerGroupHandler{handle}

	for {
		if handle.isclose == true {
			break
		}

		err := handle.consumer.Consume(ctx, topics, handler)
		if err != nil {
			log.Error("consumer failed(topic:%s,err:%v).", handle.topic, err)
		}

		if ctx.Err() != nil {
			time.Sleep(time.Duration(1)*time.Second)
			ctx = context.Background()
			log.Error("create new ctx(topic:%s).", handle.topic)
		}
	}

	handle.wg.Done()
	log.Info("consumerEvent start is exit(topic:%s).", handle.topic)
	
	return nil
}

func (handle *consumerEvent) Close() {
	handle.isclose = true
	log.Info("wait consumerEvent is close(topic:%s).", handle.topic)
	handle.wg.Wait()
	handle.consumer.Close()
	log.Info("consumerEvent is closed(topic:%s).", handle.topic)
}

type consumerGroupHandler struct {
	event *consumerEvent
}

func (handle consumerGroupHandler) Setup(sess sarama.ConsumerGroupSession) error {
	handle.event.dealhanle.Setup(sess.Claims(), sess.MemberID(), sess.GenerationID())
	return nil
}
func (handle consumerGroupHandler) Cleanup(sess sarama.ConsumerGroupSession) error {
	handle.event.dealhanle.Cleanup(sess.Claims(), sess.MemberID(), sess.GenerationID())
	return nil
}
func (handle consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		err := handle.event.dealhanle.DealMessage(msg.Value, msg.Topic, msg.Partition, msg.Offset, handle.event.groupid)
		if err != nil {
			log.Error("Message failed(topic:%s,partition:%d,offset:%d,err:%v)", msg.Topic, msg.Partition, msg.Offset, err)
		}
		sess.MarkMessage(msg, "")
	}

	return nil
}
