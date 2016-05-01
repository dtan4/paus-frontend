package main

import (
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
)

func CreateUser(etcd *Etcd, username string) error {
	err := etcd.Mkdir("/paus/users/" + username)

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to create user. username: %s", username))
	}

	return nil
}

func UserExists(etcd *Etcd, username string) bool {
	return etcd.HasKey("/paus/users/" + username)
}

func UploadPublicKey(username, pubKey string) (string, error) {
	// libcompose does not support `docker-compose run`...
	out, err := exec.Command("docker-compose", "-p", "paus", "run", "--rm", "gitreceive-upload-key", username, pubKey).Output()

	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to upload SSH public key. username: %s, pubKey: %s", username, pubKey))
	}

	return string(out), nil
}
