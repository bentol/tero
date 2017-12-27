package role

import (
	"fmt"
	"sort"
	"strings"
)

var (
	RoleJsonTemplate string = `{"kind":"role","version":"v3","metadata":{"name":"new_role"},"spec":{"options":{"max_session_ttl":"30h0m0s"},"allow":{"logins":["tmp"],"node_labels":{"tmp":"tmp"},"rules":[{"resources":["role"],"verbs":["list","create","read","update","delete"]},{"resources":["auth_connector"],"verbs":["list","create","read","update","delete"]},{"resources":["session"],"verbs":["list","read"]},{"resources":["trusted_cluster"],"verbs":["list","create","read","update","delete"]}]},"deny":{}}}`
)

type Role struct {
	Name          string
	NodePatterns  map[string]string
	AllowedLogins []string
}

func (r *Role) StringAllowedLogins() string {
	logins := r.AllowedLogins
	sort.Strings(logins)
	return strings.Join(logins, ",")
}

func (r *Role) StringNodePatterns() string {
	listNodes := make([]string, 0)
	for k, v := range r.NodePatterns {
		listNodes = append(listNodes, fmt.Sprintf("%s:%s", k, v))
	}

	return strings.Join(listNodes, ",")
}
