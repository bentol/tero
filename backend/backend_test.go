package backend

import (
	"fmt"
	"log"
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

func CreateDummyNewUserToken() (string, string) {
	random := strconv.Itoa(rand.Int())
	token := "test-token-" + random
	userName := "test-user-" + random
	jsonTemplate := fmt.Sprintf(`{
	  "token": "%s",
	  "user": {
		"name": "%s",
		"allowed_logins": [
		  "ubuntu"
		],
		"oidc_identities": null,
		"status": {
		  "is_locked": false,
		  "locked_time": "0001-01-01T00:00:00Z",
		  "lock_expires": "0001-01-01T00:00:00Z"
		},
		"expires": "0001-01-01T00:00:00Z",
		"created_by": {
		  "time": "0001-01-01T00:00:00Z",
		  "user": {
			"name": ""
		  }
		},
		"roles": null
	  },
	  "otp_key": "GBKEBZQI5OPPLRQ5",
	  "otp_qr_code": "",
	  "expires": "2017-12-31T16:04:11.506005181Z"
	}`, token, userName)

	path := fmt.Sprintf("teleport/addusertokens/%s", token)
	err := backend.GetStorage().InsertItem(path, jsonTemplate, 3600*1000*1000*1000)
	if err == nil {
		return token, userName
	} else {
		log.Print("Cannot create user: " + err.Error())
		return "", ""
	}
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

func TestConfigureNewUserToken_shouldFailIfTokenNotExist(t *testing.T) {
	err := backend.ConfigureNewUserToken("token", []string{"nakama"})
	assert.NotNil(t, err)
}

func TestConfigureNewUserToken_shouldAttachTheRoles(t *testing.T) {
	token, _ := CreateDummyNewUserToken()
	if token == "" {
		t.Error("Cannot create dummy token")
		t.Fail()
	}

	err := backend.ConfigureNewUserToken(token, []string{"nakama"})
	assert.Nil(t, err)

	addUserToken, err := backend.GetStorage().GetAddUserToken(token)
	assert.Nil(t, err)
	assert.EqualValues(t, addUserToken.GetStringRoles(), []string{"nakama"})
}

func TestGetAddUserTokenByUserName_shouldReturnAppropriateItem(t *testing.T) {
	token, userName := CreateDummyNewUserToken()

	addUserToken, err := backend.GetStorage().GetAddUserTokenByUserName(userName)
	assert.Nil(t, err)
	assert.Equal(t, token, addUserToken.Token)
}
