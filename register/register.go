package register

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"log"
	"time"
)

type Register struct {
	cli           *clientv3.Client // etcd client
	leaseID       clientv3.LeaseID // 租约ID
	keepaliveChan <-chan *clientv3.LeaseKeepAliveResponse
	cancelFunc    func()
	lease         clientv3.Lease
	//key           string
}

func NewRegister(endpoints []string, timeNum int64) (*Register, error) {
	conf := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	}

	client, err := clientv3.New(conf)

	if err != nil {
		return nil, err
	}
	lease := clientv3.NewLease(client)

	leaseResp, err := lease.Grant(context.TODO(), timeNum)
	if err != nil {
		return nil, err
	}
	//设置续租
	ctx, cancelFunc := context.WithCancel(context.TODO())

	keepaliveChan, err := lease.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return nil, err
	}
	r := &Register{
		cli:           client,
		leaseID:       leaseResp.ID,
		keepaliveChan: keepaliveChan,
		cancelFunc:    cancelFunc,
		lease:         lease,
		//key:           "",
	}

	go r.listenLeaseRespChan()
	return r, nil
}

//监听 续租情况
func (r *Register) listenLeaseRespChan() {
	for {
		select {
		case leaseKeepResp := <-r.keepaliveChan:
			if leaseKeepResp == nil {
				log.Printf("close keepaliveChan")
				return
			}
		}
	}
}

func (r *Register) Put(key, val string) error {
	kv := clientv3.NewKV(r.cli)
	_, err := kv.Put(context.TODO(), key, val, clientv3.WithLease(r.leaseID))
	return err
}

//撤销租约
func (r *Register) RevokeLease() error {
	r.cancelFunc()
	time.Sleep(2 * time.Second)
	_, err := r.lease.Revoke(context.TODO(), r.leaseID)
	log.Printf("revoke register")
	return err
}
