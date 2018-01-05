package dynamo

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/bentol/tero/role"
	"github.com/bentol/tero/token"
	"github.com/bentol/tero/user"
)

var (
	tableName *string
)

type DynamoStorage struct {
	Svc *dynamodb.DynamoDB
}

type DynamoRow struct {
	TTL       int64
	HashKey   string
	Value     []byte
	FullPath  string
	Timestamp int64
}

func New() DynamoStorage {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("localhost")},
	)
	sess.Config.Endpoint = aws.String("http://localhost:4567")
	sess.Config.DisableSSL = aws.Bool(true)
	creds := credentials.NewStaticCredentials("access_key", "secret_key", "")
	sess.Config.Credentials = creds

	if err != nil {
		log.Fatal(err)
	}

	// Create DynamoDB client
	svc := dynamodb.New(sess)
	tableName = aws.String("teleport.state")

	return DynamoStorage{
		svc,
	}
}

func (dyn DynamoStorage) GetRoles() ([]role.Role, error) {
	result := make([]role.Role, 0)
	queryParams := &dynamodb.QueryInput{
		TableName: tableName,
		KeyConditions: map[string]*dynamodb.Condition{
			"HashKey": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String("teleport"),
					},
				},
			},
			"FullPath": {
				ComparisonOperator: aws.String("BEGINS_WITH"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String("teleport/roles"),
					},
				},
			},
		},
	}

	resp, err := dyn.Svc.Query(queryParams)
	if err != nil {
		return nil, err
	}

	for _, item := range resp.Items {
		result = append(result, dynItemToRole(item))
	}
	return result, nil
}

func (dyn DynamoStorage) CreateRole(name string, allowedLogins []string, nodePatterns map[string]string) (*role.Role, error) {
	svc := dyn.Svc

	jsonTemplate, _ := gabs.ParseJSON([]byte(role.RoleJsonTemplate))
	jsonTemplate.SetP(name, "metadata.name")
	jsonTemplate.SetP(allowedLogins, "spec.allow.logins")
	jsonTemplate.SetP(nodePatterns, "spec.allow.node_labels")

	row := DynamoRow{
		0,
		"teleport",
		[]byte(jsonTemplate.String()),
		fmt.Sprintf("teleport/roles/%s/params", name),
		time.Now().UnixNano() / int64(time.Second),
	}

	av, err := dynamodbattribute.MarshalMap(row)
	if err != nil {
		return nil, err
	}

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: tableName,
		Item:      av,
	})

	if err != nil {
		return nil, err
	}

	return dyn.GetRoleByName(name)
}

func (dyn DynamoStorage) DeleteRole(name string) error {
	params_del := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"FullPath": {
				S: aws.String(fmt.Sprintf("teleport/roles/%s/params", name)),
			},
			"HashKey": {
				S: aws.String("teleport"),
			},
		},
		TableName: tableName,
	}

	_, err := dyn.Svc.DeleteItem(params_del)
	if err != nil {
		return err
	}

	return nil
}

func (dyn DynamoStorage) GetRoleByName(name string) (*role.Role, error) {
	svc := dyn.Svc

	params_get := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"FullPath": {
				S: aws.String(fmt.Sprintf("teleport/roles/%s/params", name)),
			},
			"HashKey": {
				S: aws.String("teleport"),
			},
		},
		TableName: tableName,
	}
	resp, err := svc.GetItem(params_get)
	if err != nil {
		return nil, err
	}

	if len(resp.Item) == 0 {
		return nil, nil
	}

	r := dynItemToRole(resp.Item)
	return &r, nil
}

func (dyn DynamoStorage) UpdateRole(name string, allowedLogins []string, nodePatterns map[string]string) (*role.Role, error) {
	oldRole, _ := dyn.GetRoleByName(name)
	oldRole.AllowedLogins = allowedLogins
	oldRole.NodePatterns = nodePatterns

	paramsUpdate := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"FullPath": {
				S: aws.String(fmt.Sprintf("teleport/roles/%s/params", name)),
			},
			"HashKey": {
				S: aws.String("teleport"),
			},
		},
		TableName: tableName,
		ExpressionAttributeNames: map[string]*string{
			"#Value": aws.String("Value"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v": {
				B: []byte(oldRole.GetJSON()),
			},
		},
		UpdateExpression: aws.String("SET #Value = :v"),
		ReturnValues:     aws.String("ALL_NEW"),
	}

	resp, err := dyn.Svc.UpdateItem(paramsUpdate)
	if err != nil {
		return nil, err
	}

	updatedRole := dynItemToRole(resp.Attributes)
	return &updatedRole, nil
}

