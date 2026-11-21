package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charging Bandits — Creature — Human Rogue {4}{B}, 3/3 (Portal, #5630):
//
//	"Whenever this creature attacks, it gets +2/+0 until end of turn."
//
// Vicious Kavu's attack trigger, word for word.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c67133b-be15-47d7-8cf6-79987a42044d",
		Name:         "Charging Bandits",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Charging Bandits — it gets +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Charging Bandits — +2/+0")),
		},
	})
}
