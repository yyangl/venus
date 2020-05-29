package discover

import "github.com/coreos/etcd/clientv3"

type Discover interface {
	Watch(prefix string) clientv3.WatchChan
	GetService(prefix string) ([]string, error)
}
