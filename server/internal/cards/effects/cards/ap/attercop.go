package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Attercop — Creature — Spider {1}{G} (The Hobbit), 2/1:
//
//	"Reach, deathtouch
//	 Landfall — Whenever a land you control enters, this creature gets
//	 +1/+1 until end of turn."
//
// #4454, S76. Kazandu Mammoth's landfall pump, one land one trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "45d652c4-188e-4f00-acdd-1be014922cbc",
		Name:            "Attercop",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Attercop — +1/+1 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Attercop — +1/+1 until end of turn")),
		},
	})
}
