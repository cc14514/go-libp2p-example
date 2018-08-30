package main

import (
	"github.com/cc14514/go-libp2p-example/helper"
	iaddr "gx/ipfs/QmWnUZVLLk2HKpZAMEsqW3EFNku1xGzG7bvvAHeEQQoi2V/go-ipfs-addr"
	"context"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"fmt"
	"bufio"
	"os"
	"strings"
)

var (
	BOOT_NODE  = "/ip4/101.251.230.214/tcp/40001/ipfs/QmZfJJRpXx4kLJfNq6sqKVWtGsaoaL54zG3aT2zEnA6xn7"
	LOCAL_PORT = 40001
	DATA_DIR   = ""
)

var node *helper.Node

func init() {
	prv, err := helper.LoadKey(DATA_DIR)
	if err != nil {
		prv, _ = helper.GenKey(DATA_DIR)
	}
	fmt.Println("TEST_NODE", BOOT_NODE)
	node = helper.NewNode(prv, LOCAL_PORT)

	addr, _ := iaddr.ParseString(BOOT_NODE)
	id := addr.ID()
	if id.Pretty() != node.Host.ID().Pretty() {
		/*
		err = node.Connect(context.Background(), string(id), []ma.Multiaddr{addr.Transport()})
		fmt.Println("myid :", node.Host.ID().Pretty())
		fmt.Println("connect :", err)
		*/
	}
}

func main() {
	s := make(chan int)
	go func() {
		for {
			fmt.Print("cmd #>")
			ir := bufio.NewReader(os.Stdin)
			if cmd, err := ir.ReadString('\n'); err == nil && strings.Trim(cmd, " ") != "\n" {
				cmd = strings.Trim(cmd, " ")
				cmd = cmd[:len([]byte(cmd))-1]
				cmdArg := strings.Split(cmd, " ")
				switch cmdArg[0] {
				case "exit", "quit":
					fmt.Println("bye bye ^_^ ")
					s <- 0
					return
				case "help", "?":
					help()
				case "peers":
					peers()
				case "conn":
					if len(cmdArg) >= 2 {
						conn(cmdArg[1:len(cmdArg)])
					}
				case "put":
					if len(cmdArg) == 3 {
						put(cmdArg[1], cmdArg[2])
					}
				case "get":
					if len(cmdArg) == 2 {
						val := get(cmdArg[1])
						fmt.Println(val)
					}
				default:

				}
			}
		}
	}()
	<-s
}

func peers() {
	for _, c := range node.Host.Network().Conns() {
		fmt.Printf("%s/ipfs/%s\n", c.RemoteMultiaddr().String(), c.RemotePeer().Pretty())
	}
}

func help() {
	s := `
peers				show peers 
put <key> <value> 		put key value to dht
get <key>			get value by key from dht
conn <addr>			connect to addr , "/ip4/101.251.230.214/tcp/40001/ipfs/QmZfJJRpXx4kLJfNq6sqKVWtGsaoaL54zG3aT2zEnA6xn7"	
`
	fmt.Println(s)
}

func conn(addrs []string) {
	for _, a := range addrs {
		addr, err := iaddr.ParseString(a)
		if err != nil {
			fmt.Println("addr_error :", a, err)
		} else {
			err = node.Connect(context.Background(), string(addr.ID()), []ma.Multiaddr{addr.Transport()})
			if err != nil {
				fmt.Println("conn_error :", a, err)
			} else {
				fmt.Println("success")
			}
		}
	}
}

func put(key, value string) {
	if err := node.PutValue(context.Background(), fmt.Sprintf("/cc14514/%s", key), []byte(value));err != nil {
		fmt.Println("put_error :",err)
	}else{
		fmt.Println("success")
	}

}

func get(key string) string {
	buf,err := node.GetValue(context.Background(),fmt.Sprintf("/cc14514/%s", key))
	if err != nil {
		fmt.Println("get_error :",err)
		return ""
	}
	return string(buf)
}
