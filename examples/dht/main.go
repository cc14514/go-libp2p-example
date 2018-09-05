package main

import (
	"bytes"
	"context"
	"fmt"
	"gx/ipfs/QmQ1hwb95uSSZR8jSPJysnfHxBDQAykSXsmz5TwTzxjq2Z/go-libp2p-host"
	"gx/ipfs/QmUDzeFgYrRmHL2hUB6NZmqcBVQtUzETwmFRUc9onfSSHr/go-libp2p/p2p/host/basic"
	"gx/ipfs/QmYLXCWN2myozZpx8Wx4UjrRuQuhY3YtWoMi6SHaXii6aM/go-libp2p-peerstore"
	"gx/ipfs/QmZAsayEQakfFbHyakgHRKHwBTWrwuSBTfaMyxJZUG97VC/go-libp2p-kad-dht"
	"gx/ipfs/QmcZSzKEM5yDfpZbeEEZaVmaZ1zXm6JWTbrQZSB8hCVPzk/go-libp2p-peer"
	"time"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"github.com/cc14514/go-libp2p-example/helper"
)

type Node struct {
	Host    host.Host
	Routing *dht.IpfsDHT
}

func NewLocalNode(port int) *Node {
	h, _ := basichost.NewHost(context.Background(), helper.GenSwarm(), &basichost.HostOpts{})
	maddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	if err!=nil {
		panic(err)
	}
	h.Network().Listen(maddr)
	h.Peerstore().AddAddrs(h.ID(),[]multiaddr.Multiaddr{maddr},peerstore.ProviderAddrTTL)
	d, _ := dht.New(context.Background(), h)
	return &Node{h, d}
}

func (self *Node) Close() {
	self.Routing.Close()
	self.Host.Close()
}

func (self *Node) Bootstrap(ctx context.Context) error {
	return self.Routing.Bootstrap(ctx)
}

func (self *Node) Connect(ctx context.Context, bb *Node) {
	a, b := self.Host, bb.Host
	idB := b.ID()
	addrB := b.Network().Peerstore().Addrs(idB)
	if len(addrB) == 0 {
		fmt.Println("peers setup incorrectly: no local address")
	}

	a.Peerstore().AddAddrs(idB, addrB, peerstore.TempAddrTTL)
	pi := peerstore.PeerInfo{ID: idB}
	if err := a.Connect(ctx, pi); err != nil {
		fmt.Println(err)
	}
}

func showPeerPretty(peers []peer.ID) string {
	if len(peers) == 0 {
		return ""
	}
	buf := bytes.NewBufferString("->")
	for _, p := range peers {
		buf.WriteString(p.Pretty())
		buf.WriteString(", ")
	}
	return buf.String()
}

func main() {
	// group 1
	n1 := NewLocalNode(40001)

	n2 := NewLocalNode(40002)

	n3 := NewLocalNode(40003)

	// group 2
	n4 := NewLocalNode(40004)

	n5 := NewLocalNode(40005)

	defer func() {
		n1.Close()
		n2.Close()
		n3.Close()
		n4.Close()
		n5.Close()
	}()

	n1.Connect(context.Background(), n2)
	n2.Connect(context.Background(), n3)
	n3.Connect(context.Background(), n4)
	n4.Connect(context.Background(), n5)

	fmt.Println(1, n1.Host.Network().ListenAddresses(), showPeerPretty(n1.Host.Peerstore().Peers()))
	fmt.Println(2, n2.Host.ID().Pretty(), showPeerPretty(n2.Host.Peerstore().Peers()))
	fmt.Println(3, n3.Host.ID().Pretty(), showPeerPretty(n3.Host.Peerstore().Peers()))
	fmt.Println(4, n4.Host.ID().Pretty(), showPeerPretty(n4.Host.Peerstore().Peers()))
	fmt.Println(5, n5.Host.ID().Pretty(), showPeerPretty(n5.Host.Peerstore().Peers()))

	err := n5.Bootstrap(context.Background())

	fmt.Println("----> bootstrap:", err)
	<-time.After(time.Second * 3)
	fmt.Println(4, n4.Host.ID().Pretty(), showPeerPretty(n4.Host.Peerstore().Peers()))
	fmt.Println(5, n5.Host.ID().Pretty(), showPeerPretty(n5.Host.Peerstore().Peers()))

}
