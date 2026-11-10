package game

import (
	"fmt"

	"github.com/google/uuid"
)

// add_damage.go — "if a source you control would deal noncombat damage
// to a permanent or player this turn, it deals that much damage plus X
// instead" (Taii Wakeen, Perfect Shot; #4468), created by a resolving
// ability. The additive sibling of ModMultiplyDamage (ADR 0108 §3,
// multiply_damage.go): a CR 614.1a replacement ("instead") that lasts
// until the turn ends (CR 514.2), with the multiplier's closed
// vocabulary for "which sources" and "to what", read the same way and
// as the damage would be dealt (CR 611.2c). A static "plus N" printed on
// a permanent (Torbran, Artist's Talent) is a ReplacementEffect on the
// card and lives as long as the permanent does; this record outlives
// its source, which is why it is a ScopedEffect.
//
// Its own kind rather than a flag on the multiplier, so an older binary
// refuses a table holding one instead of doubling the damage, and so the
// two do not commute: beside a multiplier the affected player orders
// them (CR 616.1), "+X then ×2" or "×2 then +X". Two additions commute
// (+X+Y is +Y+X), so no order is asked between them.

// ModAddDamage is "if <source> would deal [noncombat] damage [to
// <recipient>] this turn, it deals that much damage plus N instead"
// (CR 614.1a). Reads Amount (N, at least 1), Sources, Recipients, Player,
// CombatOnly and NoncombatOnly. Scope ScopeGame.
const ModAddDamage ModKind = "addDamage"

// DamageAddition is the queue-side description of a ModAddDamage
// record, the additive DamageMultiplier.
type DamageAddition struct {
	// EffectSource is the card whose ability made the effect; Controller
	// is its "you".
	EffectSource uuid.UUID
	Controller   uuid.UUID

	// Amount is the "plus N".
	Amount int

	// Sources is "a source you control" and the rest; Recipients is "to
	// what", with Player for "that player and their permanents".
	Sources    DamageSources
	Recipients DamageRecipients
	Player     uuid.UUID

	// CombatOnly is "would deal combat damage"; NoncombatOnly "would deal
	// noncombat damage". At most one.
	CombatOnly    bool
	NoncombatOnly bool

	// Label is the record's label.
	Label string
}

// AddDamageForEffect registers the addition for the rest of the turn.
// Reports whether a record was written: nothing is registered for an
// amount below 1 ("plus 0" changes nothing), a vocabulary value this
// binary does not know, both CombatOnly and NoncombatOnly, or a
// player-scoped recipient with no player.
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) AddDamageForEffect(d DamageAddition) bool {
	m := Mod{
		Kind:          ModAddDamage,
		Amount:        d.Amount,
		Sources:       d.Sources,
		Recipients:    d.Recipients,
		Player:        d.Player,
		CombatOnly:    d.CombatOnly,
		NoncombatOnly: d.NoncombatOnly,
	}
	if addDamageModProblem(m) != "" {
		return false
	}
	label := d.Label
	if label == "" {
		label = "damage increased"
	}
	return g.RegisterScopedRuleEffectForEffect(d.EffectSource, ScopeGame, d.Controller, []Mod{m}, g.UntilEndOfTurnDuration(), label)
}

// addDamageModProblem is registration's (and restore's) check on the
// kind's parameters. "" when the mod is sound. multiplyDamageModProblem
// sends an addDamage mod here and refuses NoncombatOnly on every other
// kind.
func addDamageModProblem(m Mod) string {
	switch {
	case m.Kind != ModAddDamage:
		return fmt.Sprintf("mod %q is not addDamage", m.Kind)
	case m.Amount < 1:
		return fmt.Sprintf("an addDamage effect needs an amount of at least 1, got %d", m.Amount)
	case !m.Sources.Known():
		return fmt.Sprintf("damage sources %q", m.Sources)
	case !m.Recipients.Known():
		return fmt.Sprintf("damage recipients %q", m.Recipients)
	case m.CombatOnly && m.NoncombatOnly:
		return "an addDamage effect is both combat-only and noncombat-only"
	case len(m.Objects) != 0:
		return "an addDamage effect names a source object, which it does not read"
	case m.Next:
		return "an addDamage effect carries a \"next time\" it does not read"
	case (m.Recipients == DamageRecipientsPlayerAndTheirPermanents) != (m.Player != uuid.Nil):
		return "an addDamage effect names a player exactly when its recipients are that player and their permanents"
	case len(m.Queries) != 0:
		return "an addDamage effect carries a property recheck it does not read"
	case m.Then != "":
		return "an addDamage effect carries a follow-up it does not read"
	}
	return ""
}

// addDamageAppliesLocked is the kind's AppliesTo.
//
// Caller must hold g.mu.
func (g *Game) addDamageAppliesLocked(e ScopedEffect, m Mod, ev *ReplacementEvent) bool {
	// CR 120.8: a source that would deal 0 damage deals none, so there
	// is no damage to add to (the Taii Wakeen ruling: "If all of the
	// damage is prevented, Taii Wakeen's effect no longer applies").
	if ev.Kind != RepEventDamage || ev.DamageAmount <= 0 || m.Amount < 1 {
		return false
	}
	if m.CombatOnly && !ev.IsCombatDamage {
		return false
	}
	if m.NoncombatOnly && ev.IsCombatDamage {
		return false
	}
	if !g.damageSourcesMatchLocked(e, m.Sources, ev) {
		return false
	}
	return g.damageRecipientsMatchLocked(e, m, ev.DamageTarget)
}

// applyAddDamageLocked is the kind's Replace: that much damage plus N.
// The added damage is dealt by the same source (the ruling): only the
// amount changes.
func applyAddDamageLocked(m Mod, ev *ReplacementEvent) {
	ev.DamageAmount += m.Amount
}
