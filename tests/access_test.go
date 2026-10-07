package tests

import (
	"testing"

	"webtyp.com/network"
)

func TestAccess(t *testing.T) {
	cases := []struct {
		a    network.Access
		name string
	}{
		{network.AccessLocal, network.AccessLocalName},
		{network.AccessInternetFiltered, network.AccessInternetFilteredName},
		{network.AccessInternet, network.AccessInternetName},
		{network.Access(0), network.AccessLocalName}, // zero is local
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if s := tc.a.String(); s != tc.name {
				t.Errorf("String() for %d: expected %s, got %s", tc.a, tc.name, s)
			}
			parsed, err := network.ParseAccess(tc.name)
			if err != nil {
				t.Errorf("ParseAccess(%s) error: %v", tc.name, err)
			}
			if parsed != tc.a && !(tc.a == 0 && parsed == network.AccessLocal) {
				t.Errorf("ParseAccess(%s): expected %d, got %d", tc.name, tc.a, parsed)
			}
		})
	}

	t.Run("unknown name", func(t *testing.T) {
		_, err := network.ParseAccess("invalid_name")
		if err != network.ErrUnknownAccess {
			t.Errorf("Expected ErrUnknownAccess, got %v", err)
		}
	})

	t.Run("unknown access String", func(t *testing.T) {
		s := network.Access(99).String()
		if s != "" {
			t.Errorf("Expected empty string, got %s", s)
		}
	})
}

func TestUnregistered(t *testing.T) {
	cases := []struct {
		u    network.Unregistered
		name string
	}{
		{network.UnregisteredNoAddress, network.UnregisteredNoAddressName},
		{network.UnregisteredLocal, network.UnregisteredLocalName},
		{network.Unregistered(0), network.UnregisteredNoAddressName}, // zero is no address
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if s := tc.u.String(); s != tc.name {
				t.Errorf("String() for %d: expected %s, got %s", tc.u, tc.name, s)
			}
			parsed, err := network.ParseUnregistered(tc.name)
			if err != nil {
				t.Errorf("ParseUnregistered(%s) error: %v", tc.name, err)
			}
			if parsed != tc.u && !(tc.u == 0 && parsed == network.UnregisteredNoAddress) {
				t.Errorf("ParseUnregistered(%s): expected %d, got %d", tc.name, tc.u, parsed)
			}
		})
	}

	t.Run("unknown name", func(t *testing.T) {
		_, err := network.ParseUnregistered("invalid_name")
		if err != network.ErrUnknownUnregistered {
			t.Errorf("Expected ErrUnknownUnregistered, got %v", err)
		}
	})

	t.Run("unknown unregistered String", func(t *testing.T) {
		s := network.Unregistered(99).String()
		if s != "" {
			t.Errorf("Expected empty string, got %s", s)
		}
	})
}
