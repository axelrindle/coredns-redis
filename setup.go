package redis

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	clog "github.com/coredns/coredns/plugin/pkg/log"
)

var log = clog.NewWithPlugin("redis")

func init() {
	caddy.RegisterPlugin("redis", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(caddy *caddy.Controller) error {
	config, err := parseConfig(caddy)
	if err != nil {
		return plugin.Error("redis", err)
	}

	redis := NewRedis(config)
	redis.Connect()

	handler := &RedisHandler{
		redis: redis,
	}

	registry := dnsserver.GetConfig(caddy)

	// redis.SetWatch(context.Background())
	// caddy.OnStartup(startWatch(redis, registry))
	// caddy.OnShutdown(stopWatch(redis))

	registry.AddPlugin(func(next plugin.Handler) plugin.Handler {
		handler.next = next
		return handler
	})

	return nil
}

// func (redis *Redis) SetWatch(ctx context.Context) {
// 	// define Node controller and reverse lookup indexer
// 	redis.indexer, redis.controller = cache.NewIndexerInformer(
// 		&cache.ListWatch{
// 			ListFunc: func(o v1.ListOptions) (runtime.Object, error) {
// 				return redis.client.CoreV1().Nodes().List(ctx, o)
// 			},
// 			WatchFunc: func(o v1.ListOptions) (watch.Interface, error) {
// 				return redis.client.CoreV1().Nodes().Watch(ctx, o)
// 			},
// 		},
// 		&core.Node{},
// 		0,
// 		cache.ResourceEventHandlerFuncs{},
// 		cache.Indexers{"reverse": func(obj interface{}) ([]string, error) {
// 			node, ok := obj.(*core.Node)
// 			if !ok {
// 				return nil, errors.New("unexpected obj type")
// 			}
// 			var idx []string
// 			for _, addr := range node.Status.Addresses {
// 				// FIXME: reverse loopup
// 				// if addr.Type != k.ipType {
// 				// 	continue
// 				// }
// 				idx = append(idx, addr.Address)
// 			}
// 			return idx, nil
// 		}},
// 	)
// }

// func startWatch(redis *Redis, config *dnsserver.Config) func() error {
// 	return func() error {
// 		var err error
// 		redis.client, err = kubeapi.Client(config)
// 		if err != nil {
// 			return err
// 		}

// 		// start the informer
// 		go redis.controller.Run(redis.stopCh)
// 		return nil
// 	}
// }

// func stopWatch(redis *Redis) func() error {
// 	return func() error {
// 		redis.stopLock.Lock()
// 		defer redis.stopLock.Unlock()
// 		if !redis.shutdown {
// 			close(redis.stopCh)
// 			redis.shutdown = true
// 			return nil
// 		}
// 		return errors.New("shutdown already in progress")
// 	}
// }
