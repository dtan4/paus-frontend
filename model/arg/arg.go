package arg

import (
	"github.com/dtan4/paus-frontend/aws"
)

const (
	buildArgsTable = "paus-build-args"
	userAppIndex   = "user-app-index"
)

// Create creates / creates build arg
func Create(username, appName, key, value string) error {
	dynamodb := aws.NewDynamoDB()

	if err := dynamodb.Update(buildArgsTable, map[string]string{
		"user":  username,
		"app":   appName,
		"key":   key,
		"value": value,
	}); err != nil {
		return err
	}

	return nil
}

// Delete deletes the given build arg
func Delete(username, appName, key string) error {
	dynamodb := aws.NewDynamoDB()

	if err := dynamodb.Delete(buildArgsTable, map[string]string{
		"user": username,
		"app":  appName,
		"key":  key,
	}); err != nil {
		return err
	}

	return nil
}

// List returns build args of given application
func List(username, appName string) (map[string]string, error) {
	dynamodb := aws.NewDynamoDB()

	items, err := dynamodb.Select(buildArgsTable, userAppIndex, map[string]string{
		"user": username,
		"app":  appName,
	})
	if err != nil {
		return make(map[string]string), err
	}

	args := make(map[string]string)

	var key, value string

	for _, attrValue := range items {
		key = *attrValue["key"].S
		value = *attrValue["value"].S
		args[key] = value
	}

	return args, nil
}
