---
PLAN: "feat: network access contract, in-memory gateway and conformance suite"
TAG: v0.1.0
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 13863358207458738616
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> Phase F2 of the network administration master plan (in the private repo
> `veltylabs/mjosefa-cms` — you do not need it). This repo is a fresh `gonew`: there is nothing to
> reference by import; this plan inlines the exact code.

# Plan — `webtyp.com/network`: the network-access port

Read first (in this repo): [AGENTS.md](../AGENTS.md) and [docs/ARCHITECTURE.md](ARCHITECTURE.md).
They define the vocabulary (host, access, managed/unmanaged, adoption, conflict, warning,
fingerprint) used below.

## Development rules (apply to every line you write)

- **Isomorphic**: every non-test file must build with `GOOS=js GOARCH=wasm go build ./...` and TinyGo.
- **Only `webtyp.com/fmt`** (`fmt.Err`, `fmt.Errf`, `fmt.Sprintf`, `fmt.Convert(...)`): no `errors`,
  `strings`, `strconv`, stdlib `fmt` — except `conformance/`, which may import `testing`.
- **No Go `map`** anywhere, tests included. **No `reflect`.**
- **No string literals in logic** — use the constants and error variables declared in Stage 1.
- **Tests in `tests/`** (package `tests`, external, public API only). Never export a symbol only for a
  test. Runner: `gotest ./...`.
- No `TODO`, no commented-out code, no stubs.

## Design gate

**1. Prior art.**
- **Terraform** (`plan` / `apply`, saved plan files, refusing a stale plan): the operator reviews an
  exact diff, applies exactly that diff. Adopted: `Plan` + `Apply(desired, fingerprint)` +
  `ErrPlanStale`.
- **Ansible** network modules (`--check --diff`, idempotent tasks): running twice changes nothing.
  Adopted: after `Apply`, `Plan` of the same desired state is empty — enforced by the conformance
  suite.
- **NetBox** (source-of-truth inventory) + a separate sync/provisioning tool: inventory and enforcement
  are different concerns joined by a contract. Adopted: `HostSource` / `Gateway`.
- **`database/sql/driver` and `webtyp/storage`**: a port + reference implementation + conformance
  suite, concrete drivers in their own repos. Adopted structure: root, `mem`, `conformance`.

**2. Novice-name test.** `network.Host`, `network.AccessInternet`, `gateway.Plan(desired)`,
`gateway.Apply(desired, plan.Fingerprint)`, `gateway.Connections()`, `inventory.Hosts()` — each reads
as a sentence. `Discover` / `Discovered` name "what is already configured on the router by hand".

**3. Complexity ledger.**
```
Concepts the developer must learn   +1 package (Host, Access, Settings, Plan, Gateway)
Files they must touch to do X       registering a device: 0 router files (was: Winbox + 1 rule per device)
Lines at the call site              plan, _ := gw.Plan(d); gw.Apply(d, plan.Fingerprint) — 2
Ways to do the same thing           0 (no network contract exists in the ecosystem)
```

**4. Where it belongs.** A port between two libraries (an inventory module and a network management
module) and the vendor implementations: per the ecosystem rule "a missing contract at a boundary is a
defect upstream", it lives in the framework, next to `storage` and `router`, not in either consumer.
No I/O here; vendors in their own repos (`veltylabs/mikrotik`).

**5. What it deletes.** Nothing in code — new capability. Operationally it replaces per-device
firewall rules maintained by hand.

## Stage 1 — root package (`network.go`, `access.go`, `plan.go`, `errors.go`)

Delete the `gonew` stub `network.go` content (`type Network struct{}`, `New()`) and replace it with
the following. Split across files exactly as named.

### `access.go`

```go
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
func (a Access) String() string

// ParseAccess is the inverse of String; an unknown name returns ErrUnknownAccess.
func ParseAccess(name string) (Access, error)

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
func (u Unregistered) String() string

// ParseUnregistered is the inverse of String; an unknown name returns ErrUnknownUnregistered.
func ParseUnregistered(name string) (Unregistered, error)
```

### `network.go`

```go
package network

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
	DHCPServer   string       // name of the gateway's DHCP server that serves the hosts
	DynamicPool  string       // pool for unregistered devices; required only with UnregisteredLocal
	FilterDNS    string       // IPv4 of the resolver forced on AccessInternetFiltered hosts
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
func (d Desired) Validate() error

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
```

### `plan.go`

```go
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

// Fingerprint identifies one exact plan against one exact gateway state.
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
```

### `errors.go`

