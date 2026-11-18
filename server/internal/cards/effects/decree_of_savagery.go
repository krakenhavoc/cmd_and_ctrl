package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "96a6a8d4-6077-470f-984e-b576a24557a2",
		Name:         "Decree of Savagery",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return putPlusOneCountersOnEachYouControlThat(4, Creature())(ctx.Game, item)
		},
		Activated: []ActivatedAbility{Cycling("{4}{G}{G}")},
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				InGraveyard(On(game.EventCycle, Self,
					"Decree of Savagery — cycled: you may put four +1/+1 counters on target creature",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						ts := ctx.LegalTargets()
						if len(ts) == 0 {
							return nil
						}
						return AddCounter{Target: ts[0].ID, Kind: game.CounterPlusOne, N: 4}.Apply(ctx)
					})),
				TargetCreature("target creature")),
				"Put four +1/+1 counters on target creature?"),
		},
	})
}
