package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Novice Occultist — Creature — Human Wizard {1}{B}, 1/2 (Innistrad: Midnight Hunt, S82):
//
//	"When this creature dies, you draw a card and you lose 1 life."
//
// Barbed Servitor's draw-and-lose body (b36DrawAndLoseOne), on a dies trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e51f5b41-5465-4b1d-b96a-544762657ca5",
		Name:         "Novice Occultist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Novice Occultist — you draw a card and you lose 1 life", b36DrawAndLoseOne),
		},
	})
}
