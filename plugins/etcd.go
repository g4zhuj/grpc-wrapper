package plugins

import (
	"context"
	"encoding/json"
	"time"

	"google.golang.org/grpc/grpclog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/naming"
	"google.golang.org/grpc/status"

	etcd "github.com/coreos/etcd/clientv3"

	wrapper "github.com/g4zhuj/grpc-wrapper"
)

const (
	resolverTimeOut = 10 * time.Second
)

type etcdRegistry struct {
	cancal context.CancelFunc
	cli    *etcd.Client
	lsCli  etcd.Lease
}

type etcdWatcher struct {
	cli       *etcd.Client
	target    string
	cancel    context.CancelFunc
	ctx       context.Context
	watchChan etcd.WatchChan
}

//NewEtcdResolver create a resolver for grpc
func NewEtcdResolver(cli *etcd.Client) naming.Resolver {
	return &etcdRegistry{
		cli: cli,
	}
}

//NewEtcdRegisty create a reistry for registering server addr
func NewEtcdRegisty(cli *etcd.Client) wrapper.Registry {
	return &etcdRegistry{
		cli:   cli,
		lsCli: etcd.NewLease(cli),
	}
}

func (er *etcdRegistry) Resolve(target string) (naming.Watcher, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), resolverTimeOut)
	w := &etcdWatcher{
		cli:    er.cli,
		target: target + "/",
		ctx:    ctx,
		cancel: cancel,
	}
	return w, nil
}

func (er *etcdRegistry) Register(ctx context.Context, target string, update naming.Update, opts ...wrapper.RegistryOptions) (err error) {
	var upBytes []byte
	if upBytes, err = json.Marshal(update); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	ctx, cancel := context.WithTimeout(context.TODO(), resolverTimeOut)
	er.cancal = cancel
	rgOpt := wrapper.RegistryOption{TTL: wrapper.DefaultRegInfTTL}
	for _, opt := range opts {
		opt(&rgOpt)
	}

	switch update.Op {
	case naming.Add:
		lsRsp, err := er.lsCli.Grant(ctx, int64(rgOpt.TTL/time.Second))
		if err != nil {
			return err
		}
		etcdOpts := []etcd.OpOption{etcd.WithLease(lsRsp.ID)}
		key := target + "/" + update.Addr
		_, err = er.cli.KV.Put(ctx, key, string(upBytes), etcdOpts...)
		if err != nil {
			return err
		}
		grpclog.Infof("etcd put key:%v channel value:%v", key, string(upBytes))
		lsRspChan, err := er.lsCli.KeepAlive(context.TODO(), lsRsp.ID)
		if err != nil {
			return err
		}
		go func() {
			for {
				_, ok := <-lsRspChan
				if !ok {
					grpclog.Fatalf("%v keepalive channel is closing", key)
					break
				}
			}
		}()
	case naming.Delete:
		_, err = er.cli.Delete(ctx, target+"/"+update.Addr)
	default:
		return status.Error(codes.InvalidArgument, "unsupported op")
	}
	return nil
}

func (er *etcdRegistry) Close() {
	er.cancal()
	er.cli.Close()
}

func (ew *etcdWatcher) Next() ([]*naming.Update, error) {
	var updates []*naming.Update
	if ew.watchChan == nil {
		//create new chan
		resp, err := ew.cli.Get(ew.ctx, ew.target, etcd.WithPrefix(), etcd.WithSerializable())
		if err != nil {
			return nil, err
		}
		for _, kv := range resp.Kvs {
			var upt naming.Update
			if err := json.Unmarshal(kv.Value, &upt); err != nil {
				continue
			}
			updates = append(updates, &upt)
		}
		opts := []etcd.OpOption{etcd.WithRev(resp.Header.Revision + 1), etcd.WithPrefix(), etcd.WithPrevKV()}
		ew.watchChan = ew.cli.Watch(context.TODO(), ew.target, opts...)
		return updates, nil
	}

	wrsp, ok := <-ew.watchChan
	if !ok {
		err := status.Error(codes.Unavailable, "etcd watch closed")
		return nil, err
	}
	if wrsp.Err() != nil {
		return nil, wrsp.Err()
	}
	for _, e := range wrsp.Events {
		var upt naming.Update
		var err error
		switch e.Type {
		case etcd.EventTypePut:
			err = json.Unmarshal(e.Kv.Value, &upt)
			upt.Op = naming.Add
		case etcd.EventTypeDelete:
			err = json.Unmarshal(e.PrevKv.Value, &upt)
			upt.Op = naming.Delete
		}

		if err != nil {
			continue
		}
		updates = append(updates, &upt)
	}
	return updates, nil
}

func (ew *etcdWatcher) Close() {
	ew.cancel()
	ew.cli.Close()
}
