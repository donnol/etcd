package clientv3

import (
	"context"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TestClientV3(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{
			"127.0.0.1:2379",
		},
		AutoSyncInterval:     0,
		DialTimeout:          0,
		DialKeepAliveTime:    0,
		DialKeepAliveTimeout: 0,
		MaxCallSendMsgSize:   0,
		MaxCallRecvMsgSize:   0,
		TLS:                  nil,
		Username:             "",
		Password:             "",
		RejectOldCluster:     false,
		DialOptions:          []grpc.DialOption{},
		Context:              nil,
		Logger:               zap.L(),
		LogConfig:            nil,
		PermitWithoutStream:  false,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	resp, err := client.Get(ctx, "key")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("get resp: %+v", resp)

	leaseResp, err := client.Grant(ctx, 5)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("lease resp: %+v", leaseResp)

	putResp, err := client.Put(ctx, "key", "value", clientv3.WithLease(leaseResp.ID))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("put resp: %+v", putResp)

	// 租约到期前后获取
	{
		resp, err := client.Get(ctx, "key")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("get resp: %+v", resp)

		time.Sleep(6 * time.Second)

		resp, err = client.Get(ctx, "key")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("get resp: %+v", resp)
	}

	// watch
	watchCh := client.Watch(ctx, "key")
	go func() {
		time.Sleep(time.Second)
		if err := client.RequestProgress(ctx); err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)

		leaseResp, err := client.Grant(ctx, 5)
		if err != nil {
			panic(err)
		}
		t.Logf("lease resp: %+v", leaseResp)

		putResp, err := client.Put(ctx, "key", "value", clientv3.WithLease(leaseResp.ID))
		if err != nil {
			panic(err)
		}
		t.Logf("put resp: %+v", putResp)
	}()
	v := <-watchCh
	t.Logf("watch value: %+v", v)
}
