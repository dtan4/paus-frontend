package main

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

const (
	DotenvLineRegexp = `\A([\w\.]+)(?:\s*=\s*|:\s+?)([^#\n]+)?(?:\s*\#.*)?\z`
)

var (
	DotenvLine = regexp.MustCompile(DotenvLineRegexp)
)

func AddEnvironmentVariable(etcd *Etcd, username, appName, key, value string) error {
	err := etcd.Set("/paus/users/"+username+"/"+appName+"/envs/"+key, value)

	return err
}

func EnvironmentVariables(etcd *Etcd, username, appName string) (*map[string]string, error) {
	envs, err := etcd.ListWithValues("/paus/users/"+username+"/"+appName+"/envs/", true)

	if err != nil {
		return nil, err
	}

	result := map[string]string{}

	for key, value := range *envs {
		envKey := strings.Replace(key, "/paus/users/"+username+"/"+appName+"/envs/", "", 1)
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
		err := AddEnvironmentVariable(etcd, username, appName, key, value)

		if err != nil {
			return err
		}
	}

	return nil
}
