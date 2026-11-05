package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// roles.go — a card's roles for the advanced stats (ADR 0217 §7, slice
// 4): Scryfall's Tagger roles, then what the catalog Purpose declares,
// then the override list in cards/roles_overrides.go.

// CardRoles is the roles of the card with this oracle_id, in
// cards.AllRoles order: the Tagger roles the index loaded, plus the roles
// the catalog Purpose adds (a declared purpose adds its role even when
// the tags lack it), with the override list applied last. A nil index is
// the same as one with no roles file. It never returns a role twice.
//
// Roles are read, never stored (ADR 0217 §7 "Drift"): a refresh changes
// every deck's composition at once and nothing is backfilled.
func CardRoles(idx *cards.Index, oracleID uuid.UUID) []string {
	var roles []string
	if idx != nil {
		roles = idx.Roles(oracleID)
	}
	if spec, ok := Lookup(oracleID.String()); ok {
		roles = append(roles, purposeRoles(spec)...)
	}
	return cards.ApplyRoleOverrides(oracleID, roles)
}

// purposeRoles reads every Purpose a catalog entry declares: the card's,
// its alternative costs' (an overload's sweep), its modes', and its
// activated and triggered abilities'.
func purposeRoles(spec Spec) []string {
	var roles []string
	add := func(p game.Purpose) { roles = append(roles, rolesOfPurpose(p)...) }
	add(spec.Purpose)
	for _, a := range spec.AlternativeCosts {
		add(a.Purpose)
	}
	modes := func(m *game.ModeSpec) {
		if m == nil {
			return
		}
		for _, o := range m.Options {
			add(o.Purpose)
		}
	}
	modes(spec.Modes)
	for _, a := range spec.Activated {
		add(a.Purpose)
		modes(a.Modes)
	}
	for _, t := range spec.Triggered {
		add(t.Purpose)
	}
	return roles
}

// rolesOfPurpose maps one Purpose to roles: Sweep to board_wipe, Tutors
// to tutor, Lands to ramp, Draws to draw (ADR 0217 §7), and Tokens to
// token_maker, because Tagger's token tag covers only repeatable makers.
// A sweep that only happens kicked or face up still counts
// (DeclaresSweep); the other amounts are read as declared.
func rolesOfPurpose(p game.Purpose) []string {
	var roles []string
	if p.DeclaresSweep() {
		roles = append(roles, cards.RoleBoardWipe)
	}
	if p.Tutors > 0 {
		roles = append(roles, cards.RoleTutor)
	}
	if p.Lands > 0 {
		roles = append(roles, cards.RoleRamp)
	}
	if p.Draws > 0 {
		roles = append(roles, cards.RoleDraw)
	}
	if p.Tokens > 0 {
		roles = append(roles, cards.RoleTokenMaker)
	}
	return roles
}
