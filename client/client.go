package client

import (
	"bytes"
	"fmt"

	"github.com/bentol/tele/backend"
	"github.com/olekukonko/tablewriter"
)

func NewRole(name, rawAllowedLogins, rawNodePatterns string) (string, error) {
	nodePatterns, err := backend.ParseNodePatterns(rawNodePatterns)
	if err != nil {
		return "", err
	}
	allowedLogins, err := backend.ParseAllowedLogins(rawAllowedLogins)
	if err != nil {
		return "", err
	}

	role, err := backend.CreateRole(name, allowedLogins, nodePatterns)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Role `%s` successfully created!", role.Name), nil
}

func ListRoles() (string, error) {
	result := new(bytes.Buffer)

	roles, _ := backend.GetRoles()

	data := make([][]string, 0)

	for _, role := range roles {
		data = append(data, []string{role.Name, role.StringAllowedLogins(), role.StringNodePatterns()})
	}

	table := tablewriter.NewWriter(result)
	table.SetHeader([]string{"Role", "Allowed Logins", "Node"})

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
