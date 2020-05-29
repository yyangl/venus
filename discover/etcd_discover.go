package discover

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

type etcdDiscover struct {
	cli *clientv3.Client
}

func NewEtcdDiscover(addr []string) (Discover, error) {
	conf := clientv3.Config{
		Endpoints:   addr,
		DialTimeout: 5 * time.Second,
	}
	if client, err := clientv3.New(conf); err == nil {
		return &etcdDiscover{
			cli: client,
			//serverList: make(map[string]string),
		}, nil
	} else {
		return nil, err
	}
}

func (d *etcdDiscover) GetService(prefix string) ([]string, error) {
	resp, err := d.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	if resp == nil || resp.Kvs == nil {
		return res, nil
	}
	for i := range resp.Kvs {
		//res = append(res, resp.Kvs[i].Key)
		if v := resp.Kvs[i].Value; v != nil {
			fmt.Printf(string(v))
		}
	}
	return res, nil
}

//func (d *etcdDiscover) extractAddrs(resp *clientv3.GetResponse) []string {
//	addrs := make([]string, 0)
//	if resp == nil || resp.Kvs == nil {
//		return addrs
//	}
//	for i := range resp.Kvs {
//		if v := resp.Kvs[i].Value; v != nil {
//			d.SetServiceList(string(resp.Kvs[i].Key), string(resp.Kvs[i].Value))
//			addrs = append(addrs, string(v))
//		}
//	}
//	return addrs
//}
//
//func (d *etcdDiscover) SetServiceList(key, val string) {
//	d.mux.Lock()
//	defer d.mux.Unlock()
//	d.serverList[key] = string(val)
//	log.Println("set data key :", key, "val:", val)
//}
//
//func (d *etcdDiscover) DelServiceList(key string) {
//	d.mux.Lock()
//	defer d.mux.Unlock()
//	delete(d.serverList, key)
//	log.Println("del data key:", key)
//}

func (d *etcdDiscover) Watch(prefix string) clientv3.WatchChan {
	rch := d.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	return rch
	//for wResp := range rch {
	//	for _, ev := range wResp.Events {
	//		switch ev.Type {
	//		case mvccpb.PUT:
	//			d.SetServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
	//		case mvccpb.DELETE:
	//			d.DelServiceList(string(ev.Kv.Key))
	//		}
	//	}
	//}
}
