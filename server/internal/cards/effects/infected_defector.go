package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infected Defector — Creature — Phyrexian Knight {4}{W}, 4/3 (March of
// the Machine):
//
//	"When this creature dies, incubate 3."
//
// A dies trigger on the stack (ADR 0018); the Incubator (CR 701.53a,
// ADR 0185) is made when it resolves and is the controller's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cff8dd65-e531-4496-aa49-e8d4350b9620",
		Name:         "Infected Defector",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Infected Defector — incubate 3", Do(Incubate{N: 3})),
		},
	})
}