func (dyn DynamoStorage) DetachRole(selectedRole *role.Role, users []user.User) ([]user.User, error) {
	updatedUserRow := make([]map[string]*dynamodb.AttributeValue, 0)

	for _, user := range users {
		updatedRoles := make([]role.Role, 0)
		for _, r := range user.Roles {
			if r.Name != selectedRole.Name {
				updatedRoles = append(updatedRoles, r)
			}
		}
		user.Roles = updatedRoles

		paramsUpdate := &dynamodb.UpdateItemInput{
			Key: map[string]*dynamodb.AttributeValue{
				"FullPath": {
					S: aws.String(fmt.Sprintf("teleport/web/users/%s/params", user.Name)),
				},
				"HashKey": {
					S: aws.String("teleport"),
				},
			},
			TableName: tableName,
			ExpressionAttributeNames: map[string]*string{
				"#Value": aws.String("Value"),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":v": {
					B: []byte(user.GetJSON()),
				},
			},
			UpdateExpression: aws.String("SET #Value = :v"),
			ReturnValues:     aws.String("ALL_NEW"),
		}

		resp, err := dyn.Svc.UpdateItem(paramsUpdate)
		if err != nil {
			return nil, err
		}

		updatedUserRow = append(updatedUserRow, resp.Attributes)
	}

	allRoles, _ := dyn.GetRoles()
	mappedUsers := dynItemsToUsers(updatedUserRow, allRoles)
	result := make([]user.User, 0, len(mappedUsers))
	for _, u := range mappedUsers {
		result = append(result, u)
	}
	return result, nil
}

func (dyn DynamoStorage) AttachRole(selectedRole *role.Role, users []user.User) ([]user.User, error) {
	updatedUserRow := make([]map[string]*dynamodb.AttributeValue, 0)

	for _, user := range users {
		user.Roles = append(user.Roles, *selectedRole)

		paramsUpdate := &dynamodb.UpdateItemInput{
			Key: map[string]*dynamodb.AttributeValue{
				"FullPath": {
					S: aws.String(fmt.Sprintf("teleport/web/users/%s/params", user.Name)),
				},
				"HashKey": {
					S: aws.String("teleport"),
				},
			},
			TableName: tableName,
			ExpressionAttributeNames: map[string]*string{
				"#Value": aws.String("Value"),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":v": {
					B: []byte(user.GetJSON()),
				},
			},
			UpdateExpression: aws.String("SET #Value = :v"),
			ReturnValues:     aws.String("ALL_NEW"),
		}

		resp, err := dyn.Svc.UpdateItem(paramsUpdate)
		if err != nil {
			return nil, err
		}

		updatedUserRow = append(updatedUserRow, resp.Attributes)
	}

	allRoles, _ := dyn.GetRoles()
	mappedUsers := dynItemsToUsers(updatedUserRow, allRoles)
	result := make([]user.User, 0, len(mappedUsers))
	for _, u := range mappedUsers {
		result = append(result, u)
	}
	return result, nil
}

func (dyn DynamoStorage) GetAllUsers() ([]user.User, error) {
	queryParams := &dynamodb.QueryInput{
		TableName: tableName,
		KeyConditions: map[string]*dynamodb.Condition{
			"HashKey": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String("teleport"),
					},
				},
			},
			"FullPath": {
				ComparisonOperator: aws.String("BEGINS_WITH"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String("teleport/web/users"),
					},
				},
			},
		},
	}

	resp, err := dyn.Svc.Query(queryParams)
	if err != nil {
		return nil, err
	}
	cleanUsers := make([]map[string]*dynamodb.AttributeValue, 0)
	for _, v := range resp.Items {
		path := *v["FullPath"].S
		if strings.HasSuffix(path, "params") {
			cleanUsers = append(cleanUsers, v)
		}
	}

	allRoles, _ := dyn.GetRoles()
	allUsers := dynItemsToUsersAsArray(cleanUsers, allRoles)
	return allUsers, nil
}

func (dyn DynamoStorage) GetUsersByRole(roleName string) ([]user.User, error) {
	allUsers, err := dyn.GetAllUsers()
	if err != nil {
		return nil, err
	}
	filteredUsers := make([]user.User, 0)

	// todo: move filtering in database side
	for _, u := range allUsers {
		for _, r := range u.Roles {
			if r.Name == roleName {
				filteredUsers = append(filteredUsers, u)
				break
			}
		}
	}
	return filteredUsers, nil
}

func (dyn DynamoStorage) GetUsersByNames(names []string) ([]user.User, error) {
	filteredUsers := make([]user.User, 0)

	// todo: move filtering in database side
	queryParams := &dynamodb.QueryInput{
		TableName: tableName,
		KeyConditions: map[string]*dynamodb.Condition{
			"HashKey": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String("teleport"),
					},
				},
			},
			"FullPath": {
				ComparisonOperator: aws.String("BEGINS_WITH"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String("teleport/web/users"),
					},
				},
			},
		},
	}

	resp, err := dyn.Svc.Query(queryParams)
	cleanUsers := make([]map[string]*dynamodb.AttributeValue, 0)
	for _, v := range resp.Items {
		path := *v["FullPath"].S
		if strings.HasSuffix(path, "params") {
			cleanUsers = append(cleanUsers, v)
		}
	}
	allRoles, _ := dyn.GetRoles()
	allUsers := dynItemsToUsers(cleanUsers, allRoles)

	for _, name := range names {
		if u, ok := allUsers[name]; ok {
			filteredUsers = append(filteredUsers, u)
		}
	}

	if err != nil {
		return nil, err
	}

	return filteredUsers, nil
}

