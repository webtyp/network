---
PLAN: "feat(network): IsPlanStale, IsConflicts — detect apply sentinels without == between interfaces"
EXECUTOR: jules
REVIEWER: none
---

# Plan — `network.IsPlanStale(err)`, `network.IsConflicts(err)`

> Master: `webtyp/docs/NO_REFLECTION_MASTER_PLAN.md` (ola 3, arreglo de raíz). El PR de
> `veltylabs/network_manager` espera este tag: hoy compara mensajes de error porque este paquete no
> ofrece cómo detectar estos dos centinelas.

## 1. El problema

`Gateway.Apply` devuelve `ErrPlanStale` o `ErrConflicts` (lo implementan `mem/gateway.go:165,168` y
`veltylabs/packages/mikrotik/v6/apply.go:16,20`). Quien los detecta hoy escribe `err == network.ErrPlanStale`
(`conformance/conformance.go:113,305`) o, peor, compara textos (`network_manager/ops.go:54`:
`err.Error() == network.ErrPlanStale.Error()`).

En TinyGo, `==` entre dos `error` (interfaces) compila a `runtime.interfaceEqual` →
`reflectValueEqual`, y mete `internal/reflectlite` (~7–9 KB) en el binario wasm. La regla del dueño:
cero reflexión en wasm. Comparar textos evita la reflexión pero es frágil (cualquier cambio de
mensaje rompe la detección en silencio).

## 2. Design gate (api-design)

1. **Antecedentes.** `os.IsNotExist(err)` (biblioteca estándar), `status.Code(err)` de gRPC,
   `apierrors.IsConflict(err)` de Kubernetes client-go (el mismo concepto que `ErrConflicts`). En este
   ecosistema ya existen `storage.IsNoRows`, `orm.IsNotFound`, `rbac.IsRoleNotFound` y, en este mismo
   paquete, `IsInvalid`.
2. **Nombres.** `IsPlanStale(err)`, `IsConflicts(err)`: "Is" + el nombre del centinela sin "Err",
   igual que `IsInvalid`.
3. **Balance.** +2 funciones · formas de detectar: hoy 2 (`==` y comparar textos), después 1.
4. **Dónde va.** `network`, dueño de los centinelas y del contrato `Gateway.Apply`.
5. **Qué borra.** Los `!=` de `conformance/conformance.go:113,305`; en `network_manager`, la
   comparación de textos.

## 3. La corrección

En `errors.go`, los centinelas pasan a un tipo string no exportado (mismo texto exacto):

```go
// domainError is the concrete type of this package's sentinel errors. IsX
// recognises them with a type assertion: TinyGo compiles that to a type-code
// comparison, while == between two error values goes through
// runtime.interfaceEqual and pulls internal/reflectlite into the wasm binary.
type domainError string

func (e domainError) Error() string { return string(e) }

const (
	ErrPlanStale           domainError = "network: the gateway changed since the plan was made, plan again"
	ErrConflicts           domainError = "network: the plan has conflicts, resolve them before applying"
	ErrUnknownAccess       domainError = "network: unknown access level"
	ErrUnknownUnregistered domainError = "network: unknown policy for unregistered devices"
	ErrInvalid             domainError = "network: invalid desired state"
)

// IsPlanStale reports whether err is ErrPlanStale: the gateway changed since
// the plan was made; plan again.
func IsPlanStale(err error) bool {
	e, ok := err.(domainError)
	return ok && e == ErrPlanStale
}

// IsConflicts reports whether err is ErrConflicts: the plan has conflicts.
func IsConflicts(err error) bool {
	e, ok := err.(domainError)
	return ok && e == ErrConflicts
}
```

- **Verificar antes de cambiar** que el texto de cada `fmt.Err("…")` actual es exactamente el string
  literal (un test lo fija para los cinco).
- `ErrInvalid` se usa envuelto (`fmt.Errf("%w: %s", ErrInvalid, …)` en `network.go`); `IsInvalid`
  sigue detectándolo por prefijo, **sin cambios** (no usa reflexión). Confirmar que `fmt.Errf` con
  `%w` sigue produciendo el mismo texto cuando `ErrInvalid` es `domainError`.
- `conformance/conformance.go:113`: `if err != network.ErrPlanStale {` → `if !network.IsPlanStale(err) {`.
  `:305`: `if err != network.ErrConflicts {` → `if !network.IsConflicts(err) {`. Esto además exige a cada
  implementación de `Gateway` devolver el centinela tal cual, sin envolverlo.
- Comentario de `Gateway.Apply` (`plan.go:58-59`): agregar "detect them with IsPlanStale /
  IsConflicts, never ==".

## 4. Tests (rojo primero, en `tests/`)

- `IsPlanStale(network.ErrPlanStale)` → true; `IsPlanStale(network.ErrConflicts)` → false;
  `IsPlanStale(nil)` → false. Lo mismo para `IsConflicts`.
- El `Error()` de los cinco centinelas es igual a su texto anterior.
- `IsInvalid` sigue verde (`tests/validate_test.go`).
- La conformance de `mem` (`tests/conformance_test.go`) sigue verde con `IsPlanStale`/`IsConflicts`.
- `gotest` verde (vet, race, tests, wasm).

## 5. Criterios de aceptación

- `grep -rnE '(==|!=) *(network\.)?Err[A-Za-z]*' --include=*.go .` → vacío (fuera de comparaciones
  `e == ErrX` sobre un `domainError` ya afirmado).
- Exportados nuevos: solo `IsPlanStale` e `IsConflicts`.
- `README.md` / `docs/ARCHITECTURE.md` muestran `IsPlanStale`/`IsConflicts` donde mencionan los
  centinelas.
- `gotest` verde.

## 6. Restricciones

Las de `AGENTS.md`, más: nada de `reflect`, `unsafe`, `errors.Is`/`errors.As`, ni `==`/`!=`/`switch`
entre valores de interfaz con operandos no nil. No tocar otros repos.
