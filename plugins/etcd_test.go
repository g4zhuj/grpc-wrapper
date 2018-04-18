package plugins

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc/naming"

	etcd "github.com/coreos/etcd/clientv3"
)

var host = "http://localhost:2379"
var key = "test"

func TestRegisterAndResolver(t *testing.T) {
	cli, err := etcd.NewFromURL(host)
	if err != nil {
		fmt.Printf("nameing err %v\n", err)
		return
	}

	reg := NewEtcdRegisty(cli)
	upAdd := naming.Update{Op: naming.Add, Addr: "127.0.0.1:2342", Metadata: "..."}

	err = reg.Register(context.TODO(), key, upAdd)
	if err != nil {
		t.Fatalf("register err %v", err)
	}

	resolver := NewEtcdResolver(cli)
	wh, err := resolver.Resolve(key)
	if err != nil {
		t.Fatalf("Resolve err %v", err)
	}

	ups, err := wh.Next()
	if err != nil {
		t.Fatalf("Next err %v", err)
	}
	for _, up := range ups {
		t.Logf("update %v", *up)
		if up.Addr != upAdd.Addr {
			t.Fatalf("expected %v got %v", upAdd.Addr, up.Addr)
		}
	}
}
