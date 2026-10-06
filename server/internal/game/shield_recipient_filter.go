package game

import (
	"fmt"

	"github.com/google/uuid"
)

// shield_recipient_filter.go — #2045 (ADR 0108 §7, amendment of
// 2026-10-06): a ModPreventFromSource shield that protects a SET the
// protection fields can't name. "Prevent all damage that would be dealt
// to creatures this turn" (Forfend), "to creatures you control" (Divine
// Light), "all combat damage that would be dealt to players" (Defend the
// Hearth), "to Dogs you control" (Pack Leader), "to artifact creatures"
// (Ethersworn Shieldmage).
//
// THE RULES.
//
//   - CR 615.1: prevention effects "apply continuously as events
//     happen—they aren't locked in ahead of time".
//   - CR 611.2c fixes the affected set of a resolved continuous effect
//     only when it modifies characteristics or changes control. A
//     prevention effect does neither, so it "modifies the rules of the
//     game" and "can affect objects that weren't affected when that
//     continuous effect began". A creature that enters after Divine Light
//     resolved, or that you gain control of, is protected; one you lose
//     control of, or that stops being a creature, is not.
//
// THE SHAPE. One plain-data DamageRecipientFilter in Mod.RecipientFilter
// (at most one entry), on an unpinned record (ScopeGame) that names no
// protected player. It composes with every source field — a chosen
// source, Queries and SourceFilter — so Surge of Salvation's "black
// and/or red sources … to creatures you control" and Chameleon Blur's
// "damage that creatures would deal to players" are one record each.
// It refuses a charge: "the next N damage … to each creature" is CR
// 615.11's one shield per creature, not one shared charge.
//
// WHEN IT IS READ. As the damage would be dealt, every time, by
// nextFromSourceProtectsLocked: a player by identity, a permanent as it
// is on the battlefield now (its effective characteristics and
// controller).

// RecipientControllerRel is the filter's controller test for a protected
// permanent, relative to the shield's controller. A closed vocabulary;
// restore refuses a value this binary does not know.
type RecipientControllerRel string

const (
	// RecipientControllerAny asks nothing — the zero value: "to
	// creatures", "to artifact creatures".
	RecipientControllerAny RecipientControllerRel = ""
	// RecipientControllerYou is "you control": the permanent's
	// controller is the record's controller as the damage would be
	// dealt (Divine Light, Pack Leader).
	RecipientControllerYou RecipientControllerRel = "you"
)

// Known reports whether this binary can interpret r.
func (r RecipientControllerRel) Known() bool {
	switch r {
	case RecipientControllerAny, RecipientControllerYou:
		return true
	}
	return false
}

// DamageRecipientFilter is the set of players and permanents a shield
// protects when the set is not "you", "you and <types> you control", a
// pinned object or everything. A recipient is protected when it is a
// player and Players is set, or a permanent and Permanents is set and
// it passes Match and Controller.
type DamageRecipientFilter struct {
	// Players is "to players": every player. Commencement of
	// Festivities and Defend the Hearth add CombatOnly.
	Players bool `json:"players,omitempty"`

	// Permanents is "to <permanents>": permanents narrowed by Match and
	// Controller.
	Permanents bool `json:"permanents,omitempty"`

	// Match are queries the permanent must match EVERY one of — unlike
	// the any-of lists elsewhere, because "artifact creatures" is a
	// permanent that is both: {Types: artifact} and {Types: creature}.
	// "Creatures" is {Types: creature}; "Dogs" is {Subtypes: Dog}.
	// Empty is every permanent. Read only with Permanents.
	Match []PermanentQuery `json:"match,omitempty"`

	// Controller is "you control". Read only with Permanents.
	Controller RecipientControllerRel `json:"controller,omitempty"`
}

// IsZero reports whether f protects nothing it names (the zero value).
func (f DamageRecipientFilter) IsZero() bool {
	return !f.Players && !f.Permanents && len(f.Match) == 0 && f.Controller == RecipientControllerAny
}

