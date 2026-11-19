package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Golem's Heart — Artifact {2} (Scars of Mirrodin, #5577):
//
//	"Whenever a player casts an artifact spell, you may gain 1 life."
//
// Angel's Feather's trigger on an artifact spell cast by any player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "650890bc-aba6-45fb-8096-3a36b0a015b3",
		Name:         "Golem's Heart",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, tmcB8APlayerCast(Artifact()),
				"Golem's Heart — you may gain 1 life", Do(GainLife{Amount: 1})),
				"Golem's Heart — gain 1 life?"),
		},
	})
}
