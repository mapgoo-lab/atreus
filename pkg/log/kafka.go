package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mapgoo-lab/atreus/pkg/conf/env"
	kafka "github.com/segmentio/kafka-go"
	"os"
	"strings"
	"time"
)

type K8SMetadata struct {
	ContainerName string `json:"container"`
	PodName       string `json:"pod"`
	Namespace     string `json:"namespace"`
	AppName       string `json:"app"`
	AppID         string `json:"appid"`
}

type KafkaLog struct {
	Level      Level        `json:"level"`
	Time       string       `json:"time"`
	Message    string       `json:"message"`
	Kubernetes *K8SMetadata `json:"kubernetes"`
}

type KafkaHandler struct {
	render      Render
	writer      *kafka.Writer
	k8SMetadata *K8SMetadata
	key         string
}

// 如果在K8S中运行，需要附加K8S元数据，以方便在日志中根据label查找
func getK8sMetadata() *K8SMetadata {
	containerName := os.Getenv("CONTAINER_NAME")
	podName := os.Getenv("POD_NAME")
	namespace := os.Getenv("POD_NAMESPACE")
	appName := os.Getenv("APP_NAME")

	if containerName != "" && podName != "" && namespace != "" && appName != "" {
		return &K8SMetadata{
			ContainerName: containerName,
			PodName:       podName,
			Namespace:     namespace,
			AppName:       appName,
			AppID:         env.AppID,
		}
	} else {
		return nil
	}
}

func NewKafka(brokers string, topic string) *KafkaHandler {
	//brokers是一个逗号分隔的字符串，包含Kafka代理的地址列表，现在分隔为一个字符串数组
	brokersList := strings.Split(brokers, ",")

	if len(brokersList) == 0 {
		return nil
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokersList...),
		Topic:    topic,
		Balancer: kafka.Murmur2Balancer{},
		Async:    true,
	}

	k8sMetadata := getK8sMetadata()
	//因为同一个服务实例产生的日志，必须保证时序，所以key必须是固定的，以保证同一个服务实例的日志写到固定分区
	key := ""
	if k8sMetadata != nil {
		//运行在K8S中，则使用PodName作为key
		key = k8sMetadata.PodName
	} else if env.RunContainer == "true" || env.RunContainer == "1" || env.RunContainer == "True" {
		//运行在容器中，但是没有运行在K8S中，则使用HOSTNAME环境变量作为key
		key = env.Hostname
	} else {
		//运行在物理机或者虚拟机中，则使用Hostname-APPID-timestamp作为key
		key = fmt.Sprintf("%s-%s-%d", env.Hostname, env.AppID, time.Now().UnixMilli())
	}

	return &KafkaHandler{
		render:      newPatternRender("%L %d-%T %f %M"),
		writer:      writer,
		k8SMetadata: k8sMetadata,
		key:         key,
	}
}

func (h *KafkaHandler) Log(ctx context.Context, lv Level, args ...D) {
	d := toMap(args...)
	// add extra fields
	addExtraField(ctx, d)
	now := time.Now()
	d[_time] = now.Format(_timeFormat)
	var w bytes.Buffer
	h.render.Render(&w, d)

	log := KafkaLog{
		Level:      lv,
		Time:       now.Format(_timeFormat),
		Message:    w.String(),
		Kubernetes: h.k8SMetadata,
	}

	if logData, err := json.Marshal(log); err == nil {
		if err := h.writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(h.key),
			Value: logData,
		}); err != nil {
			fmt.Println(err)
		}
	}
}

func (h *KafkaHandler) Close() error {
	return h.writer.Close()
}

func (h *KafkaHandler) SetFormat(format string) {
	h.render = newPatternRender(format)
}
