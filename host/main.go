package main

import (
	"crypto/rand"
	"fmt"
	"context"
	"gx/ipfs/QmcZSzKEM5yDfpZbeEEZaVmaZ1zXm6JWTbrQZSB8hCVPzk/go-libp2p-peer"
	"gx/ipfs/QmYLXCWN2myozZpx8Wx4UjrRuQuhY3YtWoMi6SHaXii6aM/go-libp2p-peerstore"
	"gx/ipfs/QmdjC8HtKZpEufBL1u7WxvQn78Lqq2Wk31NJS8WvFX3crB/go-libp2p-swarm"
	"gx/ipfs/QmSYxVNTGJ4MEy7UH26qgLbidfSgdYiEkV6affSitghk4S/go-tcp-transport"
	"gx/ipfs/Qmahr4UDXGj8zXFhXzQZpmoJgeVi6XX8pEnuVi2bFkENqj/go-conn-security-multistream"
	"gx/ipfs/QmZ69yKFZ7V5hLCYFWzJF1Nz4dH1Es4LaE8jGXAFucLe3K/go-libp2p-secio"
	"gx/ipfs/QmUDzeFgYrRmHL2hUB6NZmqcBVQtUzETwmFRUc9onfSSHr/go-libp2p/p2p/host/basic"
	"gx/ipfs/QmUDzeFgYrRmHL2hUB6NZmqcBVQtUzETwmFRUc9onfSSHr/go-libp2p/p2p/protocol/ping"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	ic "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	tptu "gx/ipfs/QmP7znopdZogwxPJyRKEZSNnP7HfnUCaQjaMNDmPw8VE2Y/go-libp2p-transport-upgrader"
	"gx/ipfs/QmcsgrV3nCAKjiHKZhKVXWc4oY3WBECJCqahXEMpHeMrev/go-smux-yamux"
	"gx/ipfs/QmWzjXAyBTygw6CeCTUnhJzhFucfxY5FJivSoiGuiSbPjS/go-smux-multistream"
)

func GenSwarm(port int) *swarm.Swarm {
	priv, pub, err := ic.GenerateKeyPairWithReader(ic.RSA, 2048, rand.Reader)
	if err != nil {
		panic(err) // oh no!
	}

	pid, err := peer.IDFromPublicKey(pub)
	if err != nil {
		panic(err)
	}
	ps := peerstore.NewPeerstore()
	ps.AddPubKey(pid, pub)
	ps.AddPrivKey(pid, priv)
	s := swarm.NewSwarm(context.Background(), pid, ps, nil)

	//NewTCPTransport
	tcpTransport := tcp.NewTCPTransport(GenUpgrader(s))

	if err := s.AddTransport(tcpTransport); err != nil {
		fmt.Println(err)
	}
	maddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	s.Listen(maddr)
	s.Peerstore().AddAddrs(pid, s.ListenAddresses(), peerstore.PermanentAddrTTL)
	return s
}

// GenUpgrader creates a new connection upgrader for use with this swarm.
func GenUpgrader(n *swarm.Swarm) *tptu.Upgrader {
	id := n.LocalPeer()
	pk := n.Peerstore().PrivKey(id)
	secMuxer := new(csms.SSMuxer)
	secMuxer.AddTransport(secio.ID, &secio.Transport{
		LocalID:    id,
		PrivateKey: pk,
	})
	multistream.NewBlankTransport()
	stMuxer := multistream.NewBlankTransport()
	stMuxer.AddTransport("/yamux/1.0.0", sm_yamux.DefaultTransport)
	return &tptu.Upgrader{
		Secure:  secMuxer,
		Muxer:   stMuxer,
		Filters: n.Filters,
	}
}

func main() {
	// 创建两个 host 对象
	h1, e1 := basichost.NewHost(context.Background(), GenSwarm(40001), &basichost.HostOpts{})
	defer h1.Close()
	fmt.Println(e1, h1.ID().Pretty(), h1.Network().ListenAddresses())

	h2, e2 := basichost.NewHost(context.Background(), GenSwarm(40002), &basichost.HostOpts{})
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