```go
package network

import "webtyp.com/fmt"

var (
	ErrPlanStale     = fmt.Err("network: the gateway changed since the plan was made, plan again")
	ErrConflicts     = fmt.Err("network: the plan has conflicts, resolve them before applying")
	ErrUnknownAccess = fmt.Err("network: unknown access level")
	ErrUnknownUnregistered = fmt.Err("network: unknown policy for unregistered devices")
	ErrInvalid       = fmt.Err("network: invalid desired state")
)
```
`Desired.Validate` returns an error whose text starts with `ErrInvalid`'s text followed by `": "` and
the field at fault, e.g. `network: invalid desired state: duplicate MAC 48:F1:7F:D9:D7:B7` — build it
with `fmt.Errf`. Callers compare with a helper you also add:

```go
// IsInvalid reports whether err came from Desired.Validate.
func IsInvalid(err error) bool
```
(implement by prefix comparison of `err.Error()` with `ErrInvalid.Error()` using `webtyp.com/fmt`).

## Stage 2 — `mem/` (package `mem`): the reference gateway

File `mem/gateway.go`. Imports: `webtyp.com/network`, `webtyp.com/fmt` only.

```go
// Gateway is an in-memory network.Gateway: the reference behaviour every
// real implementation must match (network/conformance), and the gateway a
// consumer's tests use. Safe for one goroutine.
type Gateway struct { /* unexported fields */ }

func New() *Gateway

var _ network.Gateway = (*Gateway)(nil)

// Connect simulates a device coming online (shown by Connections).
func (g *Gateway) Connect(c network.Connection)

// AddUnmanaged simulates configuration made by hand on the gateway.
// It changes the gateway state (so previous fingerprints become stale).
func (g *Gateway) AddUnmanaged(d network.Discovered)
```

State (unexported): `settings network.Settings`, `hasSettings bool`, `managed []network.Host`,
`unmanaged []network.Discovered`, `connections []network.Connection`, `version int` (incremented by
`Apply` when it changes something and by `AddUnmanaged`).

`Plan(d)` algorithm, in this order (deterministic output order matters for the fingerprint):
1. `d.Validate()`; error → return it.
2. Settings: if `!hasSettings` → `Change{Kind: ChangeAdd, Object: ObjectSettings, After: summary}`;
   else if different → `ChangeUpdate` with `Before`/`After`. `ObjectSettings = "settings"` is an
   exported constant in `mem`. Summary format: `fmt.Sprintf("dhcp=%s pool=%s dns=%s unregistered=%d", …)`.
3. For each desired host, in the order given:
   - managed host with same MAC: equal → nothing; different IP/Name/Access → `ChangeUpdate`.
   - else unmanaged entry with same MAC → `ChangeAdopt`.
   - else → `ChangeAdd`.
   - Conflict when an **unmanaged** entry with a **different** MAC has the same IP:
     `Conflict{Host: h, Reason: fmt.Sprintf("IP %s is held by unmanaged %s", ip, mac)}`.
   `Object` for host changes: `fmt.Sprintf("host %s", h.MAC)`.
4. For each managed host whose MAC is not desired → `ChangeRemove`.
5. Warnings: each unmanaged entry with `Internet == true` whose MAC is not desired →
   `Warning{MAC, Name, Reason: ReasonUnregisteredInternet}` with exported
   `ReasonUnregisteredInternet = "has Internet by a hand-made rule but is not registered"`.
6. `Fingerprint`: `fmt.Sprintf("%d:", version)` followed by every change rendered as
   `kind|object|after;` — equal plans on an equal state give equal fingerprints.

`Apply(d, expected)`: `p, err := Plan(d)`; error → return; `p.Fingerprint != expected` →
`network.ErrPlanStale`; `len(p.Conflicts) > 0` → `network.ErrConflicts`; then set settings, replace
`managed` with the desired hosts, drop adopted MACs from `unmanaged`, `version++` if `!p.Empty()`;
return `p`.

`Connections()` returns a copy of `connections`. `Discover()` returns a copy of `unmanaged`.

## Stage 3 — `conformance/` (package `conformance`)

File `conformance/conformance.go`:

```go
// Fixture is what an implementation provides to be tested.
type Fixture interface {
	Gateway() network.Gateway
	// Settings valid on this gateway (e.g. the name of a DHCP server that exists).
	Settings() network.Settings
	// AddUnmanaged creates hand-made configuration directly on the gateway,
	// bypassing the contract.
	AddUnmanaged(found network.Discovered) error
}

// Run executes every case as a t.Run subtest, each on a fresh fixture.
func Run(t *testing.T, newFixture func(t *testing.T) Fixture)
```

