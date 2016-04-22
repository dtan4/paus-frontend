package main

import (
	"os/exec"
)

func CreateUser(etcd *Etcd, username, pubKey string) (string, error) {
	if err := etcd.Mkdir("/paus/users/" + username); err != nil {
		return "", err
	}

	// libcompose does not support `docker-compose run`...
	out, err := exec.Command("docker-compose", "-p", "paus", "run", "--rm", "gitreceive-upload-key", username, pubKey).Output()

	if err != nil {
		return "", err
	}

	return string(out), nil
}
