package auditservice

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/NorskHelsenett/ror-ms-audit/internal/auditconfig"
	"github.com/spf13/viper"

	"github.com/NorskHelsenett/ror/pkg/clients/gitclient"
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

	md, err := CreateMarkdown(*acls)
	if err != nil {
		rlog.Fatalc(ctx, "could not create markdown of acl list ...", nil)
	}

	path := viper.GetString(configconsts.GIT_PATH)

	err = auditconfig.GitClient.UploadFile(path, []byte(md), fmt.Sprintf("Updated %s", path), gitclient.OptionBranch(viper.GetString(configconsts.GIT_BRANCH)), gitclient.OptionDepth(1))
	if err != nil {
		rlog.Errorc(ctx, "could not update file in git ...", err)
		return
	}

	rlog.Info("Acl updated successfully")
}

var (
	tableheader = "| Group | Read | Create | Update | Delete | Owner | Kubernetes.Logon | Issued by | Created |\n|---|---|---|---|---|---|---|---|---|\n"
)

func CreateMarkdown(acls []aclmodels.AclV2ListItem) (string, error) {
	var sb strings.Builder
	// not indenting because of result in file
	sb.WriteString(`# Autorisasjonsregister

## Etter scope

`)

	scopes := getAllScopes(acls)
	for _, scope := range scopes {
		sb.WriteString(fmt.Sprintf("### %s\n\n", scope))
		subjects := getAllSubjects(scope, acls)
		for _, subject := range subjects {
			sb.WriteString(fmt.Sprintf("#### %s\n", subject))
			sb.WriteString("??? note \"access\"\n")
			scopeAcls := getAclsBySubject(scope, acls, subject)
			sb.WriteString(getTableIntend(scopeAcls, 4))
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

func getTableIntend(acls []aclmodels.AclV2ListItem, intended int) string {
	text := getTable(acls)
	intend := strings.Repeat(" ", intended)
	text = strings.ReplaceAll(text, "\n", "\n"+intend)
	text = intend + text
	return text
}

func getTable(acls []aclmodels.AclV2ListItem) string {
	var sb strings.Builder
	// not indenting because of result in file

	sb.WriteString(tableheader)

	for _, acl := range acls {
		sb.WriteString(getTableRow(acl))
	}

	md := sb.String()
	return md
}

func getTableRow(acl aclmodels.AclV2ListItem) string {
	return fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
		acl.Group,
		getEmojiByBool(acl.Access.Read), getEmojiByBool(acl.Access.Create), getEmojiByBool(acl.Access.Update), getEmojiByBool(acl.Access.Delete), getEmojiByBool(acl.Access.Owner),
		getEmojiByBool(acl.Kubernetes.Logon), acl.IssuedBy, acl.Created.String())
}

func getAllSubjects(scope string, acls []aclmodels.AclV2ListItem) []string {
	aclscope := aclmodels.Acl2Scope(scope)
	var filteredAcls []aclmodels.AclV2ListItem
	for _, acl := range acls {
		if acl.Scope == aclscope {
			filteredAcls = append(filteredAcls, acl)
		}
	}
	subjectMap := make(map[string]bool)
	for _, acl := range filteredAcls {
		subjectMap[string(acl.Subject)] = true
	}
	var subjects []string
	for subject := range subjectMap {
		subjects = append(subjects, subject)
	}
	// Sort subjects alphabetically
	sort.Strings(subjects)
	return subjects
}

func getAllScopes(acls []aclmodels.AclV2ListItem) []string {
	scopeMap := make(map[string]bool)
	for _, acl := range acls {
		scopeMap[string(acl.Scope)] = true
	}
	var scopes []string

	// Enforce order for these scopes
	priority := []string{"ror", "project", "cluster"}
	for _, p := range priority {
		if scopeMap[p] {
			scopes = append(scopes, p)
			delete(scopeMap, p)
		}
	}
	// Add the rest (order not guaranteed)
	for scope := range scopeMap {
		scopes = append(scopes, scope)
	}
	return scopes
}

func getAclsBySubject(scope string, acls []aclmodels.AclV2ListItem, subject string) []aclmodels.AclV2ListItem {
	var filteredAcls []aclmodels.AclV2ListItem
	aclScope := aclmodels.Acl2Scope(scope)
	aclsubject := aclmodels.Acl2Subject(subject)
	for _, acl := range acls {
		if acl.Subject == aclsubject && acl.Scope == aclScope {
			filteredAcls = append(filteredAcls, acl)
		}
	}
	return filteredAcls
}

// func getAclsByScope(acls []aclmodels.AclV2ListItem, scope string) []aclmodels.AclV2ListItem {
// 	var filteredAcls []aclmodels.AclV2ListItem
// 	aclScope := aclmodels.Acl2Scope(scope)
// 	for _, acl := range acls {
// 		if acl.Scope == aclScope {
// 			filteredAcls = append(filteredAcls, acl)
// 		}
// 	}
// 	return filteredAcls

// }

func getEmojiByBool(boolean bool) string {
	if boolean {
		return ":white_check_mark:"
	} else {
		return ":x:"
	}
}
