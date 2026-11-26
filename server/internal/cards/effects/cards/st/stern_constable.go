package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stern Constable — Creature — Human Soldier {W}, 1/1 (Shadows over
// Innistrad, S82 OLD):
//
//	"{T}, Discard a card: Tap target creature."
//
// The discard is part of the cost, paid with the {T} (CR 602.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb3f493a-a5d6-4b9b-9ea0-3c295a67f730",
		Name:         "Stern Constable",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Discard a card: Tap target creature.",
			Cost:    effects.Plus(effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
