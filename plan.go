package network

type ChangeKind uint8

const (
	ChangeAdd    ChangeKind = iota + 1 // create a managed object
	ChangeUpdate                       // modify a managed object
	ChangeRemove                       // delete a managed object
	ChangeAdopt                        // turn an unmanaged object into a managed one
)

// Change is one step of a Plan, described for a human reviewer.
type Change struct {
	Kind   ChangeKind
	Object string // what is touched, in the gateway's words, e.g. "dhcp lease 172.0.0.61"
	Host   Host   // the host the change is for; zero Host for a site-wide change
	Before string // current state summary; "" for ChangeAdd
	After  string // desired state summary; "" for ChangeRemove
}

// Conflict blocks Apply: the change would break unmanaged configuration.
type Conflict struct {
	Host   Host
	Reason string
}

// Warning does not block Apply.
type Warning struct {
	MAC    string
	Name   string
	Reason string
}

// Fingerprint identifies one exact plan against one exact gateway state. It is
// opaque but always non-empty lower-case hex ([0-9a-f]+), so consumers can carry
// it through text fields and URLs without escaping.
type Fingerprint string

type Plan struct {
	Changes     []Change
	Conflicts   []Conflict
	Warnings    []Warning
	Fingerprint Fingerprint
}

// Empty reports whether applying the plan would change nothing.
func (p Plan) Empty() bool { return len(p.Changes) == 0 }

// ---- gateway side ----

type Planner interface {
	// Plan validates d (Desired.Validate) and compares it with the gateway.
	Plan(d Desired) (Plan, error)
}

type Applier interface {
	// Apply re-plans d; if the result's Fingerprint differs from expected it
	// returns ErrPlanStale and changes nothing; if it has Conflicts it returns
	// ErrConflicts and changes nothing; otherwise it applies every Change and
	// returns the plan it applied.
	Apply(d Desired, expected Fingerprint) (Plan, error)
}

type ConnectionReader interface {
	Connections() ([]Connection, error)
}

type Discoverer interface {
	// Discover returns the unmanaged configuration, one entry per MAC.
	Discover() ([]Discovered, error)
}

// Gateway is a complete implementation (e.g. github.com/veltylabs/mikrotik).
type Gateway interface {
	Planner
	Applier
	ConnectionReader
	Discoverer
}

// ---- inventory side ----

// HostSource is implemented by an inventory: the hosts that must exist.
type HostSource interface {
	Hosts() ([]Host, error)
}

// HostImporter is implemented by an inventory: create hosts from what the
// gateway discovered (one-time migration from hand-made configuration).
type HostImporter interface {
	ImportHosts(found []Discovered) (ImportResult, error)
}
