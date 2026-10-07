# `webtyp/network` — architecture

## What this is

`network` is the **contract between an inventory of devices and the box that enforces network
access** (a router). It answers three questions an application has about its local network:

1. *Which devices may use the network, and how much of it?* — a list of `Host` (one per network
   card: MAC, IP, name) with an `Access` level.
2. *What would change on the router if I applied that list, and is it safe now?* — `Plan`.
3. *Who is connected right now?* — `Connections`.

You meet it when an application lets an administrator register a PC or a printer and have the router
follow: give it a fixed IP, allow or deny the Internet, filter its DNS — without logging into the
router by hand.

It is the network equivalent of `webtyp/storage`: a port with no I/O of its own, an in-memory
reference implementation (`network/mem`) and an executable conformance suite
(`network/conformance`). Concrete implementations live in their own repositories
(`github.com/veltylabs/mikrotik` for MikroTik RouterOS).

## Vocabulary

| Term | Meaning here |
|---|---|
| **Host** | One network interface of a registered device: a PC with Wi-Fi and Ethernet is two hosts |
| **Access** | What a host may reach: `AccessLocal` (only the local network), `AccessInternetFiltered` (plus the Internet, with DNS forced through a filtering resolver), `AccessInternet` (plus the Internet, unfiltered) |
| **Gateway** | The device that enforces access — usually the router: DHCP, firewall, DNS redirection |
| **Managed object** | Something on the gateway this contract created (a DHCP lease, a firewall rule). Implementations mark them; they never touch anything else except by **adoption** |
| **Unmanaged object** | Configuration made by hand on the gateway |
| **Adoption** | A registered host whose MAC already has an unmanaged lease: applying the plan turns that lease into a managed one. Shown in the plan as `ChangeAdopt` |
| **Conflict** | A change that cannot be made without breaking hand-made configuration (another MAC holds the IP). Blocks `Apply` |
| **Warning** | Something the administrator should know but that does not block (a device with Internet by a hand-made rule that is not registered) |
| **Fingerprint** | An opaque token identifying one exact plan against one exact gateway state |

## Decisions

### Closed by default

- A device that is not a `Host` gets nothing that the settings do not explicitly grant:
  `Settings.Unregistered` zero value is `UnregisteredNoAddress` (no IP at all).
  `UnregisteredLocal` (gets an IP, local network only) must be chosen explicitly — it exists for the
  migration period, while devices are still being registered.
- A registered host's zero `Access` is `AccessLocal`; reaching the Internet is an explicit level.

### Plan, then apply — never apply blindly

Three ways to push an inventory to a router were considered:

| Option | Rejected / chosen because |
|---|---|
| Apply on every save | A typo cuts the network at that moment, possibly mid-consultation. The operator cannot choose the time |
| Periodic sync | Same, plus nobody knows *when* it happened, and it silently overwrites an emergency fix made by hand |
| **Plan, review, apply** (chosen) | The operator sees every change — and, with `Connections`, who is online and would be affected — and picks the moment |

This is `terraform plan` / `terraform apply`, and Ansible's `--check --diff`. `Apply` takes the
`Fingerprint` of the plan the operator reviewed; if the gateway changed in between (someone edited it
by hand, another operator applied) it refuses with `ErrPlanStale` instead of applying something nobody
reviewed.

### Implementations own only what they created

Hand-made configuration is reported (`Discover`, `Warnings`, `Conflicts`), never modified — except
adoption, which is explicit in the plan. This makes the contract safe to introduce on a router that
has years of manual configuration.

### The inventory side is part of the contract

`HostSource` (an inventory lists its hosts) and `HostImporter` (an inventory creates hosts from what
the gateway discovered) are declared **here**, not in the consumer: two libraries meet at this
boundary (an inventory module and a network management module), and the type that crosses it must be
named upstream, once.

### Interfaces are narrow

`Planner`, `Applier`, `ConnectionReader`, `Discoverer` are separate; `Gateway` composes them. A
dashboard that only shows who is connected depends on `ConnectionReader` and nothing else.

## Out of scope (future extensions of this contract)

- Wi-Fi networks, access points and their channels (centralized AP management).
- VLANs / a guest network.
- "Only these domains" access level.

Each arrives as new types and a new narrow interface; existing consumers do not change.
