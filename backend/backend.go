package backend

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bentol/tero/backend/dynamo"
	"github.com/bentol/tero/role"
	"github.com/bentol/tero/token"
	"github.com/bentol/tero/user"
)

var (
	storage Storage
)

type Storage interface {
	GetRoles() ([]role.Role, error)
	GetRoleByName(name string) (*role.Role, error)
	DeleteRole(name string) error
	CreateRole(name string, allowedLogins []string, nodePatterns map[string]string) (*role.Role, error)
	UpdateRole(name string, allowedLogins []string, nodePatterns map[string]string) (*role.Role, error)
	AttachRole(selectedRole *role.Role, users []user.User) ([]user.User, error)
	DetachRole(selectedRole *role.Role, users []user.User) ([]user.User, error)
	GetUsers() (map[string]user.User, error)
	GetUserByName(name string) (*user.User, error)
	GetUsersByNames(names []string) ([]user.User, error)
	GetUsersByRole(name string) ([]user.User, error)
	GetAddUserToken(token string) (*token.AddUserToken, error)
	GetAddUserTokenByUserName(userName string) (*token.AddUserToken, error)
	InsertItem(path, value string, ttl int64) error
	UpdateAddUserToken(token *token.AddUserToken) error
	SetUserLockedStatus(username string, status bool) error
}

func InitBackend(selectedStorage string) {
	switch selectedStorage {
	case "dynamodb":
		storage = dynamo.New()
	}
}

func checkStorage() {
	if storage == nil {
		log.Fatal("Error: Storage not initialized")
	}
}

func GetStorage() Storage {
	return storage
}

func GetRoles() ([]role.Role, error) {
	checkStorage()
	roles, err := storage.GetRoles()
	return roles, err
}

func DeleteRole(name string) error {
	checkStorage()
	return storage.DeleteRole(name)
}

func GetRoleByName(name string) (*role.Role, error) {
	checkStorage()
	return storage.GetRoleByName(name)
}

func GetUsersByNames(names []string) ([]user.User, error) {
	checkStorage()
	return storage.GetUsersByNames(names)
}

func CreateRole(name string, allowedLogins []string, nodePatterns map[string]string) (*role.Role, error) {
	checkStorage()

	// make sure role not exists
	existedRole, _ := storage.GetRoleByName(name)
	if existedRole != nil {
		return nil, fmt.Errorf("Role `%s` already exists", name)
	}

	return storage.CreateRole(name, allowedLogins, nodePatterns)
}

func UpdateRole(name string, allowedLogins []string, nodePatterns map[string]string) (*role.Role, error) {
	checkStorage()

	// make sure role exists
	existedRole, _ := storage.GetRoleByName(name)
	if existedRole == nil {
		return nil, fmt.Errorf("Role `%s` doesn't exists", name)
	}

	return storage.UpdateRole(name, allowedLogins, nodePatterns)
}

func AttachRole(name string, users []string) ([]user.User, error) {
	checkStorage()

	role, err := storage.GetRoleByName(name)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("Role `%s` does not exist", name)
	}

	listUsers, err := storage.GetUsersByNames(users)
	if len(users) != len(listUsers) {
		return nil, errors.New("One or more user does not exist")
	}

	return storage.AttachRole(role, listUsers)
}

func DettachRole(name string, users []string) ([]user.User, error) {
	checkStorage()

	role, err := storage.GetRoleByName(name)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("Role `%s` does not exist", name)
	}

	listUsers, err := storage.GetUsersByNames(users)
	if len(users) != len(listUsers) {
		return nil, errors.New("One or more user does not exist")
	}

	return storage.DetachRole(role, listUsers)
}

func GetUsers() (map[string]user.User, error) {
	checkStorage()
	return storage.GetUsers()
}

func GetUsersByRole(name string) ([]user.User, error) {
	checkStorage()
	return storage.GetUsersByRole(name)
}

func ConfigureNewUserToken(token string, roles []string) error {
	checkStorage()
	addUserToken, err := storage.GetAddUserToken(token)
	if addUserToken == nil {
		return errors.New("Add user token not found")
	}

	addUserToken.SetRoles(roles)
	err = storage.UpdateAddUserToken(addUserToken)
	if err != nil {
		return err
	}

	return nil
}

func ParseNodePatterns(rawPatterns string) (map[string]string, error) {
	result := map[string]string{}

	splitResult := strings.Split(rawPatterns, ",")
	for _, line := range splitResult {
		if strings.Count(line, ":") != 1 {
			return nil, errors.New("Invalid node patterns")
		}

		s := strings.SplitN(line, ":", 2)
		result[s[0]] = s[1]
	}

	return result, nil
}

func ParseAllowedLogins(rawAllowedLogins string) ([]string, error) {
	ret := strings.Split(rawAllowedLogins, ",")
	if len(ret) == 0 {
		return nil, errors.New("Allowed logins must not empty")
	}

	return ret, nil
}

func UnlockUser(username string) error {
	return storage.SetUserLockedStatus(username, false)
}

func LockUser(username string) error {
	return storage.SetUserLockedStatus(username, true)
}
