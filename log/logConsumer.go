package log

import (
	"encoding/json"
	"strconv"

	myLogger "github.com/livegoplayer/go_logger/logger"
	"github.com/sirupsen/logrus"
)

const (
	GO_FILE_STORE = iota + 1
	GO_USER
	GO_USER_RPC
)

var goUserMysqlLogger *logrus.Logger
var goFileStoreMysqlLogger *logrus.Logger
var goUserRpcMysqlLogger *logrus.Logger

func InitGoUserMysqlLogger(mysqlLogger *logrus.Logger) {
	goUserMysqlLogger = mysqlLogger
}

func InitGoUserRpcMysqlLogger(mysqlLogger *logrus.Logger) {
	goUserRpcMysqlLogger = mysqlLogger
}

func InitGoFileStoreMysqlLogger(mysqlLogger *logrus.Logger) {
	goFileStoreMysqlLogger = mysqlLogger
}

type LogMessage struct {
	Message string       `json:"message"`
	Level   logrus.Level `json:"level"`
	AppType int          `json:"app_type"`
}

//打印日志方法
func Callback(msg []byte) bool {
	logMessage := &LogMessage{}
	err := json.Unmarshal(msg, logMessage)
	if err != nil {
		myLogger.Error("json Unmarshal error:" + err.Error())
		return false
	}

	mysqlLogger := getMysqlLoggerByAppType(logMessage.AppType)
	if mysqlLogger == nil {
		message := "请先初始化mysql连接：app_type=" + strconv.Itoa(logMessage.AppType)
		myLogger.Error(message)
		return false
	}

	mysqlLogger.Log(logMessage.Level, logMessage.Message)

	return true
}

func getMysqlLoggerByAppType(appType int) (mysqlLogger *logrus.Logger) {
	switch appType {
	case GO_FILE_STORE:
		mysqlLogger = goFileStoreMysqlLogger
	case GO_USER:
		mysqlLogger = goUserMysqlLogger
	case GO_USER_RPC:
		mysqlLogger = goUserRpcMysqlLogger
	}

	return mysqlLogger
}
