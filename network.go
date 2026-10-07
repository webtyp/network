package network

import "webtyp.com/fmt"

// Host is one network interface of a registered device.
// MAC is upper-case colon form ("48:F1:7F:D9:D7:B7") and IP a dotted IPv4 —
// producers canonicalize (webtyp.com/input CanonicalMAC / CanonicalIP).
type Host struct {
	Name   string // human label, e.g. "PC ECOGRAFIA 1 (wifi)"
	MAC    string
	IP     string
	Access Access
}

// Settings apply to the whole site. The consumer stores them (they change
// without recompiling); the gateway enforces them.
type Settings struct {
	DHCPServer   string // name of the gateway's DHCP server that serves the hosts
	DynamicPool  string // pool for unregistered devices; required only with UnregisteredLocal
	FilterDNS    string // IPv4 of the resolver forced on AccessInternetFiltered hosts
	Unregistered Unregistered
}

// Desired is everything the gateway must enforce.
type Desired struct {
	Settings Settings
	Hosts    []Host
}

// Validate checks Settings (DHCPServer and FilterDNS non-empty; DynamicPool
// non-empty when Unregistered is UnregisteredLocal) and Hosts (Name, MAC, IP
// non-empty; Access defined; no two hosts with the same MAC; no two with the
// same IP). The first violation is returned wrapped as described in errors.go.
func (d Desired) Validate() error {
	if d.Settings.DHCPServer == "" {
		return fmt.Errf("%w: %s", ErrInvalid, "missing DHCPServer")
	}
	if d.Settings.FilterDNS == "" {
		return fmt.Errf("%w: %s", ErrInvalid, "missing FilterDNS")
	}
	if d.Settings.Unregistered == UnregisteredLocal && d.Settings.DynamicPool == "" {
		return fmt.Errf("%w: %s", ErrInvalid, "UnregisteredLocal requires DynamicPool")
	}

	for i := 0; i < len(d.Hosts); i++ {
		h := d.Hosts[i]
		if h.Name == "" {
			return fmt.Errf("%w: %s", ErrInvalid, "empty Name")
		}
		if h.MAC == "" {
			return fmt.Errf("%w: %s", ErrInvalid, "empty MAC")
		}
		if h.IP == "" {
			return fmt.Errf("%w: %s", ErrInvalid, "empty IP")
		}

		validAccess := false
		if h.Access == AccessLocal || h.Access == AccessInternetFiltered || h.Access == AccessInternet {
			validAccess = true
		}
		if !validAccess {
			return fmt.Errf("%w: %s", ErrInvalid, "invalid Access")
		}

		for j := 0; j < i; j++ {
			if d.Hosts[j].MAC == h.MAC {
				return fmt.Errf("%w: duplicate MAC %s", ErrInvalid, h.MAC)
			}
			if d.Hosts[j].IP == h.IP {
				return fmt.Errf("%w: duplicate IP %s", ErrInvalid, h.IP)
			}
		}
	}
	return nil
}

// Connection is a device seen on the network right now.
type Connection struct {
	MAC      string
	IP       string
	HostName string // what the device calls itself (DHCP host-name), may be empty
	Source   Source
}

// Source says how the gateway knows a Connection.
type Source uint8

const (
	SourceDHCP Source = iota + 1 // holds an active DHCP lease
	SourceARP                    // answered ARP recently (includes fixed-IP devices)
)

// Discovered is configuration found on the gateway that was made by hand
// (unmanaged): a static lease and/or a rule that grants Internet by MAC.
type Discovered struct {
	MAC      string
	IP       string // "" when only a firewall rule exists
	Name     string // the comment found on the gateway, may be empty
	Internet bool   // an enabled hand-made rule grants this MAC the Internet
}

// Skipped is a Discovered entry an importer did not turn into a host.
type Skipped struct {
	Found  Discovered
	Reason string
}

// ImportResult reports what HostImporter.ImportHosts did.
type ImportResult struct {
	Created int
	Skipped []Skipped
}
