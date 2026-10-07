package conformance

import (
	"testing"

	"webtyp.com/network"
)

const HostIP1 = "10.99.0.11"
const HostIP2 = "10.99.0.12"

// Fixture is what an implementation provides to be tested.
type Fixture interface {
	Gateway() network.Gateway
	// Settings valid on this gateway (e.g. the name of a DHCP server that exists).
	Settings() network.Settings
	// AddUnmanaged creates hand-made configuration directly on the gateway,
	// bypassing the contract.
	AddUnmanaged(found network.Discovered) error
}

func getHosts() (network.Host, network.Host) {
	h1 := network.Host{
		Name:   "PC 1",
		MAC:    "02:00:00:00:00:01",
		IP:     HostIP1,
		Access: network.AccessLocal,
	}
	h2 := network.Host{
		Name:   "PC 2",
		MAC:    "02:00:00:00:00:02",
		IP:     HostIP2,
		Access: network.AccessInternet,
	}
	return h1, h2
}

// Run executes every case as a t.Run subtest, each on a fresh fixture.
func Run(t *testing.T, newFixture func(t *testing.T) Fixture) {
	t.Run("plan then apply is idempotent", func(t *testing.T) {
		f := newFixture(t)
		h1, h2 := getHosts()
		d := network.Desired{
			Settings: f.Settings(),
			Hosts:    []network.Host{h1, h2},
		}

		gw := f.Gateway()
		p1, err := gw.Plan(d)
		if err != nil {
			t.Fatalf("Plan failed: %v", err)
		}

		var hasAddH1, hasAddH2, hasSiteWide bool
		for _, c := range p1.Changes {
			if c.Kind == network.ChangeAdd && c.Host.MAC == h1.MAC {
				hasAddH1 = true
			}
			if c.Kind == network.ChangeAdd && c.Host.MAC == h2.MAC {
				hasAddH2 = true
			}
			if c.Host.MAC == "" {
				hasSiteWide = true
			}
		}

		if !hasAddH1 || !hasAddH2 || !hasSiteWide {
			t.Errorf("Expected changes for h1, h2, and sitewide, got: %v", p1.Changes)
		}

		_, err = gw.Apply(d, p1.Fingerprint)
		if err != nil {
			t.Fatalf("Apply failed: %v", err)
		}

		p2, err := gw.Plan(d)
		if err != nil {
			t.Fatalf("Second Plan failed: %v", err)
		}

		if !p2.Empty() {
			t.Errorf("Expected empty plan, got %d changes", len(p2.Changes))
		}
	})

	t.Run("stale plan is refused", func(t *testing.T) {
		f := newFixture(t)
		h1, _ := getHosts()
		d := network.Desired{
			Settings: f.Settings(),
			Hosts:    []network.Host{h1},
		}

		gw := f.Gateway()
		p1, err := gw.Plan(d)
		if err != nil {
			t.Fatalf("Plan failed: %v", err)
		}

		err = f.AddUnmanaged(network.Discovered{
			MAC: "02:00:00:00:00:09",
			IP:  "10.99.0.19",
		})
		if err != nil {
			t.Fatalf("AddUnmanaged failed: %v", err)
		}

		_, err = gw.Apply(d, p1.Fingerprint)
		if err != network.ErrPlanStale {
			t.Errorf("Expected ErrPlanStale, got %v", err)
		}

		p2, err := gw.Plan(d)
		if err != nil {
			t.Fatalf("Second Plan failed: %v", err)
		}

		var hasAddH1 bool
		for _, c := range p2.Changes {
			if c.Kind == network.ChangeAdd && c.Host.MAC == h1.MAC {
				hasAddH1 = true
			}
		}
		if !hasAddH1 {
			t.Errorf("Expected add for h1 in new plan, got: %v", p2.Changes)
		}
	})

	t.Run("access change is one update", func(t *testing.T) {
		f := newFixture(t)
		h1, _ := getHosts()
		d := network.Desired{
			Settings: f.Settings(),
			Hosts:    []network.Host{h1},
		}

		gw := f.Gateway()
		p, _ := gw.Plan(d)
		_, err := gw.Apply(d, p.Fingerprint)
		if err != nil {
			t.Fatalf("Initial apply failed: %v", err)
		}

		h1Mod := h1
		h1Mod.Access = network.AccessInternet
		d2 := network.Desired{
			Settings: f.Settings(),
			Hosts:    []network.Host{h1Mod},
		}

		p2, err := gw.Plan(d2)
		if err != nil {
			t.Fatalf("Second Plan failed: %v", err)
		}

		if len(p2.Changes) != 1 {
			t.Fatalf("Expected 1 change, got %d", len(p2.Changes))
		}

		c := p2.Changes[0]
		if c.Kind != network.ChangeUpdate || c.Host.MAC != h1.MAC {
			t.Errorf("Expected update for h1, got %v", c)
		}

		_, err = gw.Apply(d2, p2.Fingerprint)
		if err != nil {
			t.Fatalf("Second Apply failed: %v", err)
		}

		p3, _ := gw.Plan(d2)
		if !p3.Empty() {
			t.Errorf("Expected empty plan after update")
		}
	})

	t.Run("removed host", func(t *testing.T) {
		f := newFixture(t)
		h1, h2 := getHosts()
		d := network.Desired{
			Settings: f.Settings(),
			Hosts:    []network.Host{h1, h2},
		}

		gw := f.Gateway()
		p, _ := gw.Plan(d)
		_, err := gw.Apply(d, p.Fingerprint)
		if err != nil {
			t.Fatalf("Initial apply failed: %v", err)
		}

		d2 := network.Desired{
			Settings: f.Settings(),
			Hosts:    []network.Host{h1},
		}

		p2, err := gw.Plan(d2)
		if err != nil {
			t.Fatalf("Second Plan failed: %v", err)
		}

		if len(p2.Changes) != 1 {
			t.Fatalf("Expected 1 change, got %d", len(p2.Changes))
		}

		c := p2.Changes[0]
		if c.Kind != network.ChangeRemove || c.Host.MAC != h2.MAC {
			t.Errorf("Expected remove for h2, got %v", c)
		}

		_, err = gw.Apply(d2, p2.Fingerprint)
		if err != nil {
			t.Fatalf("Second Apply failed: %v", err)
		}

		p3, _ := gw.Plan(d2)
		if !p3.Empty() {
			t.Errorf("Expected empty plan after remove")
		}
	})

	t.Run("adoption", func(t *testing.T) {
		f := newFixture(t)
		h1, _ := getHosts()

		err := f.AddUnmanaged(network.Discovered{
			MAC:  h1.MAC,
			IP:   h1.IP,
			Name: "old",
		})
		if err != nil {
			t.Fatalf("AddUnmanaged failed: %v", err)
		}

		d := network.Desired{
			Settings: f.Settings(),
			Hosts:    []network.Host{h1},
		}

		gw := f.Gateway()
		p, err := gw.Plan(d)
		if err != nil {
			t.Fatalf("Plan failed: %v", err)
		}

		var hasAdoptH1 bool
		for _, c := range p.Changes {
			if c.Kind == network.ChangeAdopt && c.Host.MAC == h1.MAC {
				hasAdoptH1 = true
			}
		}
		if !hasAdoptH1 {
			t.Errorf("Expected ChangeAdopt for h1, got: %v", p.Changes)
		}

		_, err = gw.Apply(d, p.Fingerprint)
		if err != nil {
			t.Fatalf("Apply failed: %v", err)
		}

		unmanaged, _ := gw.Discover()
		for _, u := range unmanaged {
			if u.MAC == h1.MAC {
				t.Errorf("Expected h1.MAC to be removed from unmanaged")
			}
		}

		p2, _ := gw.Plan(d)
		if !p2.Empty() {
			t.Errorf("Expected empty plan after adoption")
		}
	})

	t.Run("ip conflict blocks apply", func(t *testing.T) {
		f := newFixture(t)
		h1, _ := getHosts()

		err := f.AddUnmanaged(network.Discovered{
			MAC: "02:00:00:00:00:09",
			IP:  h1.IP,
		})
		if err != nil {
			t.Fatalf("AddUnmanaged failed: %v", err)
		}

		d := network.Desired{
			Settings: f.Settings(),
			Hosts:    []network.Host{h1},
		}

		gw := f.Gateway()
		p, err := gw.Plan(d)
		if err != nil {
			t.Fatalf("Plan failed: %v", err)
		}

		if len(p.Conflicts) != 1 {
			t.Fatalf("Expected 1 conflict, got %d", len(p.Conflicts))
		}

		_, err = gw.Apply(d, p.Fingerprint)
		if err != network.ErrConflicts {
			t.Errorf("Expected ErrConflicts, got %v", err)
		}

		unmanaged, _ := gw.Discover()
		if len(unmanaged) != 1 || unmanaged[0].MAC != "02:00:00:00:00:09" {
			t.Errorf("Expected unmanaged to be unchanged")
		}
	})

	t.Run("unregistered internet is a warning", func(t *testing.T) {
		f := newFixture(t)
		h1, _ := getHosts()

		err := f.AddUnmanaged(network.Discovered{
			MAC:      "02:00:00:00:00:08",
			Internet: true,
		})
		if err != nil {
			t.Fatalf("AddUnmanaged failed: %v", err)
		}

		d := network.Desired{
			Settings: f.Settings(),
			Hosts:    []network.Host{h1},
		}

		gw := f.Gateway()
		p, err := gw.Plan(d)
		if err != nil {
			t.Fatalf("Plan failed: %v", err)
		}

		if len(p.Warnings) != 1 || p.Warnings[0].MAC != "02:00:00:00:00:08" {
			t.Fatalf("Expected 1 warning for 02:00:00:00:00:08, got %v", p.Warnings)
		}
		if len(p.Conflicts) > 0 {
			t.Errorf("Expected 0 conflicts, got %d", len(p.Conflicts))
		}

		_, err = gw.Apply(d, p.Fingerprint)
		if err != nil {
			t.Fatalf("Apply failed: %v", err)
		}
	})

	t.Run("invalid desired", func(t *testing.T) {
		f := newFixture(t)
		h1, _ := getHosts()

		h2 := h1
		h2.IP = "10.99.0.99"

		d := network.Desired{
			Settings: f.Settings(),
			Hosts:    []network.Host{h1, h2},
		}

		gw := f.Gateway()
		_, err := gw.Plan(d)
		if !network.IsInvalid(err) {
			t.Errorf("Expected IsInvalid to be true, got err: %v", err)
		}
	})

	t.Run("discover lists unmanaged", func(t *testing.T) {
		f := newFixture(t)

		err := f.AddUnmanaged(network.Discovered{
			MAC:  "02:00:00:00:00:07",
			IP:   "10.99.0.17",
			Name: "printer",
		})
		if err != nil {
			t.Fatalf("AddUnmanaged failed: %v", err)
		}

		gw := f.Gateway()
		unmanaged, _ := gw.Discover()

		var found bool
		for _, u := range unmanaged {
			if u.MAC == "02:00:00:00:00:07" && u.IP == "10.99.0.17" && u.Name == "printer" {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected to find unmanaged entry, got: %v", unmanaged)
		}
	})
}
