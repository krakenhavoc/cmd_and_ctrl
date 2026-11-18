package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gnoll Hunter — Creature — Gnoll {1}{G}, 2/2 (Adventures in the Forgotten
// Realms):
//
//	"Pack tactics — Whenever this creature attacks, if you attacked with
//	 creatures with total power 6 or greater this combat, put a +1/+1
//	 counter on this creature."
//
// Pack tactics is s82AfrB09PackTactics; the counter goes on the Hunter if it is
// still the same object on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e445336-8b1e-4642-b8bc-aa460ad79a3f",
		Name:         "Gnoll Hunter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			s82AfrB09PackTactics("Gnoll Hunter — pack tactics: a +1/+1 counter on this creature", thisGetsAPlusOneCounter),
		},
	})
}
