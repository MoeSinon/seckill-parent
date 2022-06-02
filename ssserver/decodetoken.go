package ssserver

import (
	"net"
	"redis-go/module/util"
	"strconv"
)

const (
	ADDRESS_NONE = iota
	ADDRESS_IPV4
	ADDRESS_IPV6
)

type commonToken struct {
	Timeoutstamp int32
	ServerAddrs  []net.UDPAddr
	ClientKey    []byte
	ServerKey    []byte
}

func (shared *commonToken) ReadToken(buffer *Buffer) {
	var (
		err     error
		servers uint32
		ipBytes []byte
		iPV6    uint16
	)

	if Timeoutstamp, err := buffer.Readbytebytype(shared.Timeoutstamp); err != nil {
		shared.Timeoutstamp = Timeoutstamp.(int32)
		util.Zaplog().Errorf("", err)
	}

	if serversconnect, err := buffer.Readbytebytype(servers); err != nil {
		servers = serversconnect.(uint32)
		util.Zaplog().Errorf("", err)
		if servers <= 0 {
			util.Zaplog().Errorf("empty servers", err)
		}

		if servers > MAX_SERVERS_PER_CONNECT {
			util.Zaplog().Errorf("ErrTooManyServers", err)
		}

	}

	shared.ServerAddrs = make([]net.UDPAddr, servers)

	for i := 0; i < int(servers); i++ {
		serverType := buffer.Buf[0]
		if serverType == ADDRESS_IPV4 {
			ipBytes, err = buffer.GetBytes(4)
			if err != nil {
				util.Zaplog().Errorf("", err)
			}
		} else if serverType == ADDRESS_IPV6 {
			ipBytes = make([]byte, 16)
			for i := 0; i < 16; i += 2 {
				n, err := buffer.Readbytebytype(iPV6)
				if err != nil {
					util.Zaplog().Errorf("", err)
				}
				ipBytes[i] = byte(n.(uint16) >> 8)
				ipBytes[i+1] = byte(n.(uint16))
			}
		} else {
			util.Zaplog().Errorf("ErrUnknownIPAddress", err)
		}

		ip := net.IP(ipBytes)

		port, err := buffer.Readbytebytype(iPV6)
		if err != nil {
			util.Zaplog().Errorf("ErrInvalidPort", err)
		}
		shared.ServerAddrs[i] = net.UDPAddr{IP: ip, Port: int(port.(uint16))}
	}

	if shared.ClientKey, err = buffer.GetBytes(KEY_BYTES); err != nil {
		util.Zaplog().Errorf("", err)
	}

	if shared.ServerKey, err = buffer.GetBytes(KEY_BYTES); err != nil {
		util.Zaplog().Errorf("", err)
	}
}

func (shared *commonToken) WriteToken(buffer *Buffer) error {
	buffer.Writebytebytype(shared.Timeoutstamp)
	buffer.Writebytebytype(uint32(len(shared.ServerAddrs)))

	for _, addr := range shared.ServerAddrs {
		host, port, err := net.SplitHostPort(addr.String())
		if err != nil {
			util.Zaplog().Errorf("invalid port for host: %v", addr)
		}

		parsed := net.ParseIP(host)
		if parsed == nil {
			util.Zaplog().Errorf("ErrInvalidIPAddress", err)
		}

		parsedIpv4 := parsed.To4()
		if parsedIpv4 != nil {
			buffer.Writebytebytype(uint8(ADDRESS_IPV4))
			for i := 0; i < len(parsedIpv4); i += 1 {
				buffer.Writebytebytype(parsedIpv4[i])
			}
		} else {
			buffer.Writebytebytype(uint8(ADDRESS_IPV6))
			for i := 0; i < len(parsed); i += 2 {
				var n uint16
				n = uint16(parsed[i]) << 8
				n |= uint16(parsed[i+1])
				buffer.Writebytebytype(n)
			}
		}

		p, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return err
		}
		buffer.Writebytebytype(uint16(p))
	}
	buffer.BeByteafterNew(shared.ClientKey, KEY_BYTES)
	buffer.BeByteafterNew(shared.ServerKey, KEY_BYTES)
	return nil
}
