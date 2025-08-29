package msauditrabbitmqhandler

import (
	"github.com/NorskHelsenett/ror-ms-audit/internal/msauditconnections"

	"github.com/NorskHelsenett/ror/pkg/handlers/rabbitmqhandler"

	"github.com/NorskHelsenett/ror/pkg/rlog"

	"github.com/rabbitmq/amqp091-go"
)

var (
	QueueName = "ms-audit"
)

func StartListening() {
	queueArgs := amqp091.Table{
		amqp091.QueueTypeArg: amqp091.QueueTypeQuorum,
	}

	go func() {
		config := rabbitmqhandler.RabbitMQListnerConfig{
			Client:    msauditconnections.RabbitMQConnection,
			QueueName: QueueName,
			Consumer:  "",
			AutoAck:   false,
			Exclusive: false,
			NoLocal:   false,
			NoWait:    false,
			Args:      queueArgs,
		}
		rabbithandler := rabbitmqhandler.New(config, auditmessagehandler{})
		err := msauditconnections.RabbitMQConnection.RegisterHandler(rabbithandler)
		if err != nil {
			rlog.Fatal("could not register handler", err)
		}
	}()
}
