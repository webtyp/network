package tests

import (
	"testing"

	"webtyp.com/network"
)

func TestValidate(t *testing.T) {
	validSettings := network.Settings{
		DHCPServer:   "dhcp1",
		FilterDNS:    "1.1.1.3",
		Unregistered: network.UnregisteredNoAddress,
	}

	validHost := network.Host{
		Name:   "PC 1",
		MAC:    "02:00:00:00:00:01",
		IP:     "10.99.0.11",
		Access: network.AccessLocal,
	}

	cases := []struct {
		name string
		d    network.Desired
	}{
		{
			"missing DHCPServer",
			network.Desired{
				Settings: network.Settings{FilterDNS: "1.1.1.3"},
			},
		},
		{
			"missing FilterDNS",
			network.Desired{
				Settings: network.Settings{DHCPServer: "dhcp1"},
			},
		},
		{
			"UnregisteredLocal without pool",
			network.Desired{
				Settings: network.Settings{
					DHCPServer:   "dhcp1",
					FilterDNS:    "1.1.1.3",
					Unregistered: network.UnregisteredLocal,
				},
			},
		},
		{
			"empty Name",
			network.Desired{
				Settings: validSettings,
				Hosts: []network.Host{
					{MAC: "02:00:00:00:00:01", IP: "10.99.0.11", Access: network.AccessLocal},
				},
			},
		},
		{
			"empty MAC",
			network.Desired{
				Settings: validSettings,
				Hosts: []network.Host{
					{Name: "PC 1", IP: "10.99.0.11", Access: network.AccessLocal},
				},
			},
		},
		{
			"empty IP",
			network.Desired{
				Settings: validSettings,
				Hosts: []network.Host{
					{Name: "PC 1", MAC: "02:00:00:00:00:01", Access: network.AccessLocal},
				},
			},
		},
		{
			"invalid Access",
			network.Desired{
				Settings: validSettings,
				Hosts: []network.Host{
					{Name: "PC 1", MAC: "02:00:00:00:00:01", IP: "10.99.0.11", Access: network.Access(99)},
				},
			},
		},
		{
			"duplicate MAC",
			network.Desired{
				Settings: validSettings,
				Hosts: []network.Host{
					validHost,
					{Name: "PC 2", MAC: "02:00:00:00:00:01", IP: "10.99.0.12", Access: network.AccessLocal},
				},
			},
		},
		{
			"duplicate IP",
			network.Desired{
				Settings: validSettings,
				Hosts: []network.Host{
					validHost,
					{Name: "PC 2", MAC: "02:00:00:00:00:02", IP: "10.99.0.11", Access: network.AccessLocal},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.d.Validate()
			if !network.IsInvalid(err) {
				t.Errorf("Expected IsInvalid to be true, got err: %v", err)
			}
		})
	}

	t.Run("valid desired", func(t *testing.T) {
		d := network.Desired{
			Settings: validSettings,
			Hosts:    []network.Host{validHost},
		}
		if err := d.Validate(); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	})

	t.Run("valid UnregisteredLocal", func(t *testing.T) {
		s := validSettings
		s.Unregistered = network.UnregisteredLocal
		s.DynamicPool = "pool1"
		d := network.Desired{
			Settings: s,
		}
		if err := d.Validate(); err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	})
}
