package deployment

import (
	"github.com/dtan4/paus-frontend/aws"
)

const (
	deploymentsTable = "paus-deployments"
	userAppIndex     = "user-app-index"
)

type Deployment struct {
	Username   string
	AppName    string
	ServiceArn string
	Revision   string
}

// NewDeployment creates new Deployment object
func NewDeployment(username, appName, serviceArn, revision string) *Deployment {
	return &Deployment{
		Username:   username,
		AppName:    appName,
		ServiceArn: serviceArn,
		Revision:   revision,
	}
}

// List returns deployments of given application
func List(username, appName string) ([]*Deployment, error) {
	dynamodb := aws.NewDynamoDB()

	items, err := dynamodb.Select(deploymentsTable, userAppIndex, map[string]string{
		"user": username,
		"app":  appName,
	})
	if err != nil {
		return []*Deployment{}, err
	}

	deployments := []*Deployment{}

	var serviceArn, revision string

	for _, attrValue := range items {
		serviceArn = *attrValue["service-arn"].S
		revision = *attrValue["revision"].S
		deployments = append(deployments, NewDeployment(username, appName, serviceArn, revision))
	}

	return deployments, nil
}
