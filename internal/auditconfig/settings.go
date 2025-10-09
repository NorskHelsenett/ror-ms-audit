package auditconfig

import (
	"context"
	"fmt"

	"github.com/NorskHelsenett/ror/pkg/clients/gitclient"
	"github.com/NorskHelsenett/ror/pkg/clients/rabbitmqclient"
	"github.com/NorskHelsenett/ror/pkg/clients/rorclient"
	"github.com/NorskHelsenett/ror/pkg/clients/rorclient/transports/resttransport"
	"github.com/NorskHelsenett/ror/pkg/clients/rorclient/transports/resttransport/httpauthprovider"
	"github.com/NorskHelsenett/ror/pkg/clients/rorclient/transports/resttransport/httpclient"
	"github.com/NorskHelsenett/ror/pkg/clients/vaultclient"
	"github.com/NorskHelsenett/ror/pkg/clients/vaultclient/rabbitmqcredhelper"
	"github.com/NorskHelsenett/ror/pkg/config/configconsts"
	"github.com/NorskHelsenett/ror/pkg/config/rorversion"
	health "github.com/NorskHelsenett/ror/pkg/helpers/rorhealth"
	"github.com/NorskHelsenett/ror/pkg/rlog"

	"github.com/spf13/viper"
)

var (
	VaultClient        *vaultclient.VaultClient
	RabbitMQConnection rabbitmqclient.RabbitMQConnection
	RorClient          *rorclient.RorClient
	GitClient          *gitclient.GitClient
	ctx                context.Context = context.TODO()
)

func Load() {
	viper.SetDefault(configconsts.ROLE, "ror-ms-audit")
	viper.SetDefault(configconsts.VAULT_URL, "http://localhost:8200")
	viper.SetDefault(configconsts.HTTP_HEALTH_HOST, "0.0.0.0")
	viper.SetDefault(configconsts.HTTP_HEALTH_PORT, "8080")
	viper.SetDefault(configconsts.GIT_PATH, "auth.md")
	viper.SetDefault("RABBITMQ_QUEUE_NAME", "ms-audit")

	viper.AutomaticEnv()
	initConnections()
}

func Done() {
	// Clean up resources
	ctx.Done()
}

func initConnections() {
	VaultClient = vaultclient.NewVaultClient(viper.GetString(configconsts.ROLE), viper.GetString(configconsts.VAULT_URL))
	rmqcredhelper := rabbitmqcredhelper.NewVaultRMQCredentials(VaultClient, viper.GetString(configconsts.ROLE))
	RabbitMQConnection = rabbitmqclient.NewRabbitMQConnectionWithDefaults(rabbitmqclient.OptionCredentialsProvider(rmqcredhelper))
	RorClient = mustInitRorClient()

	GitClient = gitclient.NewGitClient(
		viper.GetString(configconsts.GIT_REPO_URL),
		viper.GetString(configconsts.GIT_TOKEN),
		gitclient.OptionAuthor(viper.GetString(configconsts.ROLE), fmt.Sprintf("%s@ror.system", viper.GetString(configconsts.ROLE))),
		gitclient.OptionBranch(viper.GetString(configconsts.GIT_BRANCH)),
	)

	//nolint:staticcheck // TODO: Migrate to RegisterWithContext
	health.Register("vault", VaultClient)
	//nolint:staticcheck // TODO: Migrate to RegisterWithContext
	health.Register("rabbitmq", RabbitMQConnection)
	//nolint:staticcheck // TODO: Migrate to RegisterWithContext
	health.Register("rorclient", RorClient)
	//nolint:staticcheck // TODO: Migrate to RegisterWithContext
	health.Register("gitclient", GitClient)
}

func mustInitRorClient() *rorclient.RorClient {
	rlog.Infof("Initializing ROR client, api: %s", viper.GetString(configconsts.API_ENDPOINT))
	if viper.GetString(configconsts.API_KEY) == "" {
		panic("API_KEY is not set")
	}
	if viper.GetString(configconsts.API_ENDPOINT) == "" {
		panic("ROR_URL is not set")
	}
	authProvider := httpauthprovider.NewAuthProvider(httpauthprovider.AuthPoviderTypeAPIKey, viper.GetString(configconsts.API_KEY))
	clientConfig := httpclient.HttpTransportClientConfig{
		BaseURL:      viper.GetString(configconsts.API_ENDPOINT),
		AuthProvider: authProvider,
		Version:      rorversion.GetRorVersion(),
		Role:         viper.GetString(configconsts.ROLE),
	}

	transport := resttransport.NewRorHttpTransport(&clientConfig)
	RorClient = rorclient.NewRorClient(transport)
	if err := RorClient.CheckConnection(); err != nil {
		rlog.Error("failed to ping RorClient", err)
		panic("failed to ping RorClient")
	}

	return RorClient
}