Cases (subtest names exactly as quoted). Hosts used: `h1 = {Name: "PC 1", MAC: "02:00:00:00:00:01",
IP: <ip1>, Access: AccessLocal}`, `h2 = {… "02:00:00:00:00:02", <ip2>, AccessInternet}` where
`ip1`/`ip2` are `"10.99.0.11"`/`"10.99.0.12"` (exported constants `HostIP1`, `HostIP2` so an
implementation's fixture can make them valid on its DHCP network).

| Subtest | Steps | Assertions |
|---|---|---|
| `"plan then apply is idempotent"` | Plan {settings, h1, h2}; Apply with its fingerprint; Plan again | 1st plan contains a `ChangeAdd` whose `Host.MAC` is h1's and one for h2, and at least one site-wide change (zero `Host`) — **never assert the total count**: a real gateway needs several site-wide objects (`mem` has 1); Apply returns no error; 2nd plan `Empty()` |
| `"stale plan is refused"` | Plan {settings, h1}; `AddUnmanaged{MAC: "02:00:00:00:00:09", IP: "10.99.0.19"}`; Apply with old fingerprint | `err == network.ErrPlanStale`; a new Plan still shows the add for h1 |
| `"access change is one update"` | apply {h1}; desired h1 with `AccessInternet` | exactly 1 change, `ChangeUpdate`, `Host.MAC == h1.MAC`; apply; plan empty |
| `"removed host"` | apply {h1, h2}; desired {h1} | exactly 1 change, `ChangeRemove` for h2; apply; plan empty |
| `"adoption"` | `AddUnmanaged{MAC: h1.MAC, IP: h1.IP, Name: "old"}`; plan {h1} | change for h1 is `ChangeAdopt`; apply; `Discover()` no longer lists h1.MAC; plan empty |
| `"ip conflict blocks apply"` | `AddUnmanaged{MAC: "02:00:00:00:00:09", IP: h1.IP}`; plan {h1} | 1 conflict; `Apply` → `network.ErrConflicts`; `Discover()` unchanged |
| `"unregistered internet is a warning"` | `AddUnmanaged{MAC: "02:00:00:00:00:08", Internet: true}`; plan {h1} | 1 warning with that MAC; no conflict; apply succeeds |
| `"invalid desired"` | plan {h1, h1-with-other-IP} (duplicate MAC) | `network.IsInvalid(err)` |
| `"discover lists unmanaged"` | `AddUnmanaged{MAC: "02:00:00:00:00:07", IP: "10.99.0.17", Name: "printer"}` | `Discover()` contains an entry with those three values |

`Connections()` is deliberately **not** in the suite (a real gateway cannot fake a device being
online); each implementation tests it on its own.

## Stage 4 — tests (`tests/`)

- `tests/conformance_test.go`: `conformance.Run(t, …)` over a fixture wrapping `mem.New()`
  (`Settings()` returns `{DHCPServer: "dhcp1", FilterDNS: "1.1.1.3", Unregistered: network.UnregisteredNoAddress}`,
  `AddUnmanaged` calls `g.AddUnmanaged`).
- `tests/access_test.go`: `String`/`ParseAccess` round-trip for the 3 levels; unknown name →
  `ErrUnknownAccess`; same for `Unregistered` (`ParseUnregistered`, `ErrUnknownUnregistered`); zero `Access` is `AccessLocal`; zero `Unregistered` is `UnregisteredNoAddress`.
- `tests/validate_test.go`: each `Desired.Validate` rule (missing DHCPServer, missing FilterDNS,
  `UnregisteredLocal` without pool, empty MAC, duplicate MAC, duplicate IP) → `IsInvalid(err)`.
- `tests/connections_test.go`: `mem` `Connect` then `Connections()` returns it.

## Stage 5 — docs

`README.md` (replace the gonew stub) with: one-paragraph description (copy the first paragraph of
ARCHITECTURE "What this is"), an "I want X → use Y" table:

| I want to… | Use |
|---|---|
| Describe what the router must enforce | `network.Desired{Settings, Hosts}` |
| See what would change | `gw.Plan(desired)` |
| Apply exactly what I reviewed | `gw.Apply(desired, plan.Fingerprint)` |
| See who is online | `gw.Connections()` |
| Import hand-made router config | `gw.Discover()` → `inventory.ImportHosts(found)` |
| Test my code without a router | `mem.New()` |
| Prove my gateway implementation | `conformance.Run(t, newFixture)` |

and links to `AGENTS.md` and `docs/ARCHITECTURE.md`. Verify `docs/ARCHITECTURE.md` matches the code;
fix the doc if a name differs.

## Acceptance criteria

- `gotest ./...` green; `GOOS=js GOARCH=wasm go build ./...` succeeds for root and `mem`.
- `grep -rn "map\[" --include=*.go .` → empty.
- `grep -rn '"errors"\|"strings"\|"strconv"' --include=*.go . | grep -v conformance` → empty.
- `grep -rn "type Network struct\|func New() \*Network" .` → empty (gonew stub removed).

| Stage | Files | Done when |
|---|---|---|
| 1 | `access.go`, `network.go`, `plan.go`, `errors.go` | compiles for wasm |
| 2 | `mem/gateway.go` | compiles, implements `network.Gateway` |
| 3 | `conformance/conformance.go` | 9 subtests |
| 4 | `tests/*.go` | green |
| 5 | `README.md`, verify `docs/ARCHITECTURE.md` | table present |
