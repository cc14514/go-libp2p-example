package main

import (
	"fmt"
	"context"
	"gx/ipfs/QmYLXCWN2myozZpx8Wx4UjrRuQuhY3YtWoMi6SHaXii6aM/go-libp2p-peerstore"
	"gx/ipfs/QmUDzeFgYrRmHL2hUB6NZmqcBVQtUzETwmFRUc9onfSSHr/go-libp2p/p2p/host/basic"
	"gx/ipfs/QmUDzeFgYrRmHL2hUB6NZmqcBVQtUzETwmFRUc9onfSSHr/go-libp2p/p2p/protocol/ping"
	"github.com/cc14514/go-libp2p-example/helper"
)

func main() {
	// 创建两个 host 对象
	h1, e1 := basichost.NewHost(context.Background(), helper.GenSwarm(40001), &basichost.HostOpts{})
	defer h1.Close()
	fmt.Println(e1, h1.ID().Pretty(), h1.Network().ListenAddresses())

	h2, e2 := basichost.NewHost(context.Background(), helper.GenSwarm(40002), &basichost.HostOpts{})
	defer h2.Close()
	fmt.Println(e2, h2.ID().Pretty(), h2.Network().ListenAddresses())

	// 将 h1 放入 h2 的 peer 列表中，否则无法 connect
	h2.Peerstore().AddAddrs(h1.ID(),h1.Network().ListenAddresses(),peerstore.PermanentAddrTTL)
	// 用 h2 连接 h1，此时如果 err == nil，则 h1 和 h2 互为邻居
	err := h2.Connect(context.Background(), peerstore.PeerInfo{
		ID:    h1.ID(),
		Addrs: h1.Addrs(),
	})
	if err != nil {
		fmt.Println("conn_err :", err)
	}

	// 创建两个 pingService
	p1 := ping.NewPingService(h1)
	p2 := ping.NewPingService(h2)

	// h1 ping h2
	ct1, e1 := p1.Ping(context.Background(), h2.ID())
	fmt.Println("ping_1_e->", e1)

	// h2 ping h1
	ct2, e2 := p2.Ping(context.Background(), h1.ID())
	fmt.Println("ping_1_e->", e2)

	t1 := <-ct1
	fmt.Println("ttl_1->", t1)

	t2 := <-ct2
	fmt.Println("ttl_2->", t2)

}
