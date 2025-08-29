package msauditconnections

import (
	"fmt"

	"github.com/NorskHelsenett/ror-ms-audit/internal/auditconfig"

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
	VaultClient        *vaultclient.VaultClient
	RabbitMQConnection rabbitmqclient.RabbitMQConnection
	RorClient          *rorclient.RorClient
	GitClient          *gitclient.GitClient
)

func InitConnections() {
	VaultClient = vaultclient.NewVaultClient(viper.GetString(configconsts.ROLE), viper.GetString(configconsts.VAULT_URL))
	rmqcredhelper := rabbitmqcredhelper.NewVaultRMQCredentials(VaultClient, viper.GetString(configconsts.ROLE))
	RabbitMQConnection = rabbitmqclient.NewRabbitMQConnectionWithDefaults(rabbitmqclient.OptionCredentialsProvider(rmqcredhelper))
	RorClient = mustInitRorClient()

	GitClient = gitclient.NewGitClient(viper.GetString(configconsts.GIT_REPO_URL), viper.GetString(configconsts.GIT_BRANCH), viper.GetString(configconsts.GIT_TOKEN))

	health.Register("vault", VaultClient)
	health.Register("rabbitmq", RabbitMQConnection)
}

func mustInitRorClient() *rorclient.RorClient {
	authProvider := httpauthprovider.NewAuthProvider(httpauthprovider.AuthPoviderTypeAPIKey, auditconfig.RorApiKey)
	clientConfig := httpclient.HttpTransportClientConfig{
		BaseURL:      auditconfig.RorApiURL,
		AuthProvider: authProvider,
		Version:      rorversion.GetRorVersion(),
		Role:         viper.GetString(configconsts.ROLE),
	}
	transport := resttransport.NewRorHttpTransport(&clientConfig)
	RorClient = rorclient.NewRorClient(transport)
	if err := RorClient.CheckConnection(); err != nil {
		fmt.Printf("failed to connect RorClient: %v", err)
		return nil
	}
	return RorClient
}
