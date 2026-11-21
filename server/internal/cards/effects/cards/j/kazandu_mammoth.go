package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kazandu Mammoth // Kazandu Valley — Modal DFC. Front: Creature —
// Elephant {1}{G}{G}, 3/3: "Landfall — Whenever a land you control
// enters, this creature gets +2/+2 until end of turn." Back: Land
// (mdfc_lands.go).
//
// Each land is its own trigger, so two lands in a turn make +4/+4. The
// boost is aimed at the Mammoth itself and does nothing once it has left
// the battlefield. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ac1c95c-2a9d-40bc-9cad-9cadfa3f19f7",
		Name:         "Kazandu Mammoth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Kazandu Mammoth — +2/+2 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Kazandu Mammoth — +2/+2 until end of turn")),
		},
	})
}
