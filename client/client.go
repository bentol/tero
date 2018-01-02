package client

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bentol/tele/backend"
	"github.com/bentol/tele/tctl"
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
		return "Role doesn't exists", nil
	}

	err := backend.DeleteRole(name)
	if err == nil {
		return fmt.Sprintf("Role `%s` deleted!", name), nil
	}
	return fmt.Sprintf("Failed to delete role: %s", err), nil
}

func UpdateRole(name, rawAllowedLogins, rawNodePatterns string) (string, error) {
	nodePatterns, err := backend.ParseNodePatterns(rawNodePatterns)
	if err != nil {
		return "", err
	}
	allowedLogins, err := backend.ParseAllowedLogins(rawAllowedLogins)
	if err != nil {
		return "", err
	}

	_, err = backend.UpdateRole(name, allowedLogins, nodePatterns)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Role `%s` successfully updated!", name), nil
}

func AttachRole(name string, rawUsers string) (string, error) {
	users := strings.Split(rawUsers, ",")
	_, err := backend.AttachRole(name, users)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Role `%s` successfully attached!", name), nil
}

func DetachRole(name string, rawUsers string) (string, error) {
	users := strings.Split(rawUsers, ",")
	_, err := backend.DettachRole(name, users)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Role `%s` successfully detached from [%s]!", name, rawUsers), nil
}

func ShowRole(name string) (string, error) {
	r, err := backend.GetRoleByName(name)
	if r == nil {
		return "", fmt.Errorf("Role `%s` does not exist", name)
	}
	if err != nil {
		return "", err
	}

	users, err := backend.GetUsersByRole(r.Name)
	if err != nil {
		return "", err
	}

	bufferRoleInfo := new(bytes.Buffer)
	table := tablewriter.NewWriter(bufferRoleInfo)
	table.SetHeader([]string{"Role", "Allowed Logins", "Node"})

	table.Append([]string{
		r.Name,
		r.StringAllowedLogins(),
		r.StringNodePatterns(),
	})
	table.Render()

	bufferUsersInfo := new(bytes.Buffer)
	tableUsers := tablewriter.NewWriter(bufferUsersInfo)
	tableUsers.SetHeader([]string{"Name", "Roles"})

	for _, u := range users {
		tableUsers.Append([]string{
			u.Name,
			strings.Join(u.RoleNames(), ", "),
		})
	}
	tableUsers.Render()

	result := "Role Info\n" + bufferRoleInfo.String() + "\n\nUsers\n" + bufferUsersInfo.String() + "\n"
	return result, nil
}

func AddUser(userName string, stringRoles string) (string, error) {
	stdout, tokenString, err := tctl.CmdAddUser(userName, "")
	if err != nil {
		return "", err
	}

	err = backend.ConfigureNewUserToken(tokenString, strings.Split(stringRoles, ","))
	if err != nil {
		return "", err
	}
	return stdout, nil
}
