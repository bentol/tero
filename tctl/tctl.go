package tctl

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
)

func CmdAddUser(name, allowedLogins string) (stdout, token string, err error) {
	out, err := exec.Command("/usr/local/bin/tctl", "users", "add", name, "allowedLogins").CombinedOutput()
	if err != nil {
		fmt.Printf("err = %+v\n", err)
		return "", "", err
	}

	r, _ := regexp.Compile("/web/newuser/(\\w{28,35})")

	token = string(r.FindAllSubmatch(out, 2)[0][1])
	stdout = string(out)
	return stdout, token, err
}

func CmdDeleteUser(userName string) (string, error) {
	out, err := exec.Command("/usr/local/bin/tctl", "users", "rm", userName).CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}

	return string(out), err
}
