package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

func AddBuildArg(etcd *Etcd, username, appName, key, value string) error {
	err := etcd.Set("/paus/users/"+username+"/apps/"+appName+"/build-args/"+key, value)

	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("Failed to add build arg. username: %s, appName: %s, key: %s, value: %s", username, appName, key, value),
		)
	}

	return nil
}

func BuildArgs(etcd *Etcd, username, appName string) (*map[string]string, error) {
	envs, err := etcd.ListWithValues("/paus/users/"+username+"/apps/"+appName+"/build-args/", true)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to list build args. username: %s, appName: %s", username, appName))
	}

	result := map[string]string{}

	for key, value := range *envs {
		envKey := strings.Replace(key, "/paus/users/"+username+"/apps/"+appName+"/build-args/", "", 1)
		result[envKey] = value
	}

	return &result, nil
}
