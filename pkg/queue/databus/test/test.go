package main

import (
	"log"
	"fmt"
    pb "./api"
	"github.com/mapgoo-lab/atreus/pkg/queue/databus"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/proto"
)

type ConsumerDealHandle struct {}

func (handle ConsumerDealHandle) DealMessage(data []byte,ctx context.Context) error {
	//解包proto
	var req *pb.EventReq = new(pb.EventReq)
	err := proto.Unmarshal(data, req)
	if err != nil {
		fmt.Printf("Unmarshal data error:%q\n", data)
		return err
	}

	if req.SubType == pb.EM_SUB_TYPE_EM_SUBTYPE_USERREG {
		//解包Any类型
		var temp *pb.UserRegEvent = new(pb.UserRegEvent)
		err = ptypes.UnmarshalAny(req.Details, temp)
		if err != nil {
			log.Printf("UnmarshalAny: %v", err)
		}
		fmt.Printf("EventReq: %v\n", req)
		fmt.Printf("UserRegEvent: %v\n", temp)
	}
	return nil 
}

func (handle ConsumerDealHandle) Setup(topicAndPartitions map[string][]int32, memberId string, generationId int32) {
	for topic, partitions := range topicAndPartitions {
		fmt.Printf("Setup topic:%q partition:%d\n", topic, partitions)
	}
	fmt.Printf("Setup MemberID:%q GenerationID:%d\n", memberId, generationId)
}

func (handle ConsumerDealHandle) Cleanup(topicAndPartitions map[string][]int32, memberId string, generationId int32) {
	for topic, partitions := range topicAndPartitions {
		fmt.Printf("Cleanup topic:%q partition:%d\n", topic, partitions)
	}
	fmt.Printf("Cleanup MemberID:%q GenerationID:%d\n", memberId, generationId)
}

func main()  {
	proparam := databus.ProducerParam {
		Address: []string{"192.168.100.203:9092","192.168.100.203:9093"},
		Topic: "test4",
		IsAck: true,
		KafkaVer: "0.11.0.0",
	}
	producer, err := databus.NewAsyncProducer(proparam)
	if err != nil {
		log.Panicf("NewAsyncProducer: %v", err)
	}
	defer producer.Close()

	var req *pb.EventReq = new(pb.EventReq)
	req.Sequence = 123456;
	req.Time = 232323;
	req.BussiType = pb.EM_BUSSI_TYPE_EM_BUSSITYPE_USER
	req.SubType = pb.EM_SUB_TYPE_EM_SUBTYPE_USERREG

	var userreg *pb.UserRegEvent = new(pb.UserRegEvent)
	userreg.RegisterType = 1
	userreg.UserId = 123

	//打包Any类型
	any, err := ptypes.MarshalAny(userreg)
	req.Details = any

	//打包请求
	data, err := proto.Marshal(req)
	if err != nil {
		log.Panicf("Marshal: %v", err)
	}

	err = producer.SendMessage(data, "123456")
	if err != nil {
		log.Panicf("SendMessage: %v", err)
	}

	fmt.Printf("-------------------------------------\n")

	//-------------------------------------
	conparam := databus.ConsumerParam {
		Address: []string{"192.168.100.203:9092","192.168.100.203:9093"},
		GroupId: "test4-group-1",
		Topic: "test4",
		DealHanle: ConsumerDealHandle{},
		KafkaVer: "0.11.0.0",
	}

	consumer, err := databus.NewConsumer(conparam)
	if err != nil {
		log.Panicf("NewConsumer: %v", err)
	}
	defer consumer.Close()

	consumer.Start()
}
