package arg

import (
	"strings"

	"github.com/dtan4/paus-frontend/store"
)

func Create(etcd *store.Etcd, username, appName, key, value string) error {
	if err := etcd.Set("/paus/users/"+username+"/apps/"+appName+"/build-args/"+key, value); err != nil {
		return err
	}

	return nil
}

func Delete(etcd *store.Etcd, username, appName, key string) error {
	if err := etcd.Delete("/paus/users/" + username + "/apps/" + appName + "/build-args/" + key); err != nil {
		return err
	}

	return nil
}

func List(etcd *store.Etcd, username, appName string) (*map[string]string, error) {
	envs, err := etcd.ListWithValues("/paus/users/"+username+"/apps/"+appName+"/build-args/", true)

	if err != nil {
		return nil, err
	}

	result := map[string]string{}

	for key, value := range *envs {
		envKey := strings.Replace(key, "/paus/users/"+username+"/apps/"+appName+"/build-args/", "", 1)
		result[envKey] = value
	}

	return &result, nil
}