func (dyn DynamoStorage) GetAddUserToken(userToken string) (*token.AddUserToken, error) {
	params_get := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"FullPath": {
				S: aws.String(fmt.Sprintf("teleport/addusertokens/%s", userToken)),
			},
			"HashKey": {
				S: aws.String("teleport"),
			},
		},
		TableName: tableName,
	}
	resp, err := dyn.Svc.GetItem(params_get)
	if err != nil {
		return nil, err
	}

	if len(resp.Item) == 0 {
		return nil, nil
	}

	addUserToken := token.AddUserToken{
		userToken,
		resp.Item["Value"].B,
	}
	return &addUserToken, nil
}

func (dyn DynamoStorage) GetAddUserTokenByUserName(searchedUserName string) (*token.AddUserToken, error) {
	// todo: move filtering in database side
	queryParams := &dynamodb.QueryInput{
		TableName: tableName,
		KeyConditions: map[string]*dynamodb.Condition{
			"HashKey": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String("teleport"),
					},
				},
			},
			"FullPath": {
				ComparisonOperator: aws.String("BEGINS_WITH"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String("teleport/addusertokens"),
					},
				},
			},
		},
	}

	resp, err := dyn.Svc.Query(queryParams)
	if err != nil {
		return nil, err
	}

	for _, v := range resp.Items {
		json, err := gabs.ParseJSON(v["Value"].B)
		if err != nil {
			return nil, err
		}
		userName := json.Path("user.name").Data().(string)
		if searchedUserName == userName {
			addUsertoken := token.AddUserToken{
				json.Path("token").Data().(string),
				v["Value"].B,
			}
			return &addUsertoken, nil
		}
	}

	return nil, nil
}

func (dyn DynamoStorage) InsertItem(path, value string, ttl int64) error {
	svc := dyn.Svc

	row := DynamoRow{
		ttl,
		"teleport",
		[]byte(value),
		fmt.Sprintf(path),
		time.Now().UnixNano() / int64(time.Second),
	}

	av, err := dynamodbattribute.MarshalMap(row)
	if err != nil {
		return err
	}

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: tableName,
		Item:      av,
	})

	return err
}

func (dyn DynamoStorage) UpdateAddUserToken(token *token.AddUserToken) error {
	path := fmt.Sprintf("teleport/addusertokens/%s", token.Token)
	return dyn.UpdateValue(path, token.JSON)
}

func (dyn DynamoStorage) UpdateValue(path string, value []byte) error {
	paramsUpdate := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"FullPath": {
				S: aws.String(path),
			},
			"HashKey": {
				S: aws.String("teleport"),
			},
		},
		TableName: tableName,
		ExpressionAttributeNames: map[string]*string{
			"#Value": aws.String("Value"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v": {
				B: []byte(value),
			},
		},
		UpdateExpression: aws.String("SET #Value = :v"),
		ReturnValues:     aws.String("ALL_NEW"),
	}

	_, err := dyn.Svc.UpdateItem(paramsUpdate)
	return err
}

func dynItemToRole(item map[string]*dynamodb.AttributeValue) role.Role {
	obj := DynamoRow{}
	dynamodbattribute.UnmarshalMap(item, &obj)
	rawRole, _ := gabs.ParseJSON(obj.Value)
	rawNodePatterns := rawRole.Path("spec.allow.node_labels").Data().(map[string]interface{})
	rawLogins := rawRole.Path("spec.allow.logins").Data().([]interface{})

	nodePatterns := make(map[string]string)
	for k, v := range rawNodePatterns {
		nodePatterns[k] = v.(string)
	}

	logins := make([]string, 0)
	for _, v := range rawLogins {
		logins = append(logins, v.(string))
	}

	r := role.Role{
		rawRole.Path("metadata.name").Data().(string),
		nodePatterns,
		logins,
	}
	return r
}

func dynItemsToUsersAsArray(items []map[string]*dynamodb.AttributeValue, roles []role.Role) []user.User {
	users := dynItemsToUsers(items, roles)
	arrayUsers := make([]user.User, 0)
	for _, u := range users {
		arrayUsers = append(arrayUsers, u)
	}
	return arrayUsers
}
func dynItemsToUsers(items []map[string]*dynamodb.AttributeValue, roles []role.Role) map[string]user.User {
	mappedRoles := make(map[string]role.Role)
	for _, r := range roles {
		mappedRoles[r.Name] = r
	}

	result := make(map[string]user.User, 0)
	for _, item := range items {
		u := dynItemToUser(item, mappedRoles)
		result[u.Name] = u
	}
	return result
}

func dynItemToUser(item map[string]*dynamodb.AttributeValue, mappedRoles map[string]role.Role) user.User {
	rawUser, _ := gabs.ParseJSON(item["Value"].B)
	rawRoles := rawUser.Path("spec.roles").Data()

	roles := make([]role.Role, 0)
	if rawRoles != nil {
		for _, role := range rawRoles.([]interface{}) {
			roles = append(roles, mappedRoles[role.(string)])
		}
	}

	u := user.User{
		rawUser.Path("metadata.name").Data().(string),
		roles,
	}
	return u
}
