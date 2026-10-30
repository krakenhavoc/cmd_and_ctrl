package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Machine Man, Model X-51 — Legendary Artifact Creature — Robot Hero {4},
// 2/3 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever you cast a noncreature spell, put a +1/+1 counter on Machine
//	 Man and he gains flying until end of turn."
//
// The trigger is Artist's Talent's (YouCast(Noncreature())). Both halves
// land on Machine Man himself, and only while he is still the object that
// triggered (CR 400.7).
//
// No simplification.
func init() {
	const label = "Machine Man, Model X-51 — a +1/+1 counter and flying until end of turn"
	Register(Spec{
		OracleID:     "853587be-9709-43a4-81c9-5b6dabe523cf",
		Name:         "Machine Man, Model X-51",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), label, func(g *game.Game, item *game.StackItem) error {
				if !b09SourceStillOnBattlefield(g, item) {
					return nil
				}
				if err := putACounterOnThis(game.CounterPlusOne)(g, item); err != nil {
					return err
				}
				return thisGainsKeywordUntilEndOfTurn("flying", label)(g, item)
			}),
		},
	})
}
