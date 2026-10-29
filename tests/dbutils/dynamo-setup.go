package dbutils

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/patrickisgreat/user-settings/internal/appserver"
	"github.com/patrickisgreat/user-settings/internal/config"
)

func ClearTables() {
	// Ignore error if not able to delete tables
	DeleteTables()
	err := CreateTables()
	if err != nil {
		panic(err)
	}
}

func DeleteTables() {
	_ = deleteTable(config.Config.UserSettingsTableName)
}

func CreateTables() error {
	return createTable(config.Config.UserSettingsTableName)
}

func createTable(tableName string) error {
	d := appserver.Context().DynamoDBClient
	_, err := d.CreateTable(context.Background(), &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("sk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("sk"),
				KeyType:       types.KeyTypeRange,
			},
		},
		TableClass: types.TableClassStandard,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	})

	return handleCreateTable(d, tableName, err)
}

func handleCreateTable(d *dynamodb.Client, tableName string, err error) error {
	if err != nil {
		log.Printf("Couldn't create table %v. reason: %v\n", aws.String(tableName), err)
		return err
	}
	waiter := dynamodb.NewTableExistsWaiter(d)
	err = waiter.Wait(context.Background(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName)}, 30*time.Second)
	if err != nil {
		log.Printf("Wait for table exists failed. reason: %v\n", err)
		return err
	}
	return nil
}

func deleteTable(tableName string) error {
	_, err := appserver.Context().DynamoDBClient.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName)})
	return handleDeleteTable(tableName, err)
}

func handleDeleteTable(tableName string, err error) error {
	if err != nil {
		log.Printf("Couldn't delete table %v. reason: %v\n", aws.String(tableName), err)
	}
	return err
}
