package helper

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
	ic "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	tptu "gx/ipfs/QmP7znopdZogwxPJyRKEZSNnP7HfnUCaQjaMNDmPw8VE2Y/go-libp2p-transport-upgrader"
	"gx/ipfs/QmcsgrV3nCAKjiHKZhKVXWc4oY3WBECJCqahXEMpHeMrev/go-smux-yamux"
	"gx/ipfs/QmWzjXAyBTygw6CeCTUnhJzhFucfxY5FJivSoiGuiSbPjS/go-smux-multistream"
)

func GenSwarmByKey(key ic.PrivKey) (*swarm.Swarm, *tptu.Upgrader) {
	ctx := context.Background()
	priv, pub := key, key.GetPublic()
	pid, err := peer.IDFromPublicKey(pub)
	if err != nil {
		panic(err)
	}
	ps := peerstore.NewPeerstore()
	ps.AddPubKey(pid, pub)
	ps.AddPrivKey(pid, priv)
	s := swarm.NewSwarm(ctx, pid, ps, nil)

	//NewTCPTransport
	u := GenUpgrader(s)
	tcpTransport := tcp.NewTCPTransport(GenUpgrader(s))

	if err := s.AddTransport(tcpTransport); err != nil {
		fmt.Println(err)
	}
	// TODO 多地址
	//maddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	//s.Listen(maddr)
	//s.AddListenAddr(maddr)
	s.Peerstore().AddAddrs(pid, s.ListenAddresses(), peerstore.PermanentAddrTTL)
	return s, u
}
func GenSwarm() (*swarm.Swarm) {
	s,_ := GenSwarm2()
	return s
}

func GenSwarm2() (*swarm.Swarm, *tptu.Upgrader) {
	priv, _, err := ic.GenerateKeyPairWithReader(ic.RSA, 2048, rand.Reader)
	if err != nil {
		panic(err) // oh no!
	}
	return GenSwarmByKey(priv)
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
