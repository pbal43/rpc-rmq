# rpc-rmq
Notifier. Based on RPC + Rabbit MQ

Описание:
RPC (Remote Procedure Call) с RabbitMQ:
Клиент отправляет число N в очередь "rpc_queue" и ждёт ответа.
Сервер слушает очередь "rpc_queue", получает число, вычисляет N * 2 и отправляет результат обратно клиенту.
