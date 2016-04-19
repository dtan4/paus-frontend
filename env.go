package main

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

const (
	DotenvLineRegexp = `\A([\w\.]+)(?:\s*=\s*|:\s+?)([^#\n]+)?(?:\s*\#.*)?\z`
)

var (
	DotenvLine = regexp.MustCompile(DotenvLineRegexp)
)

func AddEnvironmentVariable(keysAPI client.KeysAPI, username, appName, key, value string) error {
	_, err := keysAPI.Set(context.Background(), "/paus/users/"+username+"/"+appName+"/envs/"+key, value, nil)

	return err
}

func EnvironmentVariables(keysAPI client.KeysAPI, username, appName string) (*map[string]string, error) {
	resp, err := keysAPI.Get(context.Background(), "/paus/users/"+username+"/"+appName+"/envs/", &client.GetOptions{Sort: true})

	if err != nil {
		return nil, err
	}

	result := map[string]string{}

	for _, node := range resp.Node.Nodes {
		key := strings.Replace(node.Key, "/paus/users/"+username+"/"+appName+"/envs/", "", 1)
		value := node.Value
		result[key] = value
	}

	return &result, nil
}

func LoadDotenv(keysAPI client.KeysAPI, username, appName string, dotenvFile io.Reader) error {
	scanner := bufio.NewScanner(dotenvFile)

	for scanner.Scan() {
		line := scanner.Text()
		matchResult := DotenvLine.FindStringSubmatch(line)

		if len(matchResult) == 0 {
			continue
		}

		key, value := matchResult[1], matchResult[2]
		err := AddEnvironmentVariable(keysAPI, username, appName, key, value)

		if err != nil {
			return err
		}
	}

	return nil
}
