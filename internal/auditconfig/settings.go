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

	"github.com/spf13/viper"
)

var (
	RorApiURL string = "http://localhost:8080"
	RorApiKey string = "my-secret-api-key"

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
	RorApiKey = viper.GetString(configconsts.API_KEY)
	RorApiURL = viper.GetString(configconsts.API_ENDPOINT)
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

	GitClient = gitclient.NewGitClient(viper.GetString(configconsts.GIT_REPO_URL), viper.GetString(configconsts.GIT_BRANCH), viper.GetString(configconsts.GIT_TOKEN))

	health.Register("vault", VaultClient)
	health.Register("rabbitmq", RabbitMQConnection)
	health.Register("rorclient", RorClient)
	health.Register("gitclient", GitClient)
}

func mustInitRorClient() *rorclient.RorClient {
	authProvider := httpauthprovider.NewAuthProvider(httpauthprovider.AuthPoviderTypeAPIKey, RorApiKey)
	clientConfig := httpclient.HttpTransportClientConfig{
		BaseURL:      RorApiURL,
		AuthProvider: authProvider,
		Version:      rorversion.GetRorVersion(),
		Role:         viper.GetString(configconsts.ROLE),
	}
	transport := resttransport.NewRorHttpTransport(&clientConfig)
	RorClient = rorclient.NewRorClient(transport)
	if err := RorClient.CheckConnection(); err != nil {
		fmt.Printf("failed to ping RorClient: %v", err)
		return nil
	}
	return RorClient
}
