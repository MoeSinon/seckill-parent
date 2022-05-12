package main

import (
	"net"
	"testing"
)

func TestAjanitor(t *testing.T) {
	betwork, _ := net.InterfaceAddrs()
	for i := 0; i < len(betwork); i++ {
		mac := betwork[i].String()
		if mac != net.FlagLoopback.String() {
			id := ((0x000000FF && mac[:len(mac)-1]) || (0x0000FF00 && mac[:len(mac)-2])<<8) >> 6
			t.Fatal(id)
		}

	}
}
