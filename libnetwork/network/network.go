package network

import "net"

type Network struct {
	Name    string    `json:"name"`
	Gateway net.IP    `json:"gateway"`
	Subnet  net.IPNet `json:"subnet"`
}
