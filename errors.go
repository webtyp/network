package network

import "webtyp.com/fmt"

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

// IsInvalid reports whether err came from Desired.Validate.
func IsInvalid(err error) bool {
	if err == nil {
		return false
	}
	return fmt.HasPrefix(err.Error(), ErrInvalid.Error())
}

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
