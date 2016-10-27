package deployment

import (
	"github.com/dtan4/paus-frontend/aws"
)

const (
	deploymentsTable = "paus-deployments"
	userAppIndex     = "user-app-index"
)

type Deployment struct {
	ServiceArn string
	Revision   string
	Host       string
	Port       string
}

// NewDeployment creates new Deployment object
func NewDeployment(serviceArn, revision, host, port string) *Deployment {
	return &Deployment{
		ServiceArn: serviceArn,
		Revision:   revision,
		Host:       host,
		Port:       port,
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

	var serviceArn, revision, host, port string

	for _, attrValue := range items {
		serviceArn = *attrValue["service-arn"].S
		revision = *attrValue["revision"].S
		host = *attrValue["host"].S
		port = *attrValue["port"].S
		deployments = append(deployments, NewDeployment(serviceArn, revision, host, port))
	}

	return deployments, nil
}
