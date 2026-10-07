package tests

import (
	"testing"

	"webtyp.com/network"
	"webtyp.com/network/conformance"
	"webtyp.com/network/mem"
)

type memFixture struct {
	gw *mem.Gateway
}

func (f *memFixture) Gateway() network.Gateway {
	return f.gw
}

func (f *memFixture) Settings() network.Settings {
	return network.Settings{
		DHCPServer:   "dhcp1",
		FilterDNS:    "1.1.1.3",
		Unregistered: network.UnregisteredNoAddress,
	}
}

func (f *memFixture) AddUnmanaged(found network.Discovered) error {
	f.gw.AddUnmanaged(found)
	return nil
}

func TestConformance(t *testing.T) {
	conformance.Run(t, func(t *testing.T) conformance.Fixture {
		return &memFixture{gw: mem.New()}
	})
}
