package main

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

type Message struct {
	Message    []byte //json encode后的数据
	RetryTimes int    //重试次数
	Exchange   string
	RoutingKey string
}

var (
	amqpConnection *amqp.Connection
	amqpChannel    *amqp.Channel
	amqpUrl        string
	done           chan bool
)

//初始化channel
func InitMqChannel(url string) bool {
	var err error
	amqpConnection, err = amqp.Dial(url)
	if err != nil {
		panic(err)
	}

	amqpChannel, err = amqpConnection.Channel()
	if err != nil {
		panic(err)
	}

	if amqpUrl == "" {
		amqpUrl = url
	}

	return true
}

//获取channel
func GetChannel() *amqp.Channel {
	if amqpChannel != nil && !amqpConnection.IsClosed() {
		return amqpChannel
	}

	if amqpUrl != "" {
		InitMqChannel(amqpUrl)
	} else {
		panic("连接尚未初始化")
	}

	if amqpChannel != nil && amqpConnection.IsClosed() {
		panic("连接已经关闭")
	}

	return amqpChannel
}

//发布消息
func Publish(message *Message) {
	channel := GetChannel()

	//注入信息以便重新入队列
	msg := wrapMessage(message.Message, message.RetryTimes, message.Exchange, message.RoutingKey)

	err := channel.Publish(message.Exchange, message.RoutingKey, true, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        msg,
	})

	if err != nil {
		panic(err)
	}
}

type ConsumerCallBackFunc func(msg []byte) bool

//处理消息
func StartConsumer(queueName, consumeName string, callback ConsumerCallBackFunc) {
	channel := GetChannel()

	msgs, err := channel.Consume(queueName, consumeName, true, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	done := make(chan bool)

	for msg := range msgs {
		message := getMessage(msg.Body)
		success := callback(message.Message)
		if !success {
			//给他塞回去
			message.RetryTimes = message.RetryTimes + 1
			Publish(message)
		}
	}

	//可控阻塞
	<-done

	//consumer中关闭channel
	err = channel.Close()
	if err != nil {
		panic(err)
	}
}

func wrapMessage(realContext []byte, retryTimes int, Exchange string, routingKey string) []byte {
	message := Message{
		Message:    realContext,
		RetryTimes: retryTimes,
		Exchange:   Exchange,
		RoutingKey: routingKey,
	}

	msg, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	return msg
}

func getMessage(msg []byte) *Message {
	message := &Message{}
	err := json.Unmarshal(msg, message)
	if err != nil {
		panic(err)
	}

	return message
}

//关闭 consumer
func CloseConsumer() {
	done <- true
}
