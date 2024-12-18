package auditconfig

import (
	"github.com/NorskHelsenett/ror/pkg/config/configconsts"
	"github.com/NorskHelsenett/ror/pkg/config/rorversion"

	"github.com/NorskHelsenett/ror/pkg/rlog"

	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

var Version string = "0.1.0"
var Commit string = "FFFFFFF"

var (
	VaultSecret *vault.Secret
)

func init() {

	viper.SetDefault(configconsts.HELSEGITLAB_BASE_URL, "https://helsegitlab.nhn.no/api/v4/projects/")
	viper.SetDefault(configconsts.VAULT_URL, "http://localhost:8200")

	viper.AutomaticEnv()
}

func Load() {
	environment := viper.GetString(configconsts.ENVIRONMENT)
	rlog.Info("loaded environment", rlog.String("Environment", environment))

	_ = viper.WriteConfig()
}

func GetRorVersion() rorversion.RorVersion {
	return rorversion.NewRorVersion(viper.GetString(configconsts.VERSION), viper.GetString(configconsts.COMMIT))
}
