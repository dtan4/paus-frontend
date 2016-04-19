package main

import (
	"github.com/coreos/etcd/client"
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
		return nil, err
	}

	keysAPI := client.NewKeysAPI(c)

	return &Etcd{keysAPI}, nil
}

func (c *Etcd) Get(key string) (string, error) {
	resp, err := c.keysAPI.Get(context.Background(), key, &client.GetOptions{})

	if err != nil {
		return "", err
	}

	return resp.Node.Value, nil
}

func (c *Etcd) List(key string, recursive bool) ([]string, error) {
	result := []string{}

	resp, err := c.keysAPI.Get(context.Background(), key, &client.GetOptions{Recursive: recursive})

	if err != nil {
		return nil, err
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
		return &result, err
	}

	for _, node := range resp.Node.Nodes {
		result[node.Key] = node.Value
	}

	return &result, nil
}

func (c *Etcd) Set(key, value string) error {
	_, err := c.keysAPI.Set(context.Background(), key, value, &client.SetOptions{})

	return err
}
