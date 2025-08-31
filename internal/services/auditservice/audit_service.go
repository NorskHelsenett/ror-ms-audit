package auditservice

import (
	"context"
	"fmt"
	"strings"

	"github.com/NorskHelsenett/ror-ms-audit/internal/auditconfig"
	"github.com/spf13/viper"

	"github.com/NorskHelsenett/ror/pkg/config/configconsts"
	"github.com/NorskHelsenett/ror/pkg/messagebuscontracts"
	"github.com/NorskHelsenett/ror/pkg/models/aclmodels"

	"github.com/NorskHelsenett/ror/pkg/rlog"
)

func CreateAndCommitAclList(ctx context.Context, event messagebuscontracts.AclUpdateEvent) {

	acls, err := auditconfig.RorClient.Acl().GetAll(ctx)
	if err != nil {
		rlog.Fatalc(ctx, "could not get acl items ...", nil)
	}

	md, err := createMarkdown(*acls)
	if err != nil {
		rlog.Fatalc(ctx, "could not create markdown of acl list ...", nil)
	}

	path := viper.GetString(configconsts.GIT_PATH)

	err = auditconfig.GitClient.UploadFile(path, []byte(md), fmt.Sprintf("Updated %s", path))
	if err != nil {
		rlog.Fatalc(ctx, "could not update file in git ...", err)
	}

	rlog.Debugc(ctx, "acl updated")
}

func createMarkdown(acls []aclmodels.AclV2ListItem) (string, error) {
	var sb strings.Builder
	// not indenting because of result in file
	sb.WriteString(`# Autorisasjonsregister

## Liste

| Group | Scope | Subject | Read | Create | Update | Delete | Owner | Kubernetes.Logon | Issued by | Created |
|---|---|---|---|---|---|---|---|---|---|---|
`)

	for _, acl := range acls {
		aclItem := fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			acl.Group, string(acl.Scope), acl.Subject,
			getEmojiByBool(acl.Access.Read), getEmojiByBool(acl.Access.Create), getEmojiByBool(acl.Access.Update), getEmojiByBool(acl.Access.Delete), getEmojiByBool(acl.Access.Owner),
			getEmojiByBool(acl.Kubernetes.Logon), acl.IssuedBy, acl.Created.String())
		sb.WriteString(aclItem)
	}

	md := sb.String()
	return md, nil
}

func getEmojiByBool(boolean bool) string {
	if boolean {
		return ":white_check_mark:"
	} else {
		return ":x:"
	}
}
