package app

import (
	"fmt"
	"strings"

	"github.com/dtan4/paus-frontend/store"
	"github.com/pkg/errors"
)

func Create(etcd *store.Etcd, username, appName string) error {
	if err := etcd.Mkdir("/paus/users/" + username + "/apps/" + appName); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to create app. username: %s, appName: %s", username, appName))
	}

	for _, resource := range []string{"build-args", "envs", "deployments"} {
		if err := etcd.Mkdir("/paus/users/" + username + "/apps/" + appName + "/" + resource); err != nil {
			return errors.Wrap(
				err,
				fmt.Sprintf("Failed to create app resource. username: %s, appName: %s, resource: %s", username, appName, resource),
			)
		}
	}

	return nil
}

func Exists(etcd *store.Etcd, username, appName string) bool {
	return etcd.HasKey("/paus/users/" + username + "/apps/" + appName)
}

func List(etcd *store.Etcd, username string) ([]string, error) {
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

func URL(uriScheme, identifier, baseDomain string) string {
	return strings.ToLower(uriScheme + "://" + identifier + "." + baseDomain)
}

func URLs(etcd *store.Etcd, uriScheme, baseDomain, username, appName string) ([]string, error) {
	deployments, err := etcd.List("/paus/users/"+username+"/apps/"+appName+"/deployments/", true)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to list up app URLs. username: %s, appName: %s", username, appName))
	}

	result := make([]string, 0)

	for _, deployment := range deployments {
		revision, err := etcd.Get(deployment)

		if err != nil {
			return nil, errors.Wrap(err, "Failed to list up URL.")
		}

		identifier := username + "-" + appName + "-" + revision[0:8]
		result = append(result, URL(uriScheme, identifier, baseDomain))
	}

	return result, nil
}

func LatestAppURLOfUser(uriScheme, baseDomain, username, appName string) string {
	identifier := username + "-" + appName

	return URL(uriScheme, identifier, baseDomain)
}
