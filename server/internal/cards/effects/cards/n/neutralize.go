package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Neutralize — Instant {1}{U}{U} (Ikoria: Lair of Behemoths, #5512):
//
//	"Counter target spell.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Cancel's body (counterTheTargetSpell), and cycling from hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2c60973-b4d5-43fa-a0c4-b5688b28df05",
		Name:         "Neutralize",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve:    effects.CounterTheTargetSpell,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
