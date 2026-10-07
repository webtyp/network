package mem

import (
	"webtyp.com/fmt"
	"webtyp.com/network"
)

const ObjectSettings = "settings"
const ReasonUnregisteredInternet = "has Internet by a hand-made rule but is not registered"

// Gateway is an in-memory network.Gateway: the reference behaviour every
// real implementation must match (network/conformance), and the gateway a
// consumer's tests use. Safe for one goroutine.
type Gateway struct {
	settings    network.Settings
	hasSettings bool
	managed     []network.Host
	unmanaged   []network.Discovered
	connections []network.Connection
	version     int
}

func New() *Gateway {
	return &Gateway{}
}

var _ network.Gateway = (*Gateway)(nil)

func (g *Gateway) Plan(d network.Desired) (network.Plan, error) {
	if err := d.Validate(); err != nil {
		return network.Plan{}, err
	}

	var changes []network.Change
	var conflicts []network.Conflict
	var warnings []network.Warning

	summary := fmt.Sprintf("dhcp=%s pool=%s dns=%s unregistered=%s", d.Settings.DHCPServer, d.Settings.DynamicPool, d.Settings.FilterDNS, d.Settings.Unregistered.String())
	if !g.hasSettings {
		changes = append(changes, network.Change{
			Kind:   network.ChangeAdd,
			Object: ObjectSettings,
			After:  summary,
		})
	} else if g.settings != d.Settings {
		beforeSummary := fmt.Sprintf("dhcp=%s pool=%s dns=%s unregistered=%s", g.settings.DHCPServer, g.settings.DynamicPool, g.settings.FilterDNS, g.settings.Unregistered.String())
		changes = append(changes, network.Change{
			Kind:   network.ChangeUpdate,
			Object: ObjectSettings,
			Before: beforeSummary,
			After:  summary,
		})
	}

	for _, dh := range d.Hosts {
		var foundManaged bool
		for _, mh := range g.managed {
			if mh.MAC == dh.MAC {
				foundManaged = true
				if mh.IP != dh.IP || mh.Name != dh.Name || mh.Access != dh.Access {
					changes = append(changes, network.Change{
						Kind:   network.ChangeUpdate,
						Object: fmt.Sprintf("host %s", dh.MAC),
						Host:   dh,
						Before: fmt.Sprintf("ip=%s name=%s access=%s", mh.IP, mh.Name, mh.Access.String()),
						After:  fmt.Sprintf("ip=%s name=%s access=%s", dh.IP, dh.Name, dh.Access.String()),
					})
				}
				break
			}
		}

		if !foundManaged {
			var foundUnmanaged bool
			for _, uh := range g.unmanaged {
				if uh.MAC == dh.MAC {
					foundUnmanaged = true
					changes = append(changes, network.Change{
						Kind:   network.ChangeAdopt,
						Object: fmt.Sprintf("host %s", dh.MAC),
						Host:   dh,
						Before: fmt.Sprintf("ip=%s name=%s internet=%t", uh.IP, uh.Name, uh.Internet),
						After:  fmt.Sprintf("ip=%s name=%s access=%s", dh.IP, dh.Name, dh.Access.String()),
					})
					break
				}
			}
			if !foundUnmanaged {
				changes = append(changes, network.Change{
					Kind:   network.ChangeAdd,
					Object: fmt.Sprintf("host %s", dh.MAC),
					Host:   dh,
					After:  fmt.Sprintf("ip=%s name=%s access=%s", dh.IP, dh.Name, dh.Access.String()),
				})
			}
		}

		for _, uh := range g.unmanaged {
			if uh.MAC != dh.MAC && uh.IP == dh.IP {
				conflicts = append(conflicts, network.Conflict{
					Host:   dh,
					Reason: fmt.Sprintf("IP %s is held by unmanaged %s", dh.IP, uh.MAC),
				})
			}
		}
	}

	for _, mh := range g.managed {
		var desired bool
		for _, dh := range d.Hosts {
			if mh.MAC == dh.MAC {
				desired = true
				break
			}
		}
		if !desired {
			changes = append(changes, network.Change{
				Kind:   network.ChangeRemove,
				Object: fmt.Sprintf("host %s", mh.MAC),
				Host:   mh,
				Before: fmt.Sprintf("ip=%s name=%s access=%s", mh.IP, mh.Name, mh.Access.String()),
			})
		}
	}

	for _, uh := range g.unmanaged {
		if uh.Internet {
			var desired bool
			for _, dh := range d.Hosts {
				if uh.MAC == dh.MAC {
					desired = true
					break
				}
			}
			if !desired {
				warnings = append(warnings, network.Warning{
					MAC:    uh.MAC,
					Name:   uh.Name,
					Reason: ReasonUnregisteredInternet,
				})
			}
		}
	}

	fingerprintStr := fmt.Sprintf("%d:", g.version)
	for _, c := range changes {
		fingerprintStr += fmt.Sprintf("%d|%s|%s;", int(c.Kind), c.Object, c.After)
	}

	return network.Plan{
		Changes:     changes,
		Conflicts:   conflicts,
		Warnings:    warnings,
		Fingerprint: network.Fingerprint(fingerprintStr),
	}, nil
}

func (g *Gateway) Apply(d network.Desired, expected network.Fingerprint) (network.Plan, error) {
	p, err := g.Plan(d)
	if err != nil {
		return network.Plan{}, err
	}
	if p.Fingerprint != expected {
		return network.Plan{}, network.ErrPlanStale
	}
	if len(p.Conflicts) > 0 {
		return network.Plan{}, network.ErrConflicts
	}

	g.settings = d.Settings
	g.hasSettings = true

	newManaged := make([]network.Host, len(d.Hosts))
	copy(newManaged, d.Hosts)
	g.managed = newManaged

	var newUnmanaged []network.Discovered
	for _, uh := range g.unmanaged {
		var adopted bool
		for _, dh := range d.Hosts {
			if uh.MAC == dh.MAC {
				adopted = true
				break
			}
		}
		if !adopted {
			newUnmanaged = append(newUnmanaged, uh)
		}
	}
	g.unmanaged = newUnmanaged

	if !p.Empty() {
		g.version++
	}

	return p, nil
}

func (g *Gateway) Connect(c network.Connection) {
	g.connections = append(g.connections, c)
}

func (g *Gateway) AddUnmanaged(d network.Discovered) {
	g.unmanaged = append(g.unmanaged, d)
	g.version++
}

func (g *Gateway) Connections() ([]network.Connection, error) {
	res := make([]network.Connection, len(g.connections))
	copy(res, g.connections)
	return res, nil
}

func (g *Gateway) Discover() ([]network.Discovered, error) {
	res := make([]network.Discovered, len(g.unmanaged))
	copy(res, g.unmanaged)
	return res, nil
}
