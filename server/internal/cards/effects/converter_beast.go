package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Converter Beast — Creature — Phyrexian Beast {3}{G} (March of the
// Machine):
//
//	"When this creature enters, incubate 5."
//
// A plain enters trigger on the stack (ADR 0018), so opponents may
// respond to it; the Incubator (CR 701.53a, ADR 0185) is made when it
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7fbb9547-a805-4768-8574-e283e64600a4",
		Name:         "Converter Beast",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Converter Beast — incubate 5", Do(Incubate{N: 5})),
		},
	})
}
