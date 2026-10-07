# webtyp/network
<img src="docs/img/badges.svg">

`network` is the **contract between an inventory of devices and the box that enforces network access** (a router). It answers three questions an application has about its local network:
1. *Which devices may use the network, and how much of it?* — a list of `Host` (one per network card: MAC, IP, name) with an `Access` level.
2. *What would change on the router if I applied that list, and is it safe now?* — `Plan`.
3. *Who is connected right now?* — `Connections`.

| I want to… | Use |
|---|---|
| Describe what the router must enforce | `network.Desired{Settings, Hosts}` |
| See what would change | `gw.Plan(desired)` |
| Apply exactly what I reviewed | `gw.Apply(desired, plan.Fingerprint)` |
| See who is online | `gw.Connections()` |
| Import hand-made router config | `gw.Discover()` → `inventory.ImportHosts(found)` |
| Test my code without a router | `mem.New()` |
| Prove my gateway implementation | `conformance.Run(t, newFixture)` |

See [AGENTS.md](AGENTS.md) and [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for design and vocabulary.
