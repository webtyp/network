# AGENTS.md — webtyp/network

Working notes for AI agents operating in this library. Design and rationale:
[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md). End-user docs: [README.md](README.md).

## Mission

`network` is the **network-access port** of the webtyp ecosystem: the types that cross between a
device inventory and a gateway (`Host`, `Access`, `Settings`, `Plan`, `Connection`, `Discovered`),
the interfaces on both sides (`Gateway` and its parts; `HostSource`, `HostImporter`), an in-memory
reference implementation (`network/mem`) and an executable conformance suite
(`network/conformance`). It performs **no I/O**.

## Rules (do not violate)

- **Isomorphic.** The root package and `mem` are imported by domain modules that compile to WASM and
  TinyGo. Every file must build under `GOOS=js GOARCH=wasm` and TinyGo.
- **Only `webtyp.com/fmt`** for strings, numbers and errors: no `errors`, `strings`, `strconv`,
  stdlib `fmt`. `conformance` may import `testing`.
- **No Go `map`**, no `reflect`. Use slices scanned linearly (collections here are a site's devices —
  tens, not thousands).
- **No concrete gateway here.** Anything that talks to a real router (sockets, `net`, a vendor SDK)
  is a separate repository that implements `Gateway` and passes `conformance.Run`.
- **Closed by default** (see ARCHITECTURE): zero values deny.
- **No string literals in logic**: names, errors and markers are exported constants or package-level
  errors.
- Tests live in `tests/` (external package), run with `gotest`.
