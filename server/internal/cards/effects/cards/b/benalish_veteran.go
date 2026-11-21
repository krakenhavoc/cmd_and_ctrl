package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Benalish Veteran — Creature — Human Soldier {2}{W}, 2/2 (Magic 2012,
// #5573):
//
//	"Whenever this creature attacks, it gets +1/+1 until end of turn."
//
// Soltari Trooper's attack trigger: it gets +1/+1 until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5da2e5c7-ec04-4f1f-bb2f-7572f877393c",
		Name:         "Benalish Veteran",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Benalish Veteran — it gets +1/+1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Benalish Veteran — +1/+1")),
		},
	})
}
