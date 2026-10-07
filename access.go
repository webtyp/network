package network

// Access is what a registered host may reach. The zero value is the least
// privilege a registered host can have.
type Access uint8

const (
	AccessLocal            Access = iota // only the local network
	AccessInternetFiltered               // local network + Internet, DNS forced through Settings.FilterDNS
	AccessInternet                       // local network + unrestricted Internet
)

// Stable names, for storage and transport. The ONLY place these literals exist.
const (
	AccessLocalName            = "local"
	AccessInternetFilteredName = "internet_filtered"
	AccessInternetName         = "internet"
)

// String returns the stable name ("" for an undefined value).
func (a Access) String() string {
	switch a {
	case AccessLocal:
		return AccessLocalName
	case AccessInternetFiltered:
		return AccessInternetFilteredName
	case AccessInternet:
		return AccessInternetName
	default:
		return ""
	}
}

// ParseAccess is the inverse of String; an unknown name returns ErrUnknownAccess.
func ParseAccess(name string) (Access, error) {
	switch name {
	case AccessLocalName:
		return AccessLocal, nil
	case AccessInternetFilteredName:
		return AccessInternetFiltered, nil
	case AccessInternetName:
		return AccessInternet, nil
	default:
		return AccessLocal, ErrUnknownAccess
	}
}

// Unregistered decides what a device that is not a Host gets.
// The zero value is the closed one.
type Unregistered uint8

const (
	UnregisteredNoAddress Unregistered = iota // no IP at all
	UnregisteredLocal                          // an IP from Settings.DynamicPool, local network only
)

const (
	UnregisteredNoAddressName = "no_address"
	UnregisteredLocalName     = "local"
)

// String returns the stable name ("" for an undefined value).
func (u Unregistered) String() string {
	switch u {
	case UnregisteredNoAddress:
		return UnregisteredNoAddressName
	case UnregisteredLocal:
		return UnregisteredLocalName
	default:
		return ""
	}
}

// ParseUnregistered is the inverse of String; an unknown name returns ErrUnknownUnregistered.
func ParseUnregistered(name string) (Unregistered, error) {
	switch name {
	case UnregisteredNoAddressName:
		return UnregisteredNoAddress, nil
	case UnregisteredLocalName:
		return UnregisteredLocal, nil
	default:
		return UnregisteredNoAddress, ErrUnknownUnregistered
	}
}
