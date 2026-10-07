package tests

import (
	"testing"

	"webtyp.com/network"
	"webtyp.com/network/mem"
)

func TestConnections(t *testing.T) {
	gw := mem.New()

	c1 := network.Connection{
		MAC:      "02:00:00:00:00:01",
		IP:       "10.99.0.11",
		HostName: "iphone",
		Source:   network.SourceDHCP,
	}

	c2 := network.Connection{
		MAC:      "02:00:00:00:00:02",
		IP:       "10.99.0.12",
		HostName: "",
		Source:   network.SourceARP,
	}

	gw.Connect(c1)
	gw.Connect(c2)

	conns, err := gw.Connections()
	if err != nil {
		t.Fatalf("Connections() error: %v", err)
	}

	if len(conns) != 2 {
		t.Fatalf("Expected 2 connections, got %d", len(conns))
	}

	if conns[0] != c1 {
		t.Errorf("Expected %v, got %v", c1, conns[0])
	}
	if conns[1] != c2 {
		t.Errorf("Expected %v, got %v", c2, conns[1])
	}
}
