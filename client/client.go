package client

import (
	"bytes"
	"fmt"
	"log"

	"github.com/bentol/tele/backend"
	"github.com/bentol/tele/role"
	"github.com/olekukonko/tablewriter"
)

func NewRole(name, rawAllowedLogins, rawNodePatterns string) (*role.Role, error) {
	nodePatterns, err := backend.ParseNodePatterns(rawNodePatterns)
	if err != nil {
		log.Fatal(err)
	}
	allowedLogins, err := backend.ParseAllowedLogins(rawAllowedLogins)
	if err != nil {
		log.Fatal(err)
	}

	role, err := backend.CreateRole(name, allowedLogins, nodePatterns)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func ListRoles() (string, error) {
	result := new(bytes.Buffer)

	roles, _ := backend.GetRoles()

	data := make([][]string, 0)

	for _, role := range roles {
		data = append(data, []string{role.Name, role.StringAllowedLogins(), role.StringNodePatterns()})
	}

	table := tablewriter.NewWriter(result)
	table.SetHeader([]string{"Name", "Allowed Logins", "Node"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render()
	return result.String(), nil
}

func DeleteRole(name string) (string, error) {
	role, _ := backend.GetRoleByName(name)
	if role == nil {
		return "Role not exists", nil
	}

	err := backend.DeleteRole(name)
	if err == nil {
		return fmt.Sprintf("Role `%s` deleted!", name), nil
	}
	return fmt.Sprintf("Failed to delete role: %s", err), nil
}
