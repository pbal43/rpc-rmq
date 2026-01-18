package internal

import (
	"flag"
)

const (
	defRmqURL       = "amqp://admin:admin@localhost:5672/"
	defRpcQueueName = "rpc_queue"
)

type Config struct {
	RmqURL       string
	RpcQueueName string
}

func ReadConfig() Config {
	var config Config
	flag.StringVar(&config.RmqURL, "rmqurl", defRmqURL, "rmqurl")
	flag.StringVar(&config.RpcQueueName, "rpcqueuename", defRpcQueueName, "rpcqueuename")
	flag.Parse()

	return config
}
