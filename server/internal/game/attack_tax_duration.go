package game

import (
	"fmt"

	"github.com/google/uuid"
)

// attack_tax_duration.go — #3398: an attack tax made by a resolving
// spell or ability that lasts for a while. Summon: Yojimbo's "Until
// your next turn, creatures can't attack you unless their controller
// pays {2} for each of those creatures."
//
// attack_tax.go's statics are read off the defender's battlefield
// permanents. This is the SECOND source its header reserved: a
// ModAttackTax ScopedEffect in Game.ScopedEffects, timed by an ADR 0063
// Duration, read by the same pricer (priceAttackDeclarationLocked), so
// the declaration verb, the legal-move enumerator and the bot price an
// attack identically.
//
// The record is plain data (ADR 0041), so a table holding one is still
// a restore point, and it adds no snapshot key: Player is the protected
// seat, Text the mana for ONE attacking creature, and the record's Label
// is the clause. It is a new mod kind, so a binary from before it
// refuses a restore point naming it rather than reading it as no tax.
//
// The any-defender form (#3456, War Tax's "this turn, creatures can't
// attack unless their controller pays {X} for each attacking creature
// they control") leaves Player zero, as other scoped records do for
// "every player". It adds no snapshot key and restores as written: the
// mana is the record's Text, already resolved from the activation's X
// ("{2}" for X=2), so each activation's record carries its own price and
// several stack. It covers every attack target kind.
//
// Like AttackTaxOnPlayer, the tax defends the player and not their
// planeswalkers ("can't attack YOU"); the "you" is the record's Player,
// set from the controller of the resolving ability.

// ModAttackTax is "creatures can't attack <Player> unless their
// controller pays <Text> for each of them", for the record's duration.
// Reads Player (a zero Player is the any-defender form, War Tax: it
// taxes an attack on any player or planeswalker, #3456) and Text (a mana cost for one creature,
// required). Read only by the attack-tax pricer.
const ModAttackTax ModKind = "attackTax"

// AttackTaxMod is the ModAttackTax mod for `defender` charging `mana`
// (a cost string such as "{2}") per attacking creature.
func AttackTaxMod(defender uuid.UUID, mana string) Mod {
	return Mod{Kind: ModAttackTax, Player: defender, Text: mana}
}

// attackTaxModProblem is the registration and restore check for
// ModAttackTax. "" when the mod is sound or is another kind.
func attackTaxModProblem(m Mod) string {
	if m.Kind != ModAttackTax {
		return ""
	}
	if m.Text == "" {
		return "an attackTax mod names no mana cost"
	}
	if _, err := ParseCost(m.Text); err != nil {
		return fmt.Sprintf("an attackTax mod's cost %q does not parse", m.Text)
	}
	return ""
}

// AttackTaxForEffect registers "creatures can't attack `defender`
// unless their controller pays `mana` for each of those creatures" until
// `d` runs out. `label` is the printed clause, for the log. Reports
// whether a record was written.
//
// Caller must hold g.mu (write).
func (g *Game) AttackTaxForEffect(sourceID, defender uuid.UUID, mana string, d Duration, label string) bool {
	if g.playerByIDLocked(defender) == nil {
		return false
	}
	if d.IsZero() {
		d = g.UntilEndOfTurnDuration()
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		[]Mod{AttackTaxMod(defender, mana)}, d, label)
}

// AnyAttackTaxForEffect registers "creatures can't attack unless their
// controller pays `mana` for each attacking creature they control" until
// `d` runs out (default: end of turn), whoever is defended. `mana` is a
// resolved cost such as "{2}"; an empty or "{0}" cost writes nothing.
// Reports whether a record was written.
//
// Caller must hold g.mu (write).
func (g *Game) AnyAttackTaxForEffect(sourceID uuid.UUID, mana string, d Duration, label string) bool {
	if mana == "" || mana == "{0}" {
		return false
	}
	if _, err := ParseCost(mana); err != nil {
		return false
	}
	if d.IsZero() {
		d = g.UntilEndOfTurnDuration()
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		[]Mod{AttackTaxMod(uuid.Nil, mana)}, d, label)
}

// durationAttackTaxLines is the taxes the live ModAttackTax records
// charge one attacker against `defender`. The duration is asked here
// rather than left to the sweep, for handPlayBanLocked's reason:
// "until your next turn" ends as that turn begins.
//
// Caller must hold g.mu (read or write).
func (g *Game) durationAttackTaxLines(defender uuid.UUID, kind AttackTargetKind, attacker uuid.UUID) []AttackTaxLine {
	var out []AttackTaxLine
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind != ModAttackTax || m.Text == "" {
				continue
			}
			// A named player is defended as a player only; the zero
			// Player taxes every attack.
			if m.Player != uuid.Nil && (m.Player != defender || !AttackTaxOnPlayer.covers(kind)) {
				continue
			}
			if g.durationExpiredLocked(e.Duration, false) {
				continue
			}
			out = append(out, AttackTaxLine{
				Source:   e.Source.ID,
				Label:    e.Label,
				Defender: defender,
				Attacker: attacker,
				Mana:     m.Text,
			})
		}
	}
	return out
}
