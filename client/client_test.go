package client

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/bentol/tele/backend"
	"github.com/bentol/tele/client"
	"github.com/stretchr/testify/assert"
)

func init() {
	setup()
	// remove test role
	roles, _ := backend.GetRoles()
	for _, r := range roles {
		if strings.Contains(r.Name, "test-role-") {
			backend.DeleteRole(r.Name)
		}
	}
}

func setup() {
	backend.InitBackend("dynamodb")
}

func TestNewRole_shouldCreateNewRole(t *testing.T) {
	_, _ = client.DeleteRole("brand_new_role")
	out, err := client.NewRole("brand_new_role", "ubuntu,root,admin", "app:tome,env:production")
	if err != nil {
		t.Error("add role failed with valid input")
		t.Error(err)
	}
	assert.Contains(t, out, "created")

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
	_, _ = client.DeleteRole("second_role")

	_, _ = client.NewRole("second_role", "ubuntu,root,admin", "app:tome,env:production")

	_, err := client.NewRole("second_role", "ubuntu,root,admin", "app:tome,env:production")
	assert.Contains(t, err.Error(), "already exists")
}

func TestDeleteRole_shouldRemoveRole(t *testing.T) {
	client.NewRole("existed_role", "ubuntu", "env:production")
	client.DeleteRole("existed_role")
	role, _ := backend.GetRoleByName("existed_role")
	if role != nil {
		t.Error("Role should not exist after deleted")
	}
}

func TestListRole_shouldDisplayAllRoles(t *testing.T) {
	client.NewRole("role_one", "ubuntu", "env:production")
	client.NewRole("role_two", "dev", "env:staging")

	result, err := client.ListRoles()
	if err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, result, "role_one")
	assert.Contains(t, result, "role_two")
}

func TestUpdateRole_shouldChangeItsAttribute(t *testing.T) {
	_, _ = client.DeleteRole("to_be_updated")
	_, _ = client.NewRole("to_be_updated", "ubuntu", "env:staging")
	out, err := client.UpdateRole("to_be_updated", "root,dev", "app:tome,env:production")
	if err != nil {
		t.Fatal("error: " + err.Error())
	}
	assert.Equal(t, out, "Role `to_be_updated` successfully updated!")

	role, _ := backend.GetRoleByName("to_be_updated")
	assert.Contains(t, role.AllowedLogins, "root")
	assert.Contains(t, role.AllowedLogins, "dev")
	assert.Contains(t, "production", role.NodePatterns["env"])
	assert.Contains(t, "tome", role.NodePatterns["app"])
}

func TestAttachRole_shouldErrorIfRoleNotExist(t *testing.T) {
	out, err := client.AttachRole("imaginary_role", "beni,budi")
	assert.NotNil(t, err, "Attach non existant role should failed")
	assert.Equal(t, err.Error(), fmt.Sprintf("Role `%s` does not exist", "imaginary_role"))
	assert.Empty(t, out)
}

func TestAttachRole_shouldErrorIfUsersDoesNotExist(t *testing.T) {
	out, err := client.AttachRole("admin", "imaginary_user")
	assert.NotNil(t, err, "Attach role to non existant user should failed")
	assert.Empty(t, out)
}

func TestAttachRole_shouldAttachTheRoleToUser(t *testing.T) {
	roleName := "test-role-" + strconv.Itoa(rand.Int())
	_, _ = client.NewRole(roleName, "ubuntu", "env:production")

	out, err := client.AttachRole(roleName, "beni,hulk")
	assert.Nil(t, err, "Attach valid role should not error")
	assert.Contains(t, out, fmt.Sprintf("Role `%s` successfully attached!", roleName))

	users, _ := backend.GetUsersByNames([]string{"beni", "hulk"})
	for _, u := range users {
		assert.Contains(t, u.RoleNames(), roleName)
	}
}

func TestDettachRole_shouldErrorIfRoleNotExist(t *testing.T) {
	out, err := client.DetachRole("imaginary_role", "beni,budi")
	assert.NotNil(t, err, "Detach non existant role should failed")
	assert.Equal(t, err.Error(), fmt.Sprintf("Role `%s` does not exist", "imaginary_role"))
	assert.Empty(t, out)
}

