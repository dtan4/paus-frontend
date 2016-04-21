package main

import (
	"strings"
)

func AddBuildArg(etcd *Etcd, username, appName, key, value string) error {
	err := etcd.Set("/paus/users/"+username+"/"+appName+"/build-args/"+key, value)

	return err
}

func BuildArgs(etcd *Etcd, username, appName string) (*map[string]string, error) {
	envs, err := etcd.ListWithValues("/paus/users/"+username+"/"+appName+"/build-args/", true)

	if err != nil {
		return nil, err
	}

	result := map[string]string{}

	for key, value := range *envs {
		envKey := strings.Replace(key, "/paus/users/"+username+"/"+appName+"/build-args/", "", 1)
		result[envKey] = value
	}

	return &result, nil
}
