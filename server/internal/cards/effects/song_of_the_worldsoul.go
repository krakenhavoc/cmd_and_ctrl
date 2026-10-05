package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Song of the Worldsoul — Enchantment {4}{W}{W}:
//
//	"Whenever you cast a spell, populate. (Create a token that's a
//	 copy of a creature token you control.)"
//
// A cast trigger that goes on the stack above the spell and populates
// as it resolves; the spell still resolves after it. Any spell counts,
// creature or not. Populate is the shared primitive in populate.go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fbb70f54-8c33-4de0-b05c-f0bff9ffa6ce",
		Name:         "Song of the Worldsoul",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCast(nil), "Song of the Worldsoul — populate", Do(Populate{})),
		},
	})
}
