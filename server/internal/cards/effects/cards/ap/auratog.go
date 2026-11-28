package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Auratog — Creature — Atog {1}{W}, 1/2 (Tempest, S82 OLD b756, #5628):
//
//	"Sacrifice an enchantment: This creature gets +2/+2 until end of turn."
//
// Teferi's Care's cost (SacrificeN of an enchantment, any you control) on a
// self pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe47535e-e50c-49ad-beda-46750d33e6f4",
		Name:         "Auratog",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice an enchantment: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:    effects.SacrificeN(1, "an enchantment", effects.Enchantment()),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Auratog — +2/+2 until end of turn"),
		}},
	})
}
