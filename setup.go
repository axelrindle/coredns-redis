package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/kubeapi"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	clog "github.com/coredns/coredns/plugin/pkg/log"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var log = clog.NewWithPlugin("redis")

func init() {
	caddy.RegisterPlugin("redis", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	config, err := parseConfig(c)
	if err != nil {
		return plugin.Error("redis", err)
	}

	redis := &Redis{
		config: config,
	}

	redis.Connect()
	redis.LoadZones()

	redis.setWatch(context.Background())
	c.OnStartup(startWatch(redis, dnsserver.GetConfig(c)))
	c.OnShutdown(stopWatch(redis))

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		redis.Next = next
		return redis
	})

	return nil
}

func (k *Redis) setWatch(ctx context.Context) {
	// define Node controller and reverse lookup indexer
	k.indexer, k.controller = cache.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc: func(o v1.ListOptions) (runtime.Object, error) {
				return k.client.CoreV1().Nodes().List(ctx, o)
			},
			WatchFunc: func(o v1.ListOptions) (watch.Interface, error) {
				return k.client.CoreV1().Nodes().Watch(ctx, o)
			},
		},
		&core.Node{},
		0,
		cache.ResourceEventHandlerFuncs{},
		cache.Indexers{"reverse": func(obj interface{}) ([]string, error) {
			node, ok := obj.(*core.Node)
			if !ok {
				return nil, errors.New("unexpected obj type")
			}
			var idx []string
			for _, addr := range node.Status.Addresses {
				// if addr.Type != k.ipType {
				// 	continue
				// }
				idx = append(idx, addr.Address)
			}
			return idx, nil
		}},
	)
}

func startWatch(k *Redis, config *dnsserver.Config) func() error {
	return func() error {
		// retrieve client from kubeapi plugin
		var err error
		k.client, err = kubeapi.Client(config)
		if err != nil {
			return err
		}

		// start the informer
		go k.controller.Run(k.stopCh)
		return nil
	}
}

func stopWatch(k *Redis) func() error {
	return func() error {
		k.stopLock.Lock()
		defer k.stopLock.Unlock()
		if !k.shutdown {
			close(k.stopCh)
			k.shutdown = true
			return nil
		}
		return fmt.Errorf("shutdown already in progress")
	}
}
