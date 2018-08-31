package main

import (
	"os"
	"github.com/urfave/cli"
	"github.com/alecthomas/log4go"
	"fmt"
	"bufio"
	"strings"
	"github.com/cc14514/go-libp2p-example/helper"
	iaddr "gx/ipfs/QmWnUZVLLk2HKpZAMEsqW3EFNku1xGzG7bvvAHeEQQoi2V/go-ipfs-addr"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"context"
	"time"
	"gx/ipfs/QmVmDhyTTUcQXFD1rRQ64fGLMSAoaQvNH3hwuaCFAPq2hy/errors"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	inet "gx/ipfs/QmVwU7Mgwg6qaPn9XXz93ANfq1PTxcduGRzfe41Sygg4mR/go-libp2p-net"
	"path"
	"io/ioutil"
	"gx/ipfs/QmcZSzKEM5yDfpZbeEEZaVmaZ1zXm6JWTbrQZSB8hCVPzk/go-libp2p-peer"
)

var (
	version             = "0.0.1"
	logLevel            = []log4go.Level{log4go.ERROR, log4go.WARNING, log4go.INFO, log4go.DEBUG}
	app                 *cli.App
	node                *helper.Node
	stop                chan struct{}
	DATA_DIR, BOOT_NODE string
)

const (
	DEF_BOOT_NODE = "/ip4/101.251.230.214/tcp/40001/ipfs/QmZfJJRpXx4kLJfNq6sqKVWtGsaoaL54zG3aT2zEnA6xn7"
	// protocols
	P_CHANNEL_FILE = protocol.ID("/channel/file")
)

func init() {
	stop = make(chan struct{})
	app = cli.NewApp()
	app.Name = os.Args[0]
	app.Usage = "go-libp2p example"
	app.Version = version
	app.Author = "liangc"
	app.Email = "cc14514@icloud.com"
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "port,p",
			Usage: "listen port",
			Value: 40001,
		},
		cli.IntFlag{
			Name:  "loglevel",
			Usage: "0:error , 1:warning , 2:info , 3:debug",
			Value: 2,
		},
		cli.StringFlag{
			Name:        "datadir",
			Usage:       "data dir on local file system.",
			Value:       "/tmp",
			Destination: &DATA_DIR,
		},
		cli.StringFlag{
			Name:        "bootnode",
			Usage:       "seed node for build p2p network.",
			Value:       DEF_BOOT_NODE,
			Destination: &BOOT_NODE,
		},
	}
	app.Action = func(ctx *cli.Context) error {
		start(ctx)
		go func() {
			t := time.NewTicker(10 * time.Second)
			for range t.C {
				for i, c := range node.Host.Network().Conns() {
					log4go.Debug("%d -> %s/ipfs/%s", i, c.RemoteMultiaddr().String(), c.RemotePeer().Pretty())
				}
			}
		}()
		<-stop
		return nil
	}
	app.Before = func(ctx *cli.Context) error {
		idx := ctx.GlobalInt("loglevel")
		level := logLevel[idx]
		log4go.AddFilter("stdout", log4go.Level(level), log4go.NewConsoleLogWriter())
		return nil
	}
	app.After = func(ctx *cli.Context) error {
		log4go.Close()
		return nil
	}
	app.Commands = []cli.Command{
		{
			Name: "version",
			Action: func(ctx *cli.Context) error {
				fmt.Println("version\t:", version)
				fmt.Println("auth\t:", app.Author)
				fmt.Println("email\t:", app.Email)
				fmt.Println("source\t: https://github.com/cc14514")
				return nil
			},
		},
		{
			Name:   "console",
			Usage:  "一个简单的交互控制台，用来调试",
			Action: consoleCmd,
		},
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func start(ctx *cli.Context) {
	prv, err := helper.LoadKey(DATA_DIR)
	if err != nil {
		prv, _ = helper.GenKey(DATA_DIR)
	}
	log4go.Info("TEST_NODE -> %s", BOOT_NODE)
	port := ctx.GlobalInt("port")
	node = helper.NewNode(prv, port)

	addr, _ := iaddr.ParseString(BOOT_NODE)
	id := addr.ID()
	if id.Pretty() != node.Host.ID().Pretty() {
		//TODO
		err = node.Connect(context.Background(), string(id), []ma.Multiaddr{addr.Transport()})
		log4go.Info("myid : %s", node.Host.ID().Pretty())
		if err != nil {
			panic(err)
		}
		log4go.Info("connect_success")
	}

	node.Host.SetStreamHandler(P_CHANNEL_FILE, func(s inet.Stream) {
		rid := s.Conn().RemotePeer().Pretty()
		dir := path.Join(DATA_DIR, "files")
		os.Mkdir(dir, 0755)
		p := path.Join(dir, fmt.Sprintf("%s_%d", rid, time.Now().Unix()))
		log4go.Info("----> path=%s",p)
		f, _ := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0755)
		defer f.Close()
		log4go.Info("will read buffer")
		buf := bufio.NewReader(s)
		for {
			log4go.Info("read buffer")
			buff, err := buf.ReadBytes(byte(1))
			log4go.Info("%d",len(buff))
			if err != nil {
				log4go.Error(err)
				s.Reset()
				return
			}else if len(buff) < 1 {
				s.Close()
				log4go.Info("read empty")
				return
			}
			log4go.Info(len(buff))
			_buff :=buff[:len(buff)-1]
			t,e := f.Write(_buff)
			log4go.Info("total: %d , e: %s",t,e)
		}
	})
}

