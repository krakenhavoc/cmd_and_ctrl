package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demon's Horn — Artifact {2} (Darksteel, #5605):
//
//	"Whenever a player casts a black spell, you may gain 1 life."
//
// Angel's Feather's trigger with black: any player's cast counts, its
// controller's included (tmcB8APlayerCast). The colour is a word of the
// Horn's text, read through it (CR 612.2). The trigger resolves even if
// the spell is countered in response. The life gain is a "you may"
// (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee38377f-8d3a-402e-a5d7-a01cfe3a5322",
		Name:         "Demon's Horn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, tmcB8APlayerCast(OfColor("B")),
				"Demon's Horn — you may gain 1 life", Do(GainLife{Amount: 1})),
				"Demon's Horn — gain 1 life?"),
		},
	})
}
