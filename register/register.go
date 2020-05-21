package register

import (
	"go.etcd.io/etcd/clientv3"
)

type Register struct {
	cli *clientv3.Client
}