// clone deep-copies f.
func (f DamageRecipientFilter) clone() DamageRecipientFilter {
	f.Match = clonePermanentQueries(f.Match)
	return f
}

// cloneRecipientFilters deep-copies a mod's RecipientFilter slice,
// preserving nil.
func cloneRecipientFilters(in []DamageRecipientFilter) []DamageRecipientFilter {
	if len(in) == 0 {
		return nil
	}
	out := make([]DamageRecipientFilter, len(in))
	for i := range in {
		out[i] = in[i].clone()
	}
	return out
}

// recipientFilterProblem is registration's and restore's check on one
// filter. "" when it is sound.
func recipientFilterProblem(f DamageRecipientFilter) string {
	switch {
	case !f.Players && !f.Permanents:
		return "a preventFromSource recipient filter protects neither players nor permanents"
	case !f.Controller.Known():
		return fmt.Sprintf("a preventFromSource recipient filter names controller %q", f.Controller)
	case !f.Permanents && (len(f.Match) != 0 || f.Controller != RecipientControllerAny):
		return "a preventFromSource recipient filter narrows permanents it does not protect"
	}
	for _, q := range f.Match {
		if q.Equal(PermanentQuery{}) {
			return "a preventFromSource recipient filter carries an empty query"
		}
	}
	return ""
}

// recipientFilterModProblem is the record-level half: the filter beside
// the mod's other fields. "" when it is sound.
func recipientFilterModProblem(m Mod) string {
	if len(m.RecipientFilter) == 0 {
		return ""
	}
	switch {
	case m.Kind != ModPreventFromSource:
		return fmt.Sprintf("a %s mod carries a recipient filter, which only preventFromSource reads", m.Kind)
	case len(m.RecipientFilter) > 1:
		return "a preventFromSource shield carries more than one recipient filter"
	case m.Player != uuid.Nil || len(m.Types) != 0:
		return "a preventFromSource shield names a protected player and a recipient filter"
	case m.AndDealtBy:
		return "a to-and-by preventFromSource shield carries a recipient filter"
	case m.Amount != 0:
		return "a charged preventFromSource shield carries a recipient filter (CR 615.11 makes one shield per recipient)"
	}
	return recipientFilterProblem(m.RecipientFilter[0])
}

// recipientFilterScopeProblem refuses a recipient set on a pinned
// record: a pinned shield protects its pinned objects, and the set would
// go unread. "" when the record is sound.
func recipientFilterScopeProblem(scope AffectedScope, mods []Mod) string {
	if scope == ScopeGame {
		return ""
	}
	for _, m := range mods {
		if len(m.RecipientFilter) != 0 {
			return fmt.Sprintf("a %s record with scope %q carries a recipient filter; only an unpinned shield reads one", m.Kind, scope)
		}
	}
	return ""
}

// recipientFilterMatchesLocked reports whether damage to `target` is
// damage to a member of the filter's set, read now (CR 611.2c, 615.1):
// a player by identity, a permanent as it is on the battlefield, by its
// effective characteristics and its controller. Anything else — a
// target the engine can't see — is not protected, the weaker direction.
//
// Caller must hold g.mu with fresh layers, as every replacement
// applicability check does.
func (g *Game) recipientFilterMatchesLocked(e ScopedEffect, f DamageRecipientFilter, target uuid.UUID) bool {
	if target == uuid.Nil {
		return false
	}
	if g.playerByIDLocked(target) != nil {
		return f.Players
	}
	if !f.Permanents {
		return false
	}
	c := findBattlefieldCard(g, target)
	if c == nil {
		return false
	}
	if f.Controller == RecipientControllerYou && (e.Controller == uuid.Nil || c.Controller != e.Controller) {
		return false
	}
	for _, q := range f.Match {
		if !q.matchesLocked(g, c) {
			return false
		}
	}
	return true
}
