package client

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/bentol/tero/backend"
	"github.com/bentol/tero/config"
	"github.com/bentol/tero/notif"
	"github.com/bentol/tero/tctl"
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

func AddUser(userName, stringRoles, sendEmailTo string) (string, error) {
	// make sure user not exist
	results, _ := backend.GetUsersByNames([]string{userName})
	if len(results) != 0 {
		return "", errors.New(fmt.Sprintf("User `%s` already exist", userName))
	}

	stdout, tokenString, err := tctl.CmdAddUser(userName, "")
	if err != nil {
		return "", err
	}

	err = backend.ConfigureNewUserToken(tokenString, strings.Split(stringRoles, ","))
	if err != nil {
		return "", err
	}

	// send email
	if config.Get().EnableEmailToken {
		if sendEmailTo == "" {
			sendEmailTo = userName + "@tokopedia.com"
		}
		err = notif.SentMailNewUser(userName, sendEmailTo, tokenString)
		if err != nil {
			return "", err
		}

		stdout = stdout + fmt.Sprintf("\r\n\r\nEmail sent to: %s", sendEmailTo)
	}

	return stdout, nil
}

func ShowUser(name string) (string, error) {
	users, err := backend.GetUsersByNames([]string{name})
	if err != nil {
		return "", err
	}

	if len(users) == 0 {
		return "", fmt.Errorf("User `%s` does not exist", name)
	}

	bufferUsersInfo := new(bytes.Buffer)
	tableUsers := tablewriter.NewWriter(bufferUsersInfo)
	tableUsers.SetHeader([]string{"Name", "Roles"})
	tableUsers.SetColMinWidth(1, 100)
	tableUsers.SetAutoMergeCells(true)

	user := users[0]
	roleInfo := make([]string, 0)
	for _, r := range user.Roles {
		if len(r.Name) == 0 {
			continue
		}
		info := r.Name + " = " + r.StringAllowedLogins() + "@" + r.StringNodePatterns()
		roleInfo = append(roleInfo, info)
	}

	u := users[0]
	for i, info := range roleInfo {
		if i == 0 {
			name = u.Name
		} else {
			name = ""
		}
		tableUsers.Append([]string{
			name,
			info,
		})
	}
	tableUsers.Render()

	result := bufferUsersInfo.String()
	return result, nil
}

func ListUser() (string, error) {
	users, err := backend.GetUsers()
	if err != nil {
		return "", err
	}

	bufferUsersInfo := new(bytes.Buffer)
	tableUsers := tablewriter.NewWriter(bufferUsersInfo)
	tableUsers.SetHeader([]string{"Name", "Locked", "Roles"})
	tableUsers.SetColMinWidth(2, 100)
	tableUsers.SetAutoMergeCells(true)
	tableUsers.SetRowLine(true)

	for _, user := range users {
		roleInfo := make([]string, 0)
		for _, r := range user.Roles {
			if len(r.Name) == 0 {
				continue
			}
			info := r.Name + " = " + r.StringAllowedLogins() + "@" + r.StringNodePatterns()
			roleInfo = append(roleInfo, info)
		}
		lockedStatus := "no"
		if user.IsLocked {
			lockedStatus = "yes"
		}

		for _, info := range roleInfo {
			name := user.Name
			tableUsers.Append([]string{
				name,
				lockedStatus,
				info,
			})
		}
	}

	tableUsers.Render()

	result := bufferUsersInfo.String()
	return result, nil
}

func LockUser(username string) (string, error) {
	err := backend.LockUser(username)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("User `%s` is locked!", username), nil
}

func UnlockUser(username string) (string, error) {
	err := backend.UnlockUser(username)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("User `%s` is unlocked!", username), nil
}

func DeleteUser(userName string) (string, error) {
	// make sure user not exist
	results, _ := backend.GetUsersByNames([]string{userName})
	if len(results) == 0 {
		return "", errors.New(fmt.Sprintf("User `%s` not exist", userName))
	}

	_, err := tctl.CmdDeleteUser(userName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("User `%s` deleted!", userName), nil
}
