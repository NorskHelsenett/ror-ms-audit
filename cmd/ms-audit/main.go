package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/NorskHelsenett/ror-ms-audit/internal/auditconfig"
	"github.com/NorskHelsenett/ror-ms-audit/internal/handlers/msauditrabbitmqhandler"
	"github.com/NorskHelsenett/ror-ms-audit/internal/msauditconnections"

	"syscall"

	"github.com/NorskHelsenett/ror/pkg/config/configconsts"

	"github.com/NorskHelsenett/ror/pkg/telemetry/trace"

	healthserver "github.com/NorskHelsenett/ror/pkg/helpers/rorhealth/server"
	"github.com/NorskHelsenett/ror/pkg/rlog"

	"github.com/spf13/viper"
)

func main() {
	cancelChan := make(chan os.Signal, 1)
	stop := make(chan struct{})
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	rlog.Info("Audit micro service starting")
	auditconfig.Load()

	msauditconnections.InitConnections()

	healthserver.MustStartWithDefaults()

	go func() {
		trace.ConnectTracer(stop, viper.GetString(configconsts.ROLE), viper.GetString(configconsts.OPENTELEMETRY_COLLECTOR_ENDPOINT))
		sig := <-cancelChan
		_, _ = fmt.Println()
		_, _ = fmt.Println(sig)
		stop <- struct{}{}
	}()

	msauditrabbitmqhandler.StartListening()

	sig := <-cancelChan
	rlog.Info("Caught signal", rlog.Any("signal", sig))
}
