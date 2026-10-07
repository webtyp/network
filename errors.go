package network

import "webtyp.com/fmt"

var (
	ErrPlanStale           = fmt.Err("network: the gateway changed since the plan was made, plan again")
	ErrConflicts           = fmt.Err("network: the plan has conflicts, resolve them before applying")
	ErrUnknownAccess       = fmt.Err("network: unknown access level")
	ErrUnknownUnregistered = fmt.Err("network: unknown policy for unregistered devices")
	ErrInvalid             = fmt.Err("network: invalid desired state")
)

// IsInvalid reports whether err came from Desired.Validate.
func IsInvalid(err error) bool {
	if err == nil {
		return false
	}
	return fmt.HasPrefix(err.Error(), ErrInvalid.Error())
}
