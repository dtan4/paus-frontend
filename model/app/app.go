package app

import (
	"strings"

	"github.com/dtan4/paus-frontend/model/healthcheck"
	"github.com/dtan4/paus-frontend/store"
)

const (
	defaultHealthcheckPath     = "/"
	defaultHealthcheckInterval = 5
	defaultHealthcheckMaxTry   = 10
)

func Create(etcd *store.Etcd, username, appName string) error {
	appKey := "/paus/users/" + username + "/apps/" + appName

	if err := etcd.Mkdir(appKey); err != nil {
		return err
	}

	for _, resource := range []string{"build-args", "envs", "deployments", "healthcheck"} {
		if err := etcd.Mkdir(appKey + "/" + resource); err != nil {
			return err
		}
	}

	hc := healthcheck.NewHealthcheck(defaultHealthcheckPath, defaultHealthcheckInterval, defaultHealthcheckMaxTry)

	if err := healthcheck.Create(etcd, username, appName, hc); err != nil {
		return err
	}

	return nil
}

func Exists(etcd *store.Etcd, username, appName string) bool {
	return etcd.HasKey("/paus/users/" + username + "/apps/" + appName)
}

func List(etcd *store.Etcd, username string) ([]string, error) {
	apps, err := etcd.List("/paus/users/"+username+"/apps/", true)

	if err != nil {
		return nil, err
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
		return nil, err
	}

	result := make([]string, 0)

	for _, deployment := range deployments {
		revision, err := etcd.Get(deployment)

		if err != nil {
			return nil, err
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
