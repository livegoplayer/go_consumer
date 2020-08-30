package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"

	myLogger "github.com/livegoplayer/go_logger"

	"github.com/livegoplayer/go_consumer/log"

	"github.com/oklog/oklog/pkg/group"

	mqHelper "github.com/livegoplayer/go_mq_helper/rabbitmq"

	. "github.com/livegoplayer/go_helper"
)

func main() {
	//加载.env文件
	LoadEnv()

	//如果是debug模式的话，直接打印到控制台
	var appLogger *logrus.Logger
	if gin.IsDebugging() {
		appLogger = myLogger.GetConsoleLogger()
	} else {
		appLogger = myLogger.GetMysqlLogger(viper.GetString("log.app_log_mysql_host"), viper.GetString("log.app_log_mysql_port"), viper.GetString("log.app_log_mysql_db_name"), viper.GetString("log.app_log_mysql_table_name"), viper.GetString("log.app_log_mysql_user"), viper.GetString("log.app_log_mysql_pass"))
	}
	myLogger.SetLogger(appLogger)

	//initlogers
	mysqlGoUserRpcLogger := myLogger.GetMysqlLogger(viper.GetString("amqp.go_user_rpc_log_consumer.app_log_mysql_host"), viper.GetString("amqp.go_user_rpc_log_consumer.app_log_mysql_port"), viper.GetString("amqp.go_user_rpc_log_consumer.app_log_mysql_db_name"), viper.GetString("amqp.go_user_rpc_log_consumer.app_log_mysql_table_name"), viper.GetString("amqp.go_user_rpc_log_consumer.app_log_mysql_user"), viper.GetString("amqp.go_user_rpc_log_consumer.app_log_mysql_pass"))
	mysqlGoUserLogger := myLogger.GetMysqlLogger(viper.GetString("amqp.go_user_log_consumer.app_log_mysql_host"), viper.GetString("amqp.go_user_log_consumer.app_log_mysql_port"), viper.GetString("amqp.go_user_log_consumer.app_log_mysql_db_name"), viper.GetString("amqp.go_user_log_consumer.app_log_mysql_table_name"), viper.GetString("amqp.go_user_log_consumer.app_log_mysql_user"), viper.GetString("amqp.go_user_log_consumer.app_log_mysql_pass"))
	mysqlGoFileStoreLogger := myLogger.GetMysqlLogger(viper.GetString("amqp.go_file_store_log_consumer.app_log_mysql_host"), viper.GetString("amqp.go_file_store_log_consumer.app_log_mysql_port"), viper.GetString("amqp.go_file_store_log_consumer.app_log_mysql_db_name"), viper.GetString("amqp.go_file_store_log_consumer.app_log_mysql_table_name"), viper.GetString("amqp.go_file_store_log_consumer.app_log_mysql_user"), viper.GetString("amqp.go_file_store_log_consumer.app_log_mysql_pass"))
	log.InitGoFileStoreMysqlLogger(mysqlGoFileStoreLogger)
	log.InitGoUserMysqlLogger(mysqlGoUserLogger)
	log.InitGoUserRpcMysqlLogger(mysqlGoUserRpcLogger)

	//初始化mq
	mqHelper.InitMqChannel(viper.GetString("amqp.url"))

	g := &group.Group{}

	initConsumer(g)

	err := g.Run()
	if err != nil {
		myLogger.Error(err)
	}
}

func initConsumer(g *group.Group) {
	var channel1, channel2, channel3 *amqp.Channel
	//go_user_log_consumer
	g.Add(func() error {
		channel1 = mqHelper.GetNewChannel()
		fmt.Printf("go_user_log_consumer start...\n")
		err := mqHelper.AddConsumer(viper.GetString("amqp.go_user_log_consumer.queue_name"), viper.GetString("amqp.go_user_log_consumer.queue_name"), log.Callback, channel1)
		return err
	}, func(err error) {
		myLogger.Error("go_user_log_consumer" + err.Error())
		err = channel1.Close()
		if err != nil {
			myLogger.Error("go_user_log_consumer close channel err :" + err.Error())
		}
	})

	//go_file_store_log_consumer
	g.Add(func() error {
		channel2 = mqHelper.GetNewChannel()
		fmt.Printf("go_file_store_log_consumer start...\n")
		err := mqHelper.AddConsumer(viper.GetString("amqp.go_file_store_log_consumer.queue_name"), viper.GetString("amqp.go_file_store_log_consumer.queue_name"), log.Callback, channel2)
		return err
	}, func(err error) {
		myLogger.Error("go_file_store_log_consumer" + err.Error())
		err = channel2.Close()
		if err != nil {
			myLogger.Error("go_file_store_log_consumer close channel err :" + err.Error())
		}
	})

	//go_user_rpc_log_consumer
	g.Add(func() error {
		channel3 = mqHelper.GetNewChannel()
		fmt.Printf("go_user_rpc_log_consumer start...\n")

		err := mqHelper.AddConsumer(viper.GetString("amqp.go_user_rpc_log_consumer.queue_name"), viper.GetString("amqp.go_file_store_log_consumer.queue_name"), log.Callback, channel3)

		return err
	}, func(err error) {
		myLogger.Error("go_user_rpc_log_consumer" + err.Error())
		err = channel3.Close()
		if err != nil {
			myLogger.Error("go_user_rpc_log_consumer close channel err :" + err.Error())
		}
	})
}
