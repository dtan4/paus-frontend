package healthcheck

import (
	"strconv"

	"github.com/dtan4/paus-frontend/aws"
)

const (
	healthchecksTable = "paus-healthchecks"
	userAppIndex      = "user-app-index"
)

type Healthcheck struct {
	Path     string
	Interval int
	MaxTry   int
}

// NewHealthCheck creates new healthcheck object
func NewHealthcheck(path string, interval, maxTry int) *Healthcheck {
	return &Healthcheck{
		Path:     path,
		Interval: interval,
		MaxTry:   maxTry,
	}
}

// Create creates new healthcheck object of the given application
func Create(username, appName string, hc *Healthcheck) error {
	return Update(username, appName, hc)
}

// Get retrieves healthcheck of the given applciation
func Get(username, appName string) (*Healthcheck, error) {
	dynamodb := aws.NewDynamoDB()

	items, err := dynamodb.Select(healthchecksTable, userAppIndex, map[string]string{
		"user": username,
		"app":  appName,
	})
	if err != nil {
		return nil, err
	}

	hc := items[0]
	path := *hc["path"].S

	interval, err := strconv.Atoi(*hc["interval"].N)
	if err != nil {
		return nil, err
	}

	maxTry, err := strconv.Atoi(*hc["max-try"].N)
	if err != nil {
		return nil, err
	}

	return NewHealthcheck(path, interval, maxTry), nil
}

// Update updates / creates healthcheck of the given application
func Update(username, appName string, hc *Healthcheck) error {
	dynamodb := aws.NewDynamoDB()

	if err := dynamodb.Update(healthchecksTable, map[string]string{
		"user":     username,
		"app":      appName,
		"path":     hc.Path,
		"interval": strconv.Itoa(hc.Interval),
		"max-try":  strconv.Itoa(hc.MaxTry),
	}); err != nil {
		return err
	}

	return nil
}
