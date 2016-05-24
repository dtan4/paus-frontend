package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

func AppExists(etcd *Etcd, username, appName string) bool {
	return etcd.HasKey("/paus/users/" + username + "/apps/" + appName)
}

func Apps(etcd *Etcd, username string) ([]string, error) {
	apps, err := etcd.List("/paus/users/"+username+"/apps/", true)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to list up apps. username: %s", username))
	}

	result := make([]string, 0)

	for _, app := range apps {
		appName := strings.Replace(app, "/paus/users/"+username+"/apps/", "", 1)
		result = append(result, appName)
	}

	return result, nil
}

func AppURL(uriScheme, identifier, baseDomain string) string {
	return strings.ToLower(uriScheme + "://" + identifier + "." + baseDomain)
}

func AppURLs(etcd *Etcd, uriScheme, baseDomain, username, appName string) ([]string, error) {
	revisions, err := etcd.List("/paus/users/"+username+"/apps/"+appName+"/revisions/", true)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to list up app URLs. username: %s, appName: %s", username, appName))
	}

	result := make([]string, 0)

	for _, revision := range revisions {
		revision := strings.Replace(revision, "/paus/users/"+username+"/apps/"+appName+"/revisions/", "", 1)
		identifier := username + "-" + appName + "-" + revision[0:8]
		result = append(result, AppURL(uriScheme, identifier, baseDomain))
	}

	return result, nil
}

func CreateApp(etcd *Etcd, username, appName string) error {
	if err := etcd.Mkdir("/paus/users/" + username + "/apps/" + appName); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to create app. username: %s, appName: %s", username, appName))
	}

	for _, resource := range []string{"build-args", "envs", "revisions"} {
		if err := etcd.Mkdir("/paus/users/" + username + "/apps/" + appName + "/" + resource); err != nil {
			return errors.Wrap(
				err,
				fmt.Sprintf("Failed to create app resource. username: %s, appName: %s, resource: %s", username, appName, resource),
			)
		}
	}

	return nil
}

func LatestAppURLOfUser(uriScheme, baseDomain, username, appName string) string {
	identifier := username + "-" + appName

	return AppURL(uriScheme, identifier, baseDomain)
}
