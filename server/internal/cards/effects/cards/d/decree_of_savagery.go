package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Decree of Savagery — Instant {7}{G}{G}:
//
//	"Put four +1/+1 counters on each creature you control.
//	 Cycling {4}{G}{G} ({4}{G}{G}, Discard this card: Draw a card.)
//	 When you cycle this card, you may put four +1/+1 counters on target
//	 creature."
//
// The spell is one placement of four counters per creature (the set is
// read before the first goes on). The cycle trigger fires from the
// graveyard (CR 702.29c) like every cycling trigger, takes a real target,
// and asks its "you may" as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96a6a8d4-6077-470f-984e-b576a24557a2",
		Name:         "Decree of Savagery",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PutPlusOneCountersOnEachYouControlThat(4, effects.Creature())(ctx.Game, item)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{4}{G}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.InGraveyard(effects.On(game.EventCycle, effects.Self,
					"Decree of Savagery — cycled: you may put four +1/+1 counters on target creature",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						ts := ctx.LegalTargets()
						if len(ts) == 0 {
							return nil
						}
						return effects.AddCounter{Target: ts[0].ID, Kind: game.CounterPlusOne, N: 4}.Apply(ctx)
					})),
				effects.TargetCreature("target creature")),
				"Put four +1/+1 counters on target creature?"),
		},
	})
}
