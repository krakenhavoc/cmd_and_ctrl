package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Muddle the Mixture — Instant {U}{U}:
//
//	"Counter target instant or sorcery spell.
//	 Transmute {1}{U}{U} ({1}{U}{U}, Discard this card: Search your
//	 library for a card with the same mana value as this card, reveal
//	 it, put it into your hand, then shuffle. Transmute only as a
//	 sorcery.)"
//
// Transmute is the shared keyword constructor (transmute.go): an
// activated ability from the hand, paid by discarding the card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0826cf7d-7ccc-459b-a92f-29dc169628f8",
		Name:         "Muddle the Mixture",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetInstantOrSorcerySpell("target instant or sorcery spell"),
		Activated:    []effects.ActivatedAbility{effects.Transmute("{1}{U}{U}")},
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
