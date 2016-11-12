package killswitch

import (
	"bytes"
	"errors"
	"net"
)

type Network struct {
	Interfaces    []net.Interface
	UpInterfaces  map[string]string
	P2PInterfaces map[string]string
	PeerIP        string
	PFRules       bytes.Buffer
}

func New(peerIP string) (*Network, error) {
	ip := net.ParseIP(peerIP)
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	return &Network{
		Interfaces:    ifaces,
		UpInterfaces:  make(map[string]string),
		P2PInterfaces: make(map[string]string),
		PeerIP:        ip.String(),
	}, nil
}

func (n *Network) GetActive() error {
	for _, i := range n.Interfaces {
		if i.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if i.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := i.Addrs()
		if err != nil {
			return err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			if i.Flags&net.FlagPointToPoint != 0 {
				n.P2PInterfaces[i.Name] = ip.String()
			} else {
				n.UpInterfaces[i.Name] = ip.String()
			}
		}
	}
	if n.UpInterfaces == nil {
		return errors.New("No active connections, verify you are connected to the network")
	}
	return nil
}
