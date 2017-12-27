package client

import (
	"testing"

	"github.com/bentol/tele/backend"
	"github.com/bentol/tele/client"
	"github.com/stretchr/testify/assert"
)

func setup() {
	backend.InitBackend("dynamodb")
}

func TestNewRole_shouldCreateNewRole(t *testing.T) {
	setup()
	err := client.DeleteRole("brand_new_role")
	if err != nil {
		t.Error("Delete role must not fail")
	}
	_, err = client.NewRole("brand_new_role", "ubuntu,root,admin", "app:tome,env:production")
	if err != nil {
		t.Error("add role failed with valid input")
		t.Error(err)
	}

	role, _ := backend.GetRoleByName("brand_new_role")
	if role == nil {
		t.Fatal("Role not created")
	} else {
		assert.Equal(t, "brand_new_role", role.Name)
		assert.Equal(t, "tome", role.NodePatterns["app"])
		assert.Equal(t, "production", role.NodePatterns["env"])
		assert.Equal(t, []string{"ubuntu", "root", "admin"}, role.AllowedLogins)
	}
}

func TestNewRole_cannotCreateNewRoleThatAlreadyExists(t *testing.T) {
	setup()
	err := client.DeleteRole("second_role")
	if err != nil {
		t.Error("Delete role must not fail")
	}

	_, _ = client.NewRole("second_role", "ubuntu,root,admin", "app:tome,env:production")

	role, _ := client.NewRole("second_role", "ubuntu,root,admin", "app:tome,env:production")
	if role != nil {
		t.Error("Role must not created if already exists")
	}
}

func TestDeleteRole_shouldRemoveRole(t *testing.T) {
	setup()
	client.NewRole("existed_role", "ubuntu", "env:production")
	client.DeleteRole("existed_role")
	role, _ := backend.GetRoleByName("existed_role")
	if role != nil {
		t.Error("Role must deleted")
	}
}

func TestListRole_shouldDisplayAllRoles(t *testing.T) {
	setup()
	client.NewRole("role_one", "ubuntu", "env:production")
	client.NewRole("role_two", "dev", "env:staging")

	result, err := client.ListRoles()
	if err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, result, "role_one")
	assert.Contains(t, result, "role_two")
}
