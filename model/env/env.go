package env

import (
	"bufio"
	"io"
	"regexp"

	"github.com/dtan4/paus-frontend/aws"
)

const (
	envsTable    = "paus-envs"
	userAppIndex = "user-app-index"

	dotenvLineRegexp = `\A([\w\.]+)(?:\s*=\s*|:\s+?)([^#\n]+)?(?:\s*\#.*)?\z`
)

var (
	dotenvLine = regexp.MustCompile(dotenvLineRegexp)
)

// Create creates / creates environment variable
func Create(username, appName, key, value string) error {
	dynamodb := aws.NewDynamoDB()

	if err := dynamodb.Update(envsTable, map[string]string{
		"user":  username,
		"app":   appName,
		"key":   key,
		"value": value,
	}); err != nil {
		return err
	}

	return nil
}

// Delete deletes the given environment variable
func Delete(username, appName, key string) error {
	dynamodb := aws.NewDynamoDB()

	if err := dynamodb.Delete(envsTable, map[string]string{
		"user": username,
		"app":  appName,
		"key":  key,
	}); err != nil {
		return err
	}

	return nil
}

// List returns environment variables of given application
func List(username, appName string) (map[string]string, error) {
	dynamodb := aws.NewDynamoDB()

	items, err := dynamodb.Select(envsTable, userAppIndex, map[string]string{
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

// LoadDotenv loads environment variables from .env file
func LoadDotenv(username, appName string, dotenvFile io.Reader) error {
	scanner := bufio.NewScanner(dotenvFile)

	for scanner.Scan() {
		line := scanner.Text()
		matchResult := dotenvLine.FindStringSubmatch(line)

		if len(matchResult) == 0 {
			continue
		}

		key, value := matchResult[1], matchResult[2]

		if err := Create(username, appName, key, value); err != nil {
			return err
		}
	}

	return nil
}
