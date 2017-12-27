package dynamo

import (
	"fmt"
	"log"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/bentol/tele/role"
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
		Region: aws.String("ap-southeast-1")},
	)
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

func (dyn DynamoStorage) UpdateRole(name, allowedLogins []string, nodePatterns map[string]string) error {
	return nil
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
