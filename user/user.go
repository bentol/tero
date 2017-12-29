package user

import (
	"github.com/Jeffail/gabs"
	"github.com/bentol/tele/role"
)

var (
	UserJsonTemplate string = `{"kind":"user","version":"v2","metadata":{"name":"user"},"spec":{"roles":["admin"],"traits":{"logins":["root","ubuntu", "dev", "bejo"]},"status":{"is_locked":false,"locked_time":"0001-01-01T00:00:00Z","lock_expires":"0001-01-01T00:00:00Z"},"expires":"0001-01-01T00:00:00Z","created_by":{"time":"0001-01-01T00:00:00Z","user":{"name":""}}}}`
)

type User struct {
	Name  string
	Roles []role.Role
}

func (u *User) RoleNames() []string {
	names := make([]string, 0)
	for _, r := range u.Roles {
		names = append(names, r.Name)
	}
	return names
}

func (u *User) GetJSON() string {
	jsonTemplate, _ := gabs.ParseJSON([]byte(UserJsonTemplate))
	jsonTemplate.SetP(u.Name, "metadata.name")
	jsonTemplate.SetP(u.RoleNames(), "spec.roles")
	jsonTemplate.SetP([]string{"ubuntu"}, "spec.traits.logins")
	return jsonTemplate.String()
}