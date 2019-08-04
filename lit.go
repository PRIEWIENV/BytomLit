package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/mit-dci/lit/logging"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/litrpc"
	"github.com/mit-dci/lit/qln"

	flags "github.com/jessevdk/go-flags"
)

type litConfig struct { // define a struct for usage with go-flags
	// TODO: different networks
	// networks lit can try connecting to
	Bmhost string `long:"bm" description:"bytom mainnet full node."`
	// system specific configs
	LitHomeDir string `long:"dir" description:"Specify Home Directory of lit as an absolute path."`
	TrackerURL string `long:"tracker" description:"LN address tracker URL http|https://host:port"`
	ConfigFile string
	UnauthRPC  bool `long:"unauthrpc" description:"Enables unauthenticated Websocket RPC"`

	// proxy
	ProxyURL      string `long:"proxy" description:"SOCKS5 proxy to use for communicating with the network"`
	LitProxyURL   string `long:"litproxy" description:"SOCKS5 proxy to use for Lit's network communications. Overridden by the proxy flag."`
	ChainProxyURL string `long:"chainproxy" description:"SOCKS5 proxy to use for Wallit's network communications. Overridden by the proxy flag."`

	// UPnP port forwarding and NAT Traversal
	Nat string `long:"nat" description:"Toggle upnp or pmp NAT Traversal NAT Punching"`
	//resync and tower config
	Resync string `long:"resync" description:"Resync the given chain from the given tip (requires --tip) or from default params"`
	Tip    int32  `long:"tip" description:"Given tip to resync from"`
	Tower  bool   `long:"tower" description:"Watchtower: Run a watching node"`
	Hard   bool   `short:"t" long:"hard" description:"Flag to set networks."`

	// logging and debug parameters
	LogLevel []bool `short:"v" description:"Set verbosity level to verbose (-v), very verbose (-vv) or very very verbose (-vvv)"`

	// rpc server config
	Rpcport uint16 `short:"p" long:"rpcport" description:"Set RPC port to connect to"`
	Rpchost string `long:"rpchost" description:"Set RPC host to listen to"`
	// auto config
	AutoReconnect                   bool  `long:"autoReconnect" description:"Attempts to automatically reconnect to known peers periodically."`
	AutoReconnectInterval           int64 `long:"autoReconnectInterval" description:"The interval (in seconds) the reconnect logic should be executed"`
	AutoReconnectOnlyConnectedCoins bool  `long:"autoReconnectOnlyConnectedCoins" description:"Only reconnect to peers that we have channels with in a coin whose coin daemon is available"`
	AutoListenPort                  int   `long:"autoListenPort" description:"When auto reconnect enabled, starts listening on this port"`
	NoAutoListen                    bool  `long:"noautolisten" description:"Don't automatically listen on any ports."`
	Params                          *coinparam.Params
}

var (
	defaultLitHomeDirName                  = os.Getenv("HOME") + "/.blit"
	defaultTrackerURL                      = "http://hubris.media.mit.edu:46580"
	defaultKeyFileName                     = "privkey.hex"
	defaultConfigFilename                  = "lit.conf"
	defaultHomeDir                         = os.Getenv("HOME")
	defaultRpcport                         = uint16(8001)
	defaultRpchost                         = "localhost"
	defaultAutoReconnect                   = true
	defaultNoAutoListen                    = false
	defaultAutoListenPort                  = 2448
	defaultAutoReconnectInterval           = int64(60)
	defaultUpnPFlag                        = false
	defaultLogLevel                        = 0
	defaultAutoReconnectOnlyConnectedCoins = false
	defaultUnauthRPC                       = false
)

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// newConfigParser returns a new command line flags parser.
func newConfigParser(conf *litConfig, options flags.Options) *flags.Parser {
	parser := flags.NewParser(conf, options)
	return parser
}

// TODO: do we need to link bytom mainnet wallet?
func linkWallets(node *qln.LitNode, key *[32]byte, conf *litConfig) error {
	return nil
}

func main() {

	// TODO: config by cli
	conf := litConfig{
		LitHomeDir:                      defaultLitHomeDirName,
		Rpcport:                         defaultRpcport,
		Rpchost:                         defaultRpchost,
		TrackerURL:                      defaultTrackerURL,
		AutoReconnect:                   defaultAutoReconnect,
		NoAutoListen:                    defaultNoAutoListen,
		AutoListenPort:                  defaultAutoListenPort,
		AutoReconnectInterval:           defaultAutoReconnectInterval,
		AutoReconnectOnlyConnectedCoins: defaultAutoReconnectOnlyConnectedCoins,
		UnauthRPC:                       defaultUnauthRPC,
	}

	key := litSetup(&conf)
	if conf.ProxyURL != "" {
		conf.LitProxyURL = conf.ProxyURL
		conf.ChainProxyURL = conf.ProxyURL
	}

	// SIGQUIT handler for debugging
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGQUIT)
		buf := make([]byte, 1<<20)
		for {
			<-sigs
			stacklen := runtime.Stack(buf, true)
			logging.Warnf("=== received SIGQUIT ===\n*** goroutine dump...\n%s\n*** end\n", buf[:stacklen])
		}
	}()

	// Setup LN node.  Activate Tower if in hard mode.
	// give node and below file pathof lit home directory
	node, err := qln.NewLitNode(key, conf.LitHomeDir, conf.TrackerURL, conf.LitProxyURL, conf.Nat)
	if err != nil {
		logging.Fatal(err)
	}

	// node is up; link wallets based on args
	err = linkWallets(node, key, &conf)
	if err != nil {
		// if we don't link wallet, we can still continue, no worries.
		logging.Error(err)
	}

	rpcl := new(litrpc.LitRPC)
	rpcl.Node = node
	rpcl.OffButton = make(chan bool, 1)
	node.RPC = rpcl

	// "conf.UnauthRPC" enables unauthenticated Websocket RPC. Default - false.
	if conf.UnauthRPC {
		go litrpc.RPCListen(rpcl, conf.Rpchost, conf.Rpcport)
	}

	// conf.AutoReconnect Attempts to automatically reconnect to known peers periodically. Default - true
	// conf.NoAutoListen Don't automatically listen on any ports. Default - false
	if conf.AutoReconnect && !conf.NoAutoListen {
		node.AutoReconnect(conf.AutoListenPort, conf.AutoReconnectInterval, conf.AutoReconnectOnlyConnectedCoins)
	}

	<-rpcl.OffButton
	logging.Infof("Got stop request\n")
	time.Sleep(time.Second)

	return
	// New directory being created over at PWD
	// conf file being created at /
}
