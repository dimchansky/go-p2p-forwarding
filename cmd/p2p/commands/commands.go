package commands

import (
	"os"

	"github.com/dimchansky/go-p2p-forwarding/cmd"
)

type RootCommands struct {
	Version func()         `short:"v" long:"version"  description:"Print the version of tool and exit."`
	Socks5  Socks5Command  `command:"socks5"          description:"Create p2p service that acts like socks5 server."`
	Listen  ListenCommand  `command:"listen"          description:"Create p2p service and forward connections made to remote <target-address>."`
	Forward ForwardCommand `command:"forward"         description:"Forward connections made to local <listen-address> to p2p <target-address>."`
	KeyGen  KeyGenCommand  `command:"keygen"          description:"Generates identity private key."`
}

var Root RootCommands

func init() {
	Root.Version = func() {
		cmd.PrintVersion()
		os.Exit(0)
	}
}
