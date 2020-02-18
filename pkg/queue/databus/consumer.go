package databus

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"os"
	"time"
	"github.com/mapgoo-lab/atreus/pkg/log"
)

//使用者必须实现的接口
type ConsumerDeal interface {
	//数据处理的实现
	DealMessage(data []byte) error

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
	consumer sarama.ConsumerGroup
	deal ConsumerDeal
}

type ConsumerParam struct {
	Address []string
	GroupId string
	Topic string
	DealHanle ConsumerDeal
	KafkaVer string
}

func NewConsumer(param ConsumerParam) (ConsumerEvent, error) {
	config := sarama.NewConfig()

	//是否接收返回的错误消息,当发生错误时会放到Error这个通道中.从它里面获取错误消息
	config.Consumer.Return.Errors = true
	
	//指定Offset,也就是从哪里获取消息,默认时从主题的开始获取
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	
	//失败后再次尝试的间隔时间
	config.Consumer.Retry.Backoff = 2 * time.Second

	//最大等待时间
	config.Consumer.MaxWaitTime = 250 * time.Millisecond
	config.Consumer.MaxProcessingTime = 100 * time.Millisecond

	// 提交跟新Offset的频率
	config.Consumer.Offsets.CommitInterval = 1 * time.Second

	//消费组内的消费者消费的算法（BalanceStrategySticky、BalanceStrategyRange、BalanceStrategyRoundRobin）
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin;

	config.ClientID = fmt.Sprintf("Consumer-%d", os.Getpid())
	
    //设置使用的kafka版本,如果低于V0_10_0_0版本,消息中的timestrap没有作用.需要消费和生产同时配置
    //注意，版本设置不对的话，kafka会返回很奇怪的错误，并且无法成功发送消息
    version, err := sarama.ParseKafkaVersion(param.KafkaVer)
	if err != nil {
		log.Error("Error parsing Kafka version: %v", err)
		return nil, err
	}
	config.Version = version

    consumer, err := sarama.NewConsumerGroup(param.Address, param.GroupId, config)
    if err != nil {
		log.Error("Error sarama.NewConsumerGroup: %v", err)
        return nil, err
	}

	return &consumerEvent{
		address: param.Address,
		groupid: param.GroupId,
		topic: param.Topic,
		consumer: consumer,
		deal: param.DealHanle,
	}, nil
}

func (handle *consumerEvent) Start() error {
	ctx := context.Background()
	topics := []string{handle.topic};
	for {
		handler := consumerGroupHandler{handle}
		if err := handle.consumer.Consume(ctx, topics, handler); err != nil {
			log.Error("Error from consumer: %v", err)
			return err
		}
	}
}

func (handle *consumerEvent) Close() {
	log.Info("Close consumer")
	handle.consumer.Close()
}

type consumerGroupHandler struct {
	event *consumerEvent
}

func (handle consumerGroupHandler) Setup(sess sarama.ConsumerGroupSession) error {
	handle.event.deal.Setup(sess.Claims(), sess.MemberID(), sess.GenerationID())
	return nil 
}
func (handle consumerGroupHandler) Cleanup(sess sarama.ConsumerGroupSession) error {
	handle.event.deal.Cleanup(sess.Claims(), sess.MemberID(), sess.GenerationID())
	return nil 
}
func (handle consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		err := handle.event.deal.DealMessage(msg.Value)
		if err != nil {
			log.Info("Message topic:%q partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