func consoleCmd(ctx *cli.Context) error {
	start(ctx)
	funcs := map[string]func(args ... string) (interface{}, error){
		"help": func(args ... string) (interface{}, error) {
			s := `
bootstrap			build p2p network	
peers				show peers 
findpeer <id>		findpeer by peer.ID
put <key> <value> 		put key value to dht
get <key>			get value by key from dht
conn <addr>			connect to addr , "/ip4/101.251.230.214/tcp/40001/ipfs/QmZfJJRpXx4kLJfNq6sqKVWtGsaoaL54zG3aT2zEnA6xn7"	
`
			fmt.Println(s)
			return nil, nil
		},
		"peers": func(args ... string) (interface{}, error) {
			for _, c := range node.Host.Network().Conns() {
				fmt.Printf("%s/ipfs/%s\n", c.RemoteMultiaddr().String(), c.RemotePeer().Pretty())
			}
			return nil, nil
		},
		"conn": func(addrs ... string) (interface{}, error) {
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
			return nil, nil
		},
		"put": func(args ... string) (interface{}, error) {
			if len(args) != 2 {
				return nil, errors.New("fail params")
			}
			key, value := args[0], args[1]

			if err := node.PutValue(context.Background(), fmt.Sprintf("/cc14514/%s", key), []byte(value)); err != nil {
				fmt.Println("put_error :", err)
			} else {
				fmt.Println("success")
			}
			return nil, nil
		},
		"get": func(args ... string) (interface{}, error) {
			if len(args) != 1 {
				return nil, errors.New("fail params")
			}
			key := args[0]
			buf, err := node.GetValue(context.Background(), fmt.Sprintf("/cc14514/%s", key))
			if err != nil {
				fmt.Println("get_error :", err)
				return nil, err
			}
			return string(buf), nil
		},
		"bootstrap": func(args ... string) (interface{}, error) {
			err := node.Bootstrap(context.Background())
			return nil, err
		},
		"findpeer": func(args ... string) (interface{}, error) {
			if len(args) != 1 {
				return nil, errors.New("fail params")
			}
			pi, err := node.FindPeer(context.Background(), args[0])
			fmt.Println("success", pi.ID.Pretty(), pi.Addrs)
			return nil, err
		},
		"scp": func(args ... string) (interface{}, error) {
			if len(args) != 2 {
				return nil, errors.New("fail params")
			}
			to, fp := args[0], args[1]
			tid, err := peer.IDB58Decode(to)
			if err != nil {
				return nil, err
			}
			s, err := node.Host.NewStream(context.Background(), tid, P_CHANNEL_FILE)
			defer s.Close()
			if err != nil {
				return nil, err
			}
			buff, err := ioutil.ReadFile(fp)
			if err != nil {
				return nil, err
			}
			i, err := s.Write(buff)
			log4go.Info("write byte : %d ", i)
			s.Write([]byte{1})
			log4go.Info("end.")
			return nil, err
		},
	}
	<-time.After(time.Second)
	go func() {
		defer close(stop)
		fmt.Println("------------")
		fmt.Println("hello world")
		fmt.Println("------------")
		for {
			fmt.Print("cmd #>")
			ir := bufio.NewReader(os.Stdin)
			if cmd, err := ir.ReadString('\n'); err == nil && strings.Trim(cmd, " ") != "\n" {
				cmd = strings.Trim(cmd, " ")
				cmd = cmd[:len([]byte(cmd))-1]
				// TODO 用正则表达式拆分指令和参数
				cmdArg := strings.Split(cmd, " ")
				switch cmdArg[0] {
				case "exit", "quit":
					fmt.Println("bye bye ^_^ ")
					return
				case "help", "bootstrap":
					if _, err := funcs[cmdArg[0]](); err != nil {
						log4go.Error(err)
					}
				case "peers", "conn", "put", "get", "findpeer", "scp":
					log4go.Debug(cmdArg[0])
					if r, err := funcs[cmdArg[0]](cmdArg[1:len(cmdArg)]...); err != nil {
						log4go.Error(err)
					} else if r != nil {
						fmt.Println(r)
					}
				default:

				}
			}
		}
	}()
	<-stop
	return nil
}
