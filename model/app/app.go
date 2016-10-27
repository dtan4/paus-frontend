package app

import (
	"strings"

	"github.com/dtan4/paus-frontend/aws"
	"github.com/dtan4/paus-frontend/model/healthcheck"
	"github.com/dtan4/paus-frontend/store"
)

const (
	appsTable    = "paus-apps"
	userAppIndex = "user-app-index"

	defaultHealthcheckPath     = "/"
	defaultHealthcheckInterval = 5
	defaultHealthcheckMaxTry   = 10
)

// Create creates new app
func Create(username, appName string) error {
	dynamodb := aws.NewDynamoDB()

	if err := dynamodb.Update(appsTable, map[string]string{
		"user": username,
		"app":  appName,
	}); err != nil {
		return err
	}

	hc := healthcheck.NewHealthcheck(defaultHealthcheckPath, defaultHealthcheckInterval, defaultHealthcheckMaxTry)

	if err := healthcheck.Create(username, appName, hc); err != nil {
		return err
	}

	return nil
}

// Exists return whether the given app exists or not
func Exists(username, appName string) bool {
	dynamodb := aws.NewDynamoDB()

	items, err := dynamodb.Select(appsTable, userAppIndex, map[string]string{
		"user": username,
		"app":  appName,
	})
	if err != nil {
		return false
	}

	return len(items) > 0
}

// List return all apps owned by the given user
func List(username string) ([]string, error) {
	dynamodb := aws.NewDynamoDB()

	items, err := dynamodb.Select(appsTable, "", map[string]string{
		"user": username,
	})
	if err != nil {
		return []string{}, nil
	}

	result := []string{}

	for _, attrValue := range items {
		result = append(result, *attrValue["app"].S)
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
