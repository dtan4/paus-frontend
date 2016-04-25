package main

import (
	"os/exec"
)

func CreateUser(etcd *Etcd, username string) error {
	return etcd.Mkdir("/paus/users/" + username)
}

func UserExists(etcd *Etcd, username string) bool {
	return etcd.HasKey("/paus/users/" + username)
}

func UploadPublicKey(username, pubKey string) (string, error) {
	// libcompose does not support `docker-compose run`...
	out, err := exec.Command("docker-compose", "-p", "paus", "run", "--rm", "gitreceive-upload-key", username, pubKey).Output()

	if err != nil {
		return "", err
	}

	return string(out), nil
}
