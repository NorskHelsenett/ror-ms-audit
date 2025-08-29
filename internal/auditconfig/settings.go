package auditconfig

import (
	"github.com/NorskHelsenett/ror/pkg/config/configconsts"

	"github.com/spf13/viper"
)

var (
	RorApiURL string = "http://localhost:8080"
	RorApiKey string = "my-secret-api-key"
)

func Load() {
	viper.SetDefault(configconsts.ROLE, "ror-ms-audit")
	viper.SetDefault(configconsts.VAULT_URL, "http://localhost:8200")
	viper.SetDefault(configconsts.HTTP_HEALTH_HOST, "0.0.0.0")
	viper.SetDefault(configconsts.HTTP_HEALTH_PORT, "8080")
	RorApiKey = viper.GetString(configconsts.API_KEY)
	RorApiURL = viper.GetString(configconsts.API_ENDPOINT)
	viper.AutomaticEnv()
}
