package msauditrabbitmqhandler

import (
	"github.com/NorskHelsenett/ror-ms-audit/internal/auditconfig"
	"github.com/spf13/viper"

	"github.com/NorskHelsenett/ror/pkg/handlers/rabbitmqhandler"

	"github.com/NorskHelsenett/ror/pkg/rlog"

	"github.com/rabbitmq/amqp091-go"
)

func StartListening() {
	queueArgs := amqp091.Table{
		amqp091.QueueTypeArg: amqp091.QueueTypeQuorum,
	}

	queueName := viper.GetString("RABBITMQ_QUEUE_NAME")

	go func() {
		config := rabbitmqhandler.RabbitMQListnerConfig{
			Client:    auditconfig.RabbitMQConnection,
			QueueName: queueName,
			Consumer:  "",
			AutoAck:   false,
			Exclusive: false,
			NoLocal:   false,
			NoWait:    false,
			Args:      queueArgs,
		}
		rabbithandler := rabbitmqhandler.New(config, auditmessagehandler{})
		err := auditconfig.RabbitMQConnection.RegisterHandler(rabbithandler)
		if err != nil {
			rlog.Fatal("could not register handler", err)
		}
	}()
}
