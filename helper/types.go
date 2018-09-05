package helper

import (
	"context"
	"gx/ipfs/QmQ1hwb95uSSZR8jSPJysnfHxBDQAykSXsmz5TwTzxjq2Z/go-libp2p-host"
	"gx/ipfs/QmUDzeFgYrRmHL2hUB6NZmqcBVQtUzETwmFRUc9onfSSHr/go-libp2p/p2p/host/basic"
	"gx/ipfs/QmYLXCWN2myozZpx8Wx4UjrRuQuhY3YtWoMi6SHaXii6aM/go-libp2p-peerstore"
	"gx/ipfs/QmZAsayEQakfFbHyakgHRKHwBTWrwuSBTfaMyxJZUG97VC/go-libp2p-kad-dht"
	opts "gx/ipfs/QmZAsayEQakfFbHyakgHRKHwBTWrwuSBTfaMyxJZUG97VC/go-libp2p-kad-dht/opts"
	ic "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	"gx/ipfs/QmcZSzKEM5yDfpZbeEEZaVmaZ1zXm6JWTbrQZSB8hCVPzk/go-libp2p-peer"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"reflect"
	pstore "gx/ipfs/QmYLXCWN2myozZpx8Wx4UjrRuQuhY3YtWoMi6SHaXii6aM/go-libp2p-peerstore"
	"github.com/alecthomas/log4go"
	"fmt"
)

type blankValidator struct{}

func (blankValidator) Validate(_ string, _ []byte) error        { return nil }
func (blankValidator) Select(_ string, _ [][]byte) (int, error) { return 0, nil }

type Node struct {
	Host    host.Host
	Routing *dht.IpfsDHT
}

func NewLocalNode() *Node {
	h, _ := basichost.NewHost(context.Background(), GenSwarm(), &basichost.HostOpts{})
	d, _ := dht.New(context.Background(), h, opts.NamespacedValidator("cc14514", blankValidator{}), )
	return &Node{h, d}
}

func NewNode(key ic.PrivKey, port int) *Node {
	h, _ := basichost.NewHost(context.Background(), GenSwarmByKey(key), &basichost.HostOpts{})
	maddr1, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	maddr2, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	h.Network().Listen(maddr1, maddr2)

	d, _ := dht.New(context.Background(), h, opts.NamespacedValidator("cc14514", blankValidator{}), )
	return &Node{h, d}
}

func (self *Node) Close() {
	self.Routing.Close()
	self.Host.Close()
}

func (self *Node) Bootstrap(ctx context.Context) error {
	return self.Routing.Bootstrap(ctx)
}

func (self *Node) Connect(ctx context.Context, targetID interface{}, targetAddrs []ma.Multiaddr) error {
	var tid peer.ID
	if "string" == reflect.TypeOf(targetID).Name() {
		stid := targetID.(string)
		tid, _ = peer.IDFromString(stid)
	} else {
		tid = targetID.(peer.ID)
	}
	a := self.Host
	a.Peerstore().AddAddrs(tid, targetAddrs, peerstore.TempAddrTTL)
	pi := peerstore.PeerInfo{ID: tid}
	if err := a.Connect(ctx, pi); err != nil {
		return err
	}
	return nil
}

func (self *Node) PutValue(ctx context.Context, key string, value []byte) error {
	return self.Routing.PutValue(ctx, key, value)
}

func (self *Node) GetValue(ctx context.Context, key string) ([]byte, error) {
	return self.Routing.GetValue(ctx, key)
}

func (self *Node) FindPeer(ctx context.Context, targetID interface{}) (pstore.PeerInfo, error) {
	var (
		tid peer.ID
		err error
	)
	if "string" == reflect.TypeOf(targetID).Name() {
		stid := targetID.(string)
		tid, err = peer.IDB58Decode(stid)
		if err != nil {
			log4go.Error(err)
			return pstore.PeerInfo{}, err
		}
	} else {
		tid = targetID.(peer.ID)
	}
	return self.Routing.FindPeer(ctx, tid)
}
