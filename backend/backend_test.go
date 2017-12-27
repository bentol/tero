package backend

import (
	"testing"

	"github.com/bentol/tele/backend"
)

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
