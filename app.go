package main

import (
	"fmt"

	"github.com/bentol/tele/backend"
	"github.com/bentol/tele/client"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("Tele", "Roles management for teleport.")

	roles      = kingpin.Command("roles", "Manage roles")
	addRole    = roles.Command("add", "Add role")
	rolesNodes = addRole.Arg("nodes_pattern", "Node pattern this roles can login to. Ex: env:staging,app:tome").Required().String()
	rolesUsers = addRole.Arg("remote_users", "The name of user this roles allowed to use. Ex: root,ubuntu").Required().Strings()

	deleteRole      = roles.Command("delete", "Delete role")
	deletedRoleName = deleteRole.Arg("role", "Role to be deleted").Required().String()

	listRole = roles.Command("ls", "List all role")
)

func init() {
	kingpin.Version("0.0.1")

	selectedStorage := "dynamodb"
	backend.InitBackend(selectedStorage)
}

func main() {
	switch kingpin.Parse() {
	case "roles add":
		fmt.Printf("kingpin.Parse() = %+v\n", kingpin.Parse())
	case "roles ls":
		out, err := client.ListRoles()
		if err != nil {
			panic("Failed to get roles")
		}
		fmt.Printf(out)
	case "roles delete":
		out, err := client.DeleteRole(*deletedRoleName)
		if err != nil {
			panic("Failed to get roles")
		}
		fmt.Printf(out)
	}
}
