package effects

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

const (
	kayasWrathID   = "bc8de6c7-c69d-4add-8f25-825d945874f9" // Purpose{Sweep}
	proctorsGazeID = "a7fe7621-9bb1-4d0d-8b50-c179c685d79d" // Purpose{Lands: 1}
	hedronArchive  = "32263baa-d3f0-463f-92b3-4e9938476add" // an ability Purpose{Draws: 2}
)

func rolesIndex(t *testing.T, body string) *cards.Index {
	t.Helper()
	p := filepath.Join(t.TempDir(), "card-roles.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	idx := cards.NewIndex()
	if _, err := idx.LoadRoles(p); err != nil {
		t.Fatal(err)
	}
	return idx
}

func TestCardRolesAddsWhatTheCatalogPurposeDeclares(t *testing.T) {
	// The tags say nothing about any of these cards: a declared purpose
	// adds its role even when the tags lack it (ADR 0217 §7).
	idx := rolesIndex(t, `{"roles":{}}`)
	for _, c := range []struct {
		id   string
		want string
	}{
		{kayasWrathID, cards.RoleBoardWipe},
		{proctorsGazeID, cards.RoleRamp},
		{hedronArchive, cards.RoleDraw},
	} {
		got := CardRoles(idx, uuid.MustParse(c.id))
		if !slices.Contains(got, c.want) {
			t.Errorf("%s: roles %v lack %s", c.id, got, c.want)
		}
	}
}

func TestCardRolesKeepsTheTagsAndNeverRepeatsARole(t *testing.T) {
	id := uuid.MustParse(kayasWrathID)
	idx := rolesIndex(t, `{"roles":{"`+id.String()+`":["removal","board_wipe"]}}`)
	got := CardRoles(idx, id)
	if !slices.Equal(got, []string{cards.RoleRemoval, cards.RoleBoardWipe}) {
		t.Errorf("roles = %v, want the tag roles once each", got)
	}
}

func TestCardRolesWithoutARolesFile(t *testing.T) {
	// A missing file means no tag roles; the catalog still speaks.
	for _, idx := range []*cards.Index{nil, cards.NewIndex()} {
		if got := CardRoles(idx, uuid.MustParse(kayasWrathID)); !slices.Equal(got, []string{cards.RoleBoardWipe}) {
			t.Errorf("roles = %v, want board_wipe from Purpose alone", got)
		}
		if got := CardRoles(idx, uuid.New()); len(got) != 0 {
			t.Errorf("a card outside the catalog and the tags has roles %v", got)
		}
	}
}

func TestPurposeRolesAreAlwaysRoles(t *testing.T) {
	for _, s := range registry {
		for _, r := range purposeRoles(s) {
			if !slices.Contains(cards.AllRoles, r) {
				t.Fatalf("%s: purpose role %q is not in cards.AllRoles", s.Name, r)
			}
		}
	}
}
