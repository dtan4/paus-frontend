package user

import (
	"fmt"
	"os/exec"

	"github.com/dtan4/paus-frontend/aws"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

const (
	usersTable = "paus-users"
)

// Create creates new user
func Create(user *github.User) error {
	dynamodb := aws.NewDynamoDB()

	if err := dynamodb.Update(usersTable, map[string]string{
		"user":       *user.Login,
		"avater-url": *user.AvatarURL,
	}); err != nil {
		return err
	}

	return nil
}

// Exists returns whether the given user exists or not
func Exists(username string) bool {
	dynamodb := aws.NewDynamoDB()

	items, err := dynamodb.Select(usersTable, "", map[string]string{
		"user": username,
	})
	if err != nil {
		return false
	}

	return len(items) > 0
}

// GetAvaterURL returns avater URL of the given user
func GetAvaterURL(username string) (string, error) {
	dynamodb := aws.NewDynamoDB()

	items, err := dynamodb.Select(usersTable, "", map[string]string{
		"user": username,
	})
	if err != nil {
		return "", err
	}

	return *items[0]["avater-url"].S, nil
}

func UploadPublicKey(username, pubKey string) (string, error) {
	// libcompose does not support `docker-compose run`...
	out, err := exec.Command("docker-compose", "-p", "paus", "run", "--rm", "gitreceive-upload-key", username, pubKey).Output()

	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to upload SSH public key. username: %s, pubKey: %s", username, pubKey))
	}

	return string(out), nil
}
