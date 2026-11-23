package game

import (
	"fmt"

	"github.com/google/uuid"
)

// targeting_bar_table.go — "players and permanents can't be the targets
// of spells or activated abilities" for a duration (Peace Talks, #5718).
//
// A TABLE-WIDE bar, and the reason it is its own record: hexproof and
// shroud are abilities of one object or one player and stop triggered
// abilities too (CR 702.11, 702.18a); the Vines of Vastwood record
// (targeting_bar.go) covers one object and is relative to its caster.
// This one covers every player and every permanent, whoever controls
// them, and refuses only a spell or an activated ability (CR 113.3b): a
// triggered ability (CR 113.3c) still targets, and so does a copy of one.
// An Aura spell targets as it is cast (CR 303.4a, 115.1b), so it has
// nothing to enchant; an Aura put onto the battlefield without being cast
// does not target (CR 303.4f) and is not refused.
//
// It changes no characteristic, so it is a rule (CR 611.2c): a permanent
// that enters, and a player, are covered for as long as it lasts. It is
// one ScopedEffect, ScopeGame with a ModCantBeTargetedBySpellsOrActivated
// mod, read live by both halves of the targeting choke point
// (canBeTargetedBy for a permanent, canPlayerBeTargetedByLocked for a
// player), so the picker, the legal-move enumerator, the CR 601.2c
// announce gate and the CR 608.2b resolution re-check give one answer:
// a spell that was aimed at a creature before the record began loses
// that target as it resolves. Only the battlefield is covered: a spell
// on the stack and a card in a graveyard are neither players nor
// permanents. A walk that is not targeting (a cost, a "choose") never
// asks (CR 115.10a).

// PlayersAndPermanentsCantBeTargetedForEffect registers the bar from a
// resolving spell or ability for duration `d`. Reports whether a record
// was written. Caller must hold g.mu (write).
func (g *Game) PlayersAndPermanentsCantBeTargetedForEffect(sourceID, controller uuid.UUID, d Duration, label string) bool {
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, controller,
		[]Mod{{Kind: ModCantBeTargetedBySpellsOrActivated}}, d, label)
}

// tableTargetingBarLocked reports whether a live record bars `src` from
// targeting any player or permanent: some record carries the mod, and
// the source is not a triggered ability. The sweep ends the record with
// its duration, so nothing here reads the turn.
//
// Nil-receiver safe for the game-less CanBeTargetedBy, which bars
// nothing. Caller must hold g.mu (read or write). Reads only.
func (g *Game) tableTargetingBarLocked(src TargetSource) bool {
	if g == nil || src.IsTriggered() || len(g.ScopedEffects) == 0 {
		return false
	}
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if e.Scope == ScopeGame && scopedEffectHasMod(e, ModCantBeTargetedBySpellsOrActivated) {
			return true
		}
	}
	return false
}

// tableTargetingBarModProblem is the restore check for the record: its
// mod is read only on a ScopeGame record and reads nothing. "" when
// sound.
func tableTargetingBarModProblem(scope AffectedScope, m Mod) string {
	if m.Kind != ModCantBeTargetedBySpellsOrActivated {
		return ""
	}
	if scope != ScopeGame {
		return fmt.Sprintf("a %s mod on a %q record", m.Kind, scope)
	}
	if m.Player != uuid.Nil || len(m.Objects) > 0 {
		return fmt.Sprintf("a %s mod that names a player or an object", m.Kind)
	}
	return ""
}
