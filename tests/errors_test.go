package tests

import (
	"testing"

	"webtyp.com/network"
)

func TestIsPlanStale(t *testing.T) {
	if !network.IsPlanStale(network.ErrPlanStale) {
		t.Errorf("IsPlanStale(ErrPlanStale) = false, want true")
	}
	if network.IsPlanStale(network.ErrConflicts) {
		t.Errorf("IsPlanStale(ErrConflicts) = true, want false")
	}
	if network.IsPlanStale(nil) {
		t.Errorf("IsPlanStale(nil) = true, want false")
	}
}

func TestIsConflicts(t *testing.T) {
	if !network.IsConflicts(network.ErrConflicts) {
		t.Errorf("IsConflicts(ErrConflicts) = false, want true")
	}
	if network.IsConflicts(network.ErrPlanStale) {
		t.Errorf("IsConflicts(ErrPlanStale) = true, want false")
	}
	if network.IsConflicts(nil) {
		t.Errorf("IsConflicts(nil) = true, want false")
	}
}

func TestErrorStrings(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{network.ErrPlanStale, "network: the gateway changed since the plan was made, plan again"},
		{network.ErrConflicts, "network: the plan has conflicts, resolve them before applying"},
		{network.ErrUnknownAccess, "network: unknown access level"},
		{network.ErrUnknownUnregistered, "network: unknown policy for unregistered devices"},
		{network.ErrInvalid, "network: invalid desired state"},
	}

	for _, tc := range cases {
		if got := tc.err.Error(); got != tc.want {
			t.Errorf("Error() = %q, want %q", got, tc.want)
		}
	}
}
