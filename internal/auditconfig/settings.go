package auditconfig

import (
	"github.com/NorskHelsenett/ror/pkg/config/configconsts"
	"github.com/NorskHelsenett/ror/pkg/config/rorversion"

	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

var (
	Version   string = "0.1.0"
	Commit    string = "FFFFFFF"
	RorApiURL string = "http://localhost:8080"
	RorApiKey string = "my-secret-api-key"
)

var (
	VaultSecret *vault.Secret
)

func Load() {
	viper.SetDefault(configconsts.HELSEGITLAB_BASE_URL, "https://helsegitlab.nhn.no/api/v4/projects/")
	viper.SetDefault(configconsts.VAULT_URL, "http://localhost:8200")
	RorApiKey = viper.GetString(configconsts.API_KEY)
	RorApiURL = viper.GetString(configconsts.API_ENDPOINT)
	viper.AutomaticEnv()
}

func GetRorVersion() rorversion.RorVersion {
	return rorversion.NewRorVersion(viper.GetString(configconsts.VERSION), viper.GetString(configconsts.COMMIT))
}
