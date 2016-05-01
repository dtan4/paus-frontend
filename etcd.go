package main

import (
	"fmt"

	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type Etcd struct {
	keysAPI client.KeysAPI
}

func NewEtcd(etcdEndpoint string) (*Etcd, error) {
	config := client.Config{
		Endpoints: []string{etcdEndpoint},
		Transport: client.DefaultTransport,
	}

	c, err := client.New(config)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to create etcd client.")
	}

	keysAPI := client.NewKeysAPI(c)

	return &Etcd{keysAPI}, nil
}

func (c *Etcd) Get(key string) (string, error) {
	resp, err := c.keysAPI.Get(context.Background(), key, &client.GetOptions{})

	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to get etcd value. key: %s", key))
	}

	return resp.Node.Value, nil
}

func (c *Etcd) HasKey(key string) bool {
	_, err := c.keysAPI.Get(context.Background(), key, &client.GetOptions{})

	return err == nil
}

func (c *Etcd) List(key string, recursive bool) ([]string, error) {
	result := []string{}

	resp, err := c.keysAPI.Get(context.Background(), key, &client.GetOptions{Recursive: recursive})

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to list up etcd keys. key: %s, recursive: %v", key, recursive))
	}

	for _, node := range resp.Node.Nodes {
		result = append(result, node.Key)
	}

	return result, nil
}

func (c *Etcd) ListWithValues(key string, recursive bool) (*map[string]string, error) {
	result := map[string]string{}

	resp, err := c.keysAPI.Get(context.Background(), key, &client.GetOptions{Recursive: recursive})

	if err != nil {
		return &result, errors.Wrap(err, fmt.Sprintf("Failed to list up etcd keys with value. key: %s, recursive: %v", key, recursive))
	}

	for _, node := range resp.Node.Nodes {
		result[node.Key] = node.Value
	}

	return &result, nil
}

func (c *Etcd) Mkdir(key string) error {
	_, err := c.keysAPI.Set(context.Background(), key, "", &client.SetOptions{Dir: true})

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to create etcd directory. key: %s", key))
	}

	return nil
}

func (c *Etcd) Set(key, value string) error {
	_, err := c.keysAPI.Set(context.Background(), key, value, &client.SetOptions{})

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to set etcd value. key: %s, value: %s", key, value))
	}

	return nil
}
