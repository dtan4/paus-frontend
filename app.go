package main

import (
	"strings"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

func Apps(keysAPI client.KeysAPI, username string) ([]string, error) {
	resp, err := keysAPI.Get(context.Background(), "/paus/users/"+username+"/", &client.GetOptions{Sort: true})

	if err != nil {
		return nil, err
	}

	result := make([]string, 0)

	for _, node := range resp.Node.Nodes {
		appName := strings.Replace(node.Key, "/paus/users/"+username+"/", "", 1)
		result = append(result, appName)
	}

	return result, nil
}

func AppURL(uriScheme, identifier, baseDomain string) string {
	return uriScheme + "://" + identifier + "." + baseDomain
}

func AppURLs(keysAPI client.KeysAPI, uriScheme, baseDomain, username, appName string) ([]string, error) {
	resp, err := keysAPI.Get(context.Background(), "/paus/users/"+username+"/"+appName+"/revisions/", &client.GetOptions{Sort: true})

	if err != nil {
		return nil, err
	}

	result := make([]string, 0)

	for _, node := range resp.Node.Nodes {
		revision := strings.Replace(node.Key, "/paus/users/"+username+"/"+appName+"/revisions/", "", 1)
		identifier := username + "-" + appName + "-" + revision
		result = append(result, AppURL(uriScheme, identifier, baseDomain))
	}

	return result, nil
}
