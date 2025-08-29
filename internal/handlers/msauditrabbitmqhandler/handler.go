package msauditrabbitmqhandler

import (
	"context"
	"encoding/json"

	"github.com/NorskHelsenett/ror-ms-audit/internal/services/auditservice"
	"github.com/NorskHelsenett/ror/pkg/messagebuscontracts"
	"github.com/NorskHelsenett/ror/pkg/rlog"
	"github.com/rabbitmq/amqp091-go"
)

type auditmessagehandler struct {
}

func (amh auditmessagehandler) HandleMessage(ctx context.Context, message amqp091.Delivery) error {
	var event messagebuscontracts.AclUpdateEvent
	err := json.Unmarshal(message.Body, &event)
	if err != nil {
		rlog.Error("could not convert to json", err)
	}

	auditservice.CreateAndCommitAclList(ctx, event)
	return nil
}
