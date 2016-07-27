package healthcheck

import (
	"encoding/json"

	"github.com/dtan4/paus-frontend/store"
)

type Healthcheck struct {
	Path     string
	Interval int
	MaxTry   int
}

func etcdKey(username, appName string) string {
	return "/paus/users/" + username + "/apps/" + appName + "/healthcheck"
}

func NewHealthcheck(path string, interval, maxTry int) *Healthcheck {
	return &Healthcheck{
		Path:     path,
		Interval: interval,
		MaxTry:   maxTry,
	}
}

func Create(etcd *store.Etcd, username, appName string, hc *Healthcheck) error {
	return Update(etcd, username, appName, hc)
}

func Get(etcd *store.Etcd, username, appName string) (*Healthcheck, error) {
	val, err := etcd.Get(etcdKey(username, appName))
	if err != nil {
		return nil, err
	}

	var hc Healthcheck

	if err := json.Unmarshal([]byte(val), &hc); err != nil {
		return nil, err
	}

	return &hc, nil
}

func Update(etcd *store.Etcd, username, appName string, hc *Healthcheck) error {
	b, err := json.Marshal(*hc)
	if err != nil {
		return err
	}

	if err := etcd.Set(etcdKey(username, appName), string(b)); err != nil {
		return err
	}

	return nil
}
