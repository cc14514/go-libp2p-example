package main

import (
	"github.com/cc14514/go-libp2p-example/helper"
	iaddr "gx/ipfs/QmWnUZVLLk2HKpZAMEsqW3EFNku1xGzG7bvvAHeEQQoi2V/go-ipfs-addr"
	"context"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"fmt"
)

var (
	BOOT_NODE  = "/ip4/101.251.230.214/tcp/40001/ipfs/QmZfJJRpXx4kLJfNq6sqKVWtGsaoaL54zG3aT2zEnA6xn7"
	LOCAL_PORT = 40001
	DATA_DIR = ""
)

func main() {
	prv,err := helper.LoadKey(DATA_DIR)
	if err != nil {
		prv,_ = helper.GenKey(DATA_DIR)
	}
	addr,_ := iaddr.ParseString(BOOT_NODE)
	id := addr.ID()
	fmt.Println("target_id -->",id.Pretty(),addr.Transport())
	node := helper.NewNode(prv,LOCAL_PORT)
	err = node.Connect(context.Background(),string(id),[]ma.Multiaddr{addr.Transport()})
	fmt.Println("myid :",node.Host.ID().Pretty())
	fmt.Println("connect :",err)
}
