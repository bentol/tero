package backend

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bentol/tele/backend/dynamo"
	"github.com/bentol/tele/role"
)

var (
	storage Storage
)

type Storage interface {
	GetRoles() ([]role.Role, error)
	GetRoleByName(name string) (*role.Role, error)
	DeleteRole(name string) error
	CreateRole(name string, allowedLogins []string, nodePatterns map[string]string) (*role.Role, error)
	UpdateRole(name, allowedLogins []string, nodePatterns map[string]string) error
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

func CreateRole(name string, allowedLogins []string, nodePatterns map[string]string) (*role.Role, error) {
	checkStorage()

	// make sure role not exists
	existedRole, _ := storage.GetRoleByName(name)
	if existedRole != nil {
		return nil, fmt.Errorf("Role `%s` already exists", name)
	}

	return storage.CreateRole(name, allowedLogins, nodePatterns)
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
