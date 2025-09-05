package main

import (
	"github.com/NorskHelsenett/ror-ms-audit/internal/auditconfig"
	"github.com/NorskHelsenett/ror-ms-audit/internal/handlers/msauditrabbitmqhandler"

	"github.com/NorskHelsenett/ror/pkg/services/rorservice"
)

func main() {
	//TODO: mshelper
	// config load interface
	rorservice.MustInit()

	auditconfig.Load()

	msauditrabbitmqhandler.StartListening()

	rorservice.Wait()
}
