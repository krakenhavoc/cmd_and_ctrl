package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel's Feather — Artifact {2} (Darksteel, #5605):
//
//	"Whenever a player casts a white spell, you may gain 1 life."
//
// Any player's cast counts, its controller's included (tmcB8APlayerCast).
// The colour is a word of the Feather's text, read through it, so a
// Sleight of Mind that changed "white" changes what it watches (CR 612.2).
// The trigger resolves even if the spell is countered in response. The
// life gain is a "you may" (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d0d4ba4-d5aa-4d9f-89ab-46687637fecf",
		Name:         "Angel's Feather",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, tmcB8APlayerCast(OfColor("W")),
				"Angel's Feather — you may gain 1 life", Do(GainLife{Amount: 1})),
				"Angel's Feather — gain 1 life?"),
		},
	})
}
