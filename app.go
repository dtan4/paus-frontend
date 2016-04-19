package main

import (
	"strings"
)

func Apps(etcd *Etcd, username string) ([]string, error) {
	apps, err := etcd.List("/paus/users/"+username+"/", true)

	if err != nil {
		return nil, err
	}

	result := make([]string, 0)

	for _, app := range apps {
		appName := strings.Replace(app, "/paus/users/"+username+"/", "", 1)
		result = append(result, appName)
	}

	return result, nil
}

func AppURL(uriScheme, identifier, baseDomain string) string {
	return uriScheme + "://" + identifier + "." + baseDomain
}

func AppURLs(etcd *Etcd, uriScheme, baseDomain, username, appName string) ([]string, error) {
	revisions, err := etcd.List("/paus/users/"+username+"/"+appName+"/revisions/", true)

	if err != nil {
		return nil, err
	}

	result := make([]string, 0)

	for _, revision := range revisions {
		revision := strings.Replace(revision, "/paus/users/"+username+"/"+appName+"/revisions/", "", 1)
		identifier := username + "-" + appName + "-" + revision
		result = append(result, AppURL(uriScheme, identifier, baseDomain))
	}

	return result, nil
}

func LatestAppURLOfUser(uriScheme, baseDomain, username, appName string) string {
	identifier := username + "-" + appName

	return AppURL(uriScheme, identifier, baseDomain)
}