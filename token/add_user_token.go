package token

import (
	"github.com/Jeffail/gabs"
)

type AddUserToken struct {
	Token string
	JSON  []byte
}

func (t *AddUserToken) GetStringRoles() []string {
	json, _ := gabs.ParseJSON(t.JSON)
	roles := make([]string, 0)
	rawRoles := json.Path("user.roles").Data()
	if rawRoles == nil {
		return make([]string, 0)
	}

	for _, raw := range rawRoles.([]interface{}) {
		roles = append(roles, raw.(string))
	}
	return roles
}

func (t *AddUserToken) SetRoles(roles []string) {
	json, _ := gabs.ParseJSON(t.JSON)
	json.SetP(roles, "user.roles")
	t.JSON = json.Bytes()
}
