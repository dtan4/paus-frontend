package env

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/dtan4/paus-frontend/store"
)

const (
	DotenvLineRegexp = `\A([\w\.]+)(?:\s*=\s*|:\s+?)([^#\n]+)?(?:\s*\#.*)?\z`
)

var (
	DotenvLine = regexp.MustCompile(DotenvLineRegexp)
)

func Create(etcd *store.Etcd, username, appName, key, value string) error {
	if err := etcd.Set("/paus/users/"+username+"/apps/"+appName+"/envs/"+key, value); err != nil {
		return err
	}

	return nil
}

func Delete(etcd *store.Etcd, username, appName, key string) error {
	if err := etcd.Delete("/paus/users/" + username + "/apps/" + appName + "/envs/" + key); err != nil {
		return err
	}

	return nil
}

func List(etcd *store.Etcd, username, appName string) (*map[string]string, error) {
	envs, err := etcd.ListWithValues("/paus/users/"+username+"/apps/"+appName+"/envs/", true)

	if err != nil {
		return nil, err
	}

	result := map[string]string{}

	for key, value := range *envs {
		envKey := strings.Replace(key, "/paus/users/"+username+"/apps/"+appName+"/envs/", "", 1)
		result[envKey] = value
	}

	return &result, nil
}

func LoadDotenv(etcd *store.Etcd, username, appName string, dotenvFile io.Reader) error {
	scanner := bufio.NewScanner(dotenvFile)

	for scanner.Scan() {
		line := scanner.Text()
		matchResult := DotenvLine.FindStringSubmatch(line)

		if len(matchResult) == 0 {
			continue
		}

		key, value := matchResult[1], matchResult[2]

		if err := Create(etcd, username, appName, key, value); err != nil {
			return err
		}
	}

	return nil
}
