package cards

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
)

func writeRoles(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "card-roles.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestIndexRolesLoadsTheFileBesideTheDump(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	p := writeRoles(t, `{"fetched_at":"2026-11-04T12:00:00Z","source":"oracle_tags x","tags":{"ramp":2},
		"roles":{"`+a.String()+`":["ramp","mana_rock"],"`+b.String()+`":["tutor"]}}`)
	idx := NewIndex()
	if _, ok := idx.RolesLoaded(); ok {
		t.Fatal("a fresh index must report roles unavailable")
	}
	n, err := idx.LoadRoles(p)
	if err != nil || n != 2 {
		t.Fatalf("LoadRoles = %d, %v", n, err)
	}
	if got := idx.Roles(a); !slices.Equal(got, []string{RoleRamp, RoleManaRock}) {
		t.Errorf("Roles = %v", got)
	}
	if got := idx.Roles(uuid.New()); got != nil {
		t.Errorf("an untagged card has roles %v", got)
	}
	if at, ok := idx.RolesLoaded(); !ok || !at.Equal(time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("RolesLoaded = %v, %v", at, ok)
	}
	// The slice is the caller's: changing it leaves the index alone.
	got := idx.Roles(a)
	got[0] = "x"
	if idx.Roles(a)[0] != RoleRamp {
		t.Error("Roles returned the index's own slice")
	}
}

func TestIndexRolesMissingOrBrokenFile(t *testing.T) {
	idx := NewIndex()
	if _, err := idx.LoadRoles(filepath.Join(t.TempDir(), "none.json")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a missing file: err = %v, want os.ErrNotExist", err)
	}
	if got := idx.Roles(uuid.New()); got != nil {
		t.Errorf("no file, yet roles %v", got)
	}
	// A broken file leaves the roles already loaded in place.
	id := uuid.New()
	good := writeRoles(t, `{"roles":{"`+id.String()+`":["fog"]}}`)
	if _, err := idx.LoadRoles(good); err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{`{`, `{"roles":{"not-a-uuid":["fog"]}}`} {
		if _, err := idx.LoadRoles(writeRoles(t, body)); err == nil {
			t.Errorf("LoadRoles(%s) = nil error", body)
		}
	}
	if got := idx.Roles(id); !slices.Equal(got, []string{RoleFog}) {
		t.Errorf("a broken file changed the roles: %v", got)
	}
}

func TestApplyRoleOverrides(t *testing.T) {
	id, other := uuid.New(), uuid.New()
	prev := roleOverrides
	t.Cleanup(func() { roleOverrides = prev })
	roleOverrides = map[uuid.UUID]roleEdit{
		id: {Add: []string{RoleWincon}, Remove: []string{RoleBurn}},
	}
	in := []string{RoleBurn, RoleRamp, RoleDraw}
	if got := ApplyRoleOverrides(id, in); !slices.Equal(got, []string{RoleRamp, RoleDraw, RoleWincon}) {
		t.Errorf("override = %v", got)
	}
	if !slices.Equal(in, []string{RoleBurn, RoleRamp, RoleDraw}) {
		t.Errorf("the input was modified: %v", in)
	}
	// No override: the roles come back in AllRoles order, once each.
	if got := ApplyRoleOverrides(other, []string{RoleDraw, RoleRemoval, RoleDraw}); !slices.Equal(got, []string{RoleRemoval, RoleDraw}) {
		t.Errorf("no override = %v", got)
	}
	if got := ApplyRoleOverrides(other, nil); len(got) != 0 {
		t.Errorf("nil in, %v out", got)
	}
}

func TestEveryOverrideNamesARealRole(t *testing.T) {
	for id, e := range roleOverrides {
		for _, r := range append(slices.Clone(e.Add), e.Remove...) {
			if !slices.Contains(AllRoles, r) {
				t.Errorf("override %s names %q, which is not a role", id, r)
			}
		}
	}
}

func TestAllRolesAreDistinct(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range AllRoles {
		if seen[r] {
			t.Errorf("role %q listed twice", r)
		}
		seen[r] = true
	}
	if len(AllRoles) != 17 {
		t.Errorf("%d roles; ADR 0217 §7 lists 17", len(AllRoles))
	}
}
