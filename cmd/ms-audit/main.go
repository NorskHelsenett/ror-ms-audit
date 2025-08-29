package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/NorskHelsenett/ror-ms-audit/internal/auditconfig"
	"github.com/NorskHelsenett/ror-ms-audit/internal/clients/rabbitmq/msauditrabbitmqdefinitions"
	"github.com/NorskHelsenett/ror-ms-audit/internal/clients/rabbitmq/msauditrabbitmqhandler"
	"github.com/NorskHelsenett/ror-ms-audit/internal/httpserver"
	"github.com/NorskHelsenett/ror-ms-audit/internal/msauditconnections"

	"syscall"

	"github.com/NorskHelsenett/ror/pkg/config/configconsts"

	"github.com/NorskHelsenett/ror/pkg/telemetry/trace"

	"github.com/NorskHelsenett/ror/pkg/rlog"

	"github.com/spf13/viper"

	// https://blog.devgenius.io/know-gomaxprocs-before-deploying-your-go-app-to-kubernetes-7a458fb63af1
	"go.uber.org/automaxprocs/maxprocs"
)

func init() {
	_, _ = maxprocs.Set(maxprocs.Logger(rlog.Infof))
}
func main() {
	cancelChan := make(chan os.Signal, 1)
	stop := make(chan struct{})
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	rlog.Info("Audit micro service starting")
	auditconfig.Load()

	msauditconnections.InitConnections()

	msauditrabbitmqdefinitions.InitOrDie()

	go func() {
		httpserver.InitHttpServer()
	}()

	go func() {
		trace.ConnectTracer(stop, viper.GetString(configconsts.ROLE), viper.GetString(configconsts.OPENTELEMETRY_COLLECTOR_ENDPOINT))
		sig := <-cancelChan
		_, _ = fmt.Println()
		_, _ = fmt.Println(sig)
		stop <- struct{}{}
	}()

	msauditrabbitmqhandler.StartListening()

	sig := <-cancelChan
	rlog.Info("Caught signal", rlog.Any("singal", sig))
	// shutdown other goroutines gracefully
	// close other resources
}
