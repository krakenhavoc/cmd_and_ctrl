package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ultima — {3}{W}{W} Sorcery:
//
//	"Destroy all artifacts and creatures. End the turn. (Exile all
//	 spells and abilities from the stack, including this card. The
//	 player whose turn it is discards down to their maximum hand size.
//	 Damage wears off, and "this turn" and "until end of turn" effects
//	 end.)"
//
// DestroyAllMatching over artifacts and creatures, then EndTheTurn
// (end_the_turn.go, CR 724.1, #2165), in the printed order.
//
// The order is what makes the card. The wipe's dies and "leaves the
// battlefield" triggers fire during Ultima's resolution, before the
// turn ends, so they are triggered abilities that have not been put on
// the stack when the process begins — and CR 724.1a says those cease
// to exist. Blood Artist does not drain for the board Ultima kills. A
// commander caught in the wipe is still offered the command zone: that
// is a state-based action question (CR 903.9a), not a trigger.
//
// Indestructible permanents survive and regeneration shields apply,
// as for any "destroy" (CR 702.12b, CR 701.19c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c9a45c28-826e-4e6e-8adc-1368ea1af11a",
		Name:         "Ultima",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (DestroyAllMatching{Match: Or(Artifact(), Creature())}).Apply(ctx); err != nil {
				return err
			}
			return EndTheTurn{}.Apply(ctx)
		},
	})
}