func TestDetachRole_shouldErrorIfUsersDoesNotExist(t *testing.T) {
	roleName := "test-role-" + strconv.Itoa(rand.Int())
	_, _ = client.NewRole(roleName, "ubuntu", "env:production")
	out, err := client.DetachRole(roleName, "imaginary_user")
	assert.NotNil(t, err, "Detach role to non existant user should failed")
	assert.Empty(t, out)
}

func TestDetachRole_shouldDetachTheRoleFromUser(t *testing.T) {
	roleName := "test-role-" + strconv.Itoa(rand.Int())
	_, _ = client.NewRole(roleName, "ubuntu", "env:production")
	_, err := client.AttachRole(roleName, "beni,hulk")

	users, _ := backend.GetUsersByNames([]string{"beni", "hulk"})
	for _, u := range users {
		assert.Contains(t, u.RoleNames(), roleName)
	}

	out, err := client.DetachRole(roleName, "beni,hulk")
	assert.Nil(t, err, "Detach role with valid input should not error")
	assert.Contains(t,
		out,
		fmt.Sprintf("Role `%s` successfully detached from [%s]!", roleName, "beni,hulk"),
	)

	users, _ = backend.GetUsersByNames([]string{"beni", "hulk"})
	for _, u := range users {
		assert.NotContains(t, u.RoleNames(), roleName)
	}
}

func TestShowRole_shouldErrorIfRoleDoesNotExist(t *testing.T) {
	out, err := client.ShowRole("imaginary_role")
	assert.Empty(t, out)
	assert.Equal(t, err.Error(), fmt.Sprintf("Role `%s` does not exist", "imaginary_role"))
}

func TestShowRole_shouldDisplayItsAllowedLoginsAndNodePatterns(t *testing.T) {
	roleName := "test-role-" + strconv.Itoa(rand.Int())
	_, _ = client.NewRole(roleName, "avengers,monster", "env:production,app:jet")
	_, _ = client.AttachRole(roleName, "hulk")

	out, err := client.ShowRole(roleName)
	assert.Nil(t, err)
	assert.Contains(t, out, "avengers")
	assert.Contains(t, out, "monster")
	assert.Contains(t, out, "env:production")
	assert.Contains(t, out, "app:jet")

	assert.Contains(t, out, "hulk")
}

func TestAddUser_shouldCreateAddUserSessionWithSelectedRole(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test")
	}
	userName := "test-user-" + strconv.Itoa(rand.Int())
	roleName1 := "test-role-" + strconv.Itoa(rand.Int())
	_, _ = client.NewRole(roleName1, "avengers,monster", "env:production,app:jet")
	roleName2 := "test-role-" + strconv.Itoa(rand.Int())
	_, _ = client.NewRole(roleName2, "hydra", "env:staging,app:bus")

	client.AddUser(userName, roleName1+","+roleName2)
	addUserToken, err := backend.GetStorage().GetAddUserTokenByUserName(userName)
	assert.NotNil(t, addUserToken)
	assert.Nil(t, err)
	assert.EqualValues(t, addUserToken.GetStringRoles(), []string{roleName1, roleName2})
}

func TestShowUser_shouldErrorIfUserDoesNotExist(t *testing.T) {
	out, err := client.ShowUser("imaginary_user")
	assert.Empty(t, out)
	assert.Equal(t, err.Error(), fmt.Sprintf("User `%s` does not exist", "imaginary_user"))
}

func TestShowUser_shouldDisplayItsRoleAndAllowedLogins(t *testing.T) {
	roleName1 := "test-role-" + strconv.Itoa(rand.Int())
	roleName2 := "test-role-" + strconv.Itoa(rand.Int())
	_, _ = client.NewRole(roleName1, "ubuntu", "env:production,app:plane")
	_, _ = client.NewRole(roleName2, "root", "env:staging,app:ship")
	_, _ = client.AttachRole(roleName1, "hulk")
	_, _ = client.AttachRole(roleName2, "hulk")

	out, err := client.ShowUser("hulk")
	assert.Nil(t, err)
	assert.Contains(t, out, "ubuntu")
	assert.Contains(t, out, "root")
	assert.Contains(t, out, "env:production")
	assert.Contains(t, out, "env:staging")
	assert.Contains(t, out, "app:plane")
	assert.Contains(t, out, "app:ship")

	assert.Contains(t, out, "hulk")
}
