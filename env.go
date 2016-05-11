package main

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const (
	DotenvLineRegexp = `\A([\w\.]+)(?:\s*=\s*|:\s+?)([^#\n]+)?(?:\s*\#.*)?\z`
)

var (
	DotenvLine = regexp.MustCompile(DotenvLineRegexp)
)

func AddEnvironmentVariable(etcd *Etcd, username, appName, key, value string) error {
	if err := etcd.Set("/paus/users/"+username+"/apps/"+appName+"/envs/"+key, value); err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("Failed to add environment variable. username: %s, appName: %s, key: %s, value: %s", username, appName, key, value),
		)
	}

	return nil
}

func DeleteEnvironmentVariable(etcd *Etcd, username, appName, key string) error {
	if err := etcd.Delete("/paus/users/" + username + "/apps/" + appName + "/envs/" + key); err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("Failed to delete environment variable. username: %s, appName: %s, key: %s", username, appName, key),
		)
	}

	return nil
}

func EnvironmentVariables(etcd *Etcd, username, appName string) (*map[string]string, error) {
	envs, err := etcd.ListWithValues("/paus/users/"+username+"/apps/"+appName+"/envs/", true)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to list environment variables. username: %s, appName: %s", username, appName))
	}

	result := map[string]string{}

	for key, value := range *envs {
		envKey := strings.Replace(key, "/paus/users/"+username+"/apps/"+appName+"/envs/", "", 1)
		result[envKey] = value
	}

	return &result, nil
}

func LoadDotenv(etcd *Etcd, username, appName string, dotenvFile io.Reader) error {
	scanner := bufio.NewScanner(dotenvFile)

	for scanner.Scan() {
		line := scanner.Text()
		matchResult := DotenvLine.FindStringSubmatch(line)

		if len(matchResult) == 0 {
			continue
		}

		key, value := matchResult[1], matchResult[2]

		if err := AddEnvironmentVariable(etcd, username, appName, key, value); err != nil {
			return errors.Wrap(err, "Failed to load dotenv.")
		}
	}

	return nil
}
