package main

import (
	"fmt"
	"os/exec"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

func CreateUser(etcd *Etcd, user *github.User) error {
	username := *user.Login

	if err := etcd.Mkdir("/paus/users/" + username); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to create user. username: %s", username))
	}

	if err := etcd.Set("/paus/users/"+username+"/avater_url", *user.AvatarURL); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to set avater URL. username: %s", username))
	}

	if err := etcd.Mkdir("/paus/users/" + username + "/apps"); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to create user app directory. username: %s", username))
	}

	return nil
}

func GetAvaterURL(etcd *Etcd, username string) string {
	avaterURL, _ := etcd.Get("/paus/users/" + username + "/avater_url")

	return avaterURL
}

func GetLoginUser(etcd *Etcd, accessToken string) string {
	username, _ := etcd.Get("/paus/sessions/" + accessToken)

	return username
}

func RegisterAccessToken(etcd *Etcd, username, accessToken string) error {
	return etcd.Set("/paus/sessions/"+accessToken, username)
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
