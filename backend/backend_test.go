package backend

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/bentol/tele/backend"
	"github.com/bentol/tele/client"
	"github.com/stretchr/testify/assert"
)

func init() {
	setup()
}

func setup() {
	backend.InitBackend("dynamodb")
}

func TestParseNodePatterns_should_parse_properly(t *testing.T) {
	input := "app:tome,env:production"
	ret, _ := backend.ParseNodePatterns(input)
	if (ret["app"] != "tome") || (ret["env"] != "production") {
		t.Error("node patterns not parsed properly")
		t.Errorf("Input: %s, result: %s", input, ret)
	}

	input = "host:192.168.12.1"
	ret, _ = backend.ParseNodePatterns(input)
	if ret["host"] != "192.168.12.1" {
		t.Error("node patterns not parsed properly")
		t.Errorf("Input: %s, result: %s", input, ret)
	}

	input = "host"
	_, err := backend.ParseNodePatterns(input)
	if err == nil {
		t.Error("node patterns should return error if input invalid")
		t.Errorf("Input: %s, result: %s", input, ret)
	}
}

func TestParseAllowedLogins(t *testing.T) {
	input := "ubuntu"
	ret, _ := backend.ParseAllowedLogins(input)
	if (len(ret) != 1) || (ret[0] != "ubuntu") {
		t.Errorf("Failed to to parse logins")
		t.Errorf("input: %s, result: %s", input, ret)
	}

	input = "intern,dev"
	ret, _ = backend.ParseAllowedLogins(input)
	if (len(ret) != 2) || (ret[0] != "intern") || (ret[1] != "dev") {
		t.Errorf("Failed to to parse logins")
		t.Errorf("input: %s, result: %s", input, ret)
	}
}

func TestGetUsersByRole_shouldReturnItsUser(t *testing.T) {
	roleName := "test-role-" + strconv.Itoa(rand.Int())
	_, _ = client.NewRole(roleName, "ubuntu", "env:production")

	users, _ := backend.GetUsersByRole(roleName)
	assert.Equal(t, len(users), 0)

	backend.AttachRole(roleName, []string{"beni", "hulk"})
	users, _ = backend.GetUsersByRole(roleName)
	assert.Equal(t, len(users), 2)

	_, _ = backend.DettachRole(roleName, []string{"hulk"})
	users, _ = backend.GetUsersByRole(roleName)
	assert.Equal(t, len(users), 1)
	assert.Equal(t, users[0].Name, "beni")
}
