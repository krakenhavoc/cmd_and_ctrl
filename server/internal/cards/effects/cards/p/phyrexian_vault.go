package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Vault — Artifact {3} (Mirage):
//
//	"{2}, {T}, Sacrifice a creature: Draw a card."
//
// Mana, tap and a creature sacrificed as the cost (CR 602.2b); the sacrifice
// outlet is any creature its controller has, itself included if it is one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b628150b-08a1-4ea3-978d-60255dfb0b7e",
		Name:         "Phyrexian Vault",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice a creature: Draw a card.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeACreature()),
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerSacOutlet},
			Effect:  effects.DrawOneCardEffect,
		}},
	})
}
