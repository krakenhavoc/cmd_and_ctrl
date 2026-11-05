package cards

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/google/uuid"
)

// roles.go — card roles for the advanced stats (ADR 0217 §7, slice 4).
//
// A role is what a card is for in a Commander deck: removal, ramp, a
// tutor. The roles come from Scryfall's community Oracle tags (Tagger),
// which cmd/cardroles reads from Scryfall's Oracle Tags bulk file once a
// week and writes to data/scryfall/card-roles.json. This file is the
// schema of that file and the Index's reader of it. The catalog Purpose
// and the short override list sit on top of the tag roles; see
// effects.CardRoles, which composes the three.
//
// A card's TYPES (land, creature, ...) are not roles; they come from the
// type line.

// The roles (ADR 0217 §7). A role's string is its on-disk identity in
// card-roles.json: never renamed, never reused.
const (
	RoleRemoval         = "removal" // spot removal
	RoleBoardWipe       = "board_wipe"
	RoleCounterspell    = "counterspell"
	RoleFog             = "fog"
	RoleRamp            = "ramp"
	RoleManaRock        = "mana_rock"
	RoleManaDork        = "mana_dork"
	RoleDraw            = "draw"
	RoleTutor           = "tutor"
	RoleProtection      = "protection"
	RoleRecursion       = "recursion"
	RoleGraveyardHate   = "graveyard_hate"
	RoleBurn            = "burn"
	RoleLifegain        = "lifegain"
	RoleTokenMaker      = "token_maker"
	RoleSacrificeOutlet = "sacrifice_outlet"
	RoleWincon          = "wincon"
)

// AllRoles lists every role, in the order a card's roles are reported.
var AllRoles = []string{
	RoleRemoval, RoleBoardWipe, RoleCounterspell, RoleFog,
	RoleRamp, RoleManaRock, RoleManaDork,
	RoleDraw, RoleTutor, RoleProtection, RoleRecursion, RoleGraveyardHate,
	RoleBurn, RoleLifegain, RoleTokenMaker, RoleSacrificeOutlet, RoleWincon,
}

// RolesFile is the on-disk shape of data/scryfall/card-roles.json. It is
// generated, holds no personal data and is never hand-edited.
type RolesFile struct {
	// FetchedAt is when cmd/cardroles wrote the file.
	FetchedAt time.Time `json:"fetched_at"`
	// Source is the Scryfall bulk file the roles were read from, by its
	// updated_at stamp: "oracle_tags 2026-11-04T10:00:31.789+00:00".
	Source string `json:"source,omitempty"`
	// Tags is how many cards each Tagger tag put into its role, by tag
	// slug, children of the tag included. A tag is only listed when it
	// matched at least one card.
	Tags map[string]int `json:"tags"`
	// Roles is the roles of each card by oracle_id, in AllRoles order.
	Roles map[string][]string `json:"roles"`
}

// LoadRoles reads the roles file at path into the index and returns how
// many cards have a role. A missing file returns os.ErrNotExist
// unchanged: the index then has no tag roles, and the caller treats that
// as "roles unavailable", not as an error (ADR 0217 §7). A malformed
// file leaves the previous roles in place. Like Load, it is called once
// at boot, before readers.
func (i *Index) LoadRoles(path string) (int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var f RolesFile
	if err := json.Unmarshal(b, &f); err != nil {
		return 0, fmt.Errorf("parse %s: %w", path, err)
	}
	roles := make(map[uuid.UUID][]string, len(f.Roles))
	for k, v := range f.Roles {
		id, err := uuid.Parse(k)
		if err != nil {
			return 0, fmt.Errorf("parse %s: oracle id %q: %w", path, k, err)
		}
		roles[id] = slices.Clone(v)
	}
	i.mu.Lock()
	i.roles = roles
	i.rolesAt = f.FetchedAt
	i.mu.Unlock()
	return len(roles), nil
}

// Roles is the Tagger roles of the card with this oracle_id, in AllRoles
// order, or nil when it has none or no roles file was loaded. It does
// not include the catalog Purpose or the override list; effects.CardRoles
// does. The slice is the caller's to keep.
func (i *Index) Roles(oracleID uuid.UUID) []string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return slices.Clone(i.roles[oracleID])
}

// RolesLoaded reports whether a roles file is loaded, and when it was
// fetched. The stats page says "roles unavailable" when it is not.
func (i *Index) RolesLoaded() (time.Time, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.rolesAt, i.roles != nil
}
