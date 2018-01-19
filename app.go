package main

import (
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/bentol/tero/backend"
	"github.com/bentol/tero/client"
	"github.com/bentol/tero/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

const configfile = "/etc/tero.toml"

var (
	app = kingpin.New("Tele", "Roles management for teleport.")

	users = kingpin.Command("users", "Manage users")

	addUser        = users.Command("add", "Add user")
	addUserName    = addUser.Arg("name", "User name").Required().String()
	addUserRoles   = addUser.Flag("roles", "The name roles of this user allowed to use. Ex: intern,dba").Required().String()
	addUserEmailTo = addUser.Flag("email", "Send registration token to, default: <username>@tokopedia.com").String()

	listUsers = users.Command("ls", "List user")

	lockUser     = users.Command("lock", "Lock user")
	lockUserName = lockUser.Arg("name", "User name").Required().String()

	unlockUser     = users.Command("unlock", "Unlock user")
	unlockUserName = unlockUser.Arg("name", "User name").Required().String()

	deleteUser     = users.Command("delete", "Delete user")
	deleteUserName = deleteUser.Arg("name", "User name").Required().String()

	resetUser        = users.Command("reset", "Reset user (delete it, then send registration link again)")
	resetUserName    = resetUser.Arg("name", "User name").Required().String()
	resetUserEmailTo = resetUser.Flag("email", "Send registration token to, default: <username>@tokopedia.com").String()

	showUser     = users.Command("show", "Show user info")
	showUserName = showUser.Arg("name", "User name").Required().String()

	roles = kingpin.Command("roles", "Manage roles")

	showRole     = roles.Command("show", "Show role info")
	showRoleName = showRole.Arg("name", "Role name").Required().String()

	addRole     = roles.Command("add", "Add role")
	addRoleName = addRole.Arg("name", "Role name").Required().String()
	rolesUsers  = addRole.Flag("logins", "The name of user this roles allowed to use. Ex: root,ubuntu").Required().String()
	rolesNodes  = addRole.Flag("nodes", "Node pattern this roles can login to. Ex: env:staging,app:postgres").Required().String()

	updateRole       = roles.Command("update", "Update role")
	updateRoleName   = updateRole.Arg("name", "Role name").Required().String()
	updateRolesUsers = updateRole.Flag("logins", "The name of user this roles allowed to use. Ex: root,ubuntu").Required().String()
	updateRolesNodes = updateRole.Flag("nodes", "Node pattern this roles can login to. Ex: env:staging,app:postgres").Required().String()

	deleteRole      = roles.Command("delete", "Delete role")
	deletedRoleName = deleteRole.Arg("role", "Role to be deleted").Required().String()

	attachRole      = kingpin.Command("attach", "Attach role to user(s)")
	attachRoleName  = attachRole.Arg("role", "Role name to be attached").Required().String()
	attachRoleUsers = attachRole.Flag("users", "User to be attached. If more than user use comma separated. Ex: adi,budi").Required().String()

	detachRole       = kingpin.Command("detach", "Detach role from user(s)")
	dettachRoleName  = detachRole.Arg("role", "Role name to be detached").Required().String()
	dettachRoleUsers = detachRole.Flag("users", "User to be detached. If more than user use comma separated. Ex: adi,budi").Required().String()

	listRole = roles.Command("ls", "List all role")
)

func init() {
	kingpin.Version("0.0.1")

	selectedStorage := "dynamodb"
	backend.InitBackend(selectedStorage)

	// read config
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file is missing: ", configfile)
	}

	var conf config.Config
	if _, err := toml.DecodeFile(configfile, &conf); err != nil {
		log.Fatal(err)
	}

	config.Set(conf)
}

func main() {
	switch kingpin.Parse() {
	case "roles add":
		out, err := client.NewRole(*addRoleName, *rolesUsers, *rolesNodes)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		fmt.Printf(out + "\n")
	case "roles update":
		out, err := client.UpdateRole(*updateRoleName, *updateRolesUsers, *updateRolesNodes)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		fmt.Printf(out + "\n")
	case "roles ls":
		out, err := client.ListRoles()
		if err != nil {
			panic("Failed to get roles")
		}
		fmt.Printf(out)
	case "roles delete":
		out, err := client.DeleteRole(*deletedRoleName)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		fmt.Printf(out + "\n")
	case "attach":
		out, err := client.AttachRole(*attachRoleName, *attachRoleUsers)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		fmt.Printf(out + "\n")
	case "detach":
		out, err := client.DetachRole(*dettachRoleName, *dettachRoleUsers)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		fmt.Printf(out + "\n")
	case "roles show":
		out, err := client.ShowRole(*showRoleName)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		fmt.Printf(out + "\n")
	case "users show":
		out, err := client.ShowUser(*showUserName)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		fmt.Printf(out + "\n")
	case "users ls":
		out, err := client.ListUser()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		fmt.Printf(out + "\n")
	case "users add":
		out, err := client.AddUser(*addUserName, *addUserRoles, *addUserEmailTo)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		fmt.Printf(out + "\n")
	case "users lock":
		out, err := client.LockUser(*lockUserName)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		fmt.Printf(out + "\n")
	case "users unlock":
		out, err := client.UnlockUser(*unlockUserName)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		fmt.Printf(out + "\n")
	case "users delete":
		fmt.Print("This command will delete user.\nAre you sure ? ")
		if askForConfirmation() != true {
			return
		}
		out, err := client.DeleteUser(*deleteUserName)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		fmt.Printf(out + "\n")
	case "users reset":
		fmt.Print("This command will reset user.\nAre you sure ? ")
		if askForConfirmation() != true {
			return
		}
		out, err := client.ResetUser(*resetUserName, *resetUserEmailTo)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		fmt.Printf(out + "\n")
	default:
		fmt.Printf("Error: Unreconized command\n")
		return
	}
}

func askForConfirmation() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return askForConfirmation()
	}
}

func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}
