package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warden of the Grove — Creature — Hydra {2}{G}, 2/2 (Tarkir:
// Dragonstorm, #4463):
//
//	"At the beginning of your end step, put a +1/+1 counter on this
//	 creature.
//	 Whenever another nontoken creature you control enters, it endures X,
//	 where X is the number of counters on this creature."
//
// Two abilities. The end step one is Do(CounterOnThis). The endure one
// is endure_n.go over the creature that entered (the trigger's own
// event): X counts counters of every kind on the Warden as the ability
// resolves, and from its last-known state once it has left (the ruling,
// CR 608.2h). The entering creature's controller chooses counters or the
// Spirit; a creature that has left makes only the Spirit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13bf8103-3db3-4ee6-a6bb-441aabfd9e41",
		Name:         "Warden of the Grove",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Warden of the Grove — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B15AnotherNontokenCreatureYouControlEntered(ev, source, g)
			}, "Warden of the Grove — it endures X",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.EndureN(ctx, ctx.Trigger().Event.CardID, effects.StdB9CountersOnSource(ctx), "Warden of the Grove")
				}),
		},
	})
}
