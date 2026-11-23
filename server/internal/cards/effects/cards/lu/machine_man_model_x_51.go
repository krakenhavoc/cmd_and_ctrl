package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "853587be-9709-43a4-81c9-5b6dabe523cf",
		Name:         "Machine Man, Model X-51",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), label, func(g *game.Game, item *game.StackItem) error {
				if !effects.B09SourceStillOnBattlefield(g, item) {
					return nil
				}
				if err := effects.PutACounterOnThis(game.CounterPlusOne)(g, item); err != nil {
					return err
				}
				return effects.ThisGainsKeywordUntilEndOfTurn("flying", label)(g, item)
			}),
		},
	})
}
