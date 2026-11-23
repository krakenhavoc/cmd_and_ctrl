package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fend Off — Instant {1}{W}:
//
//	"Prevent all combat damage that would be dealt by target creature this turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// ADR 0108 §7 (#1904, Delivery PR 7b): a preventFromSource record whose
// source is the targeted creature, pinned as the spell resolves (CR
// 400.7), combat damage only, to anything, for the rest of the turn.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa8f3827-8cd9-4896-ab0c-26fecacceb40",
		Name:         "Fend Off",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.ShieldAgainstTargetsSpell(true),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
