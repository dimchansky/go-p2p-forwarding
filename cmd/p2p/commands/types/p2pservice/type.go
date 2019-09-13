//go:generate enumer -type=Type

package p2pservice

type Type int

const (
	PortForwarder Type = iota
	Socks5
)
