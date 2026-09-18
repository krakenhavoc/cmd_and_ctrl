package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Resourceful Defense — Enchantment {2}{W} (EDHREC rank 1602):
//
//	"Whenever a permanent you control leaves the battlefield, if it
//	 had counters on it, put those counters on target permanent you
//	 control.
//	 {4}{W}: Move any number of counters from target permanent you
//	 control onto a second target permanent you control."
//
// The counters-never-die enchantment. Both abilities are live.
//
// The trigger watches every leave — dies, exiled, bounced, tucked —
// and "if it had counters on it" is an intervening-if (CR 603.4)
// read from last-known information: the card's Counters are cleared
// by the move, so the kinds and counts it carried are read back off
// the event log (b14LastKnownCounterKinds) at trigger time, and the
// same snapshot is what the resolution puts on the target. Every
// kind moves — +1/+1, loyalty, charge, lore — as printed. The target
// is chosen as the trigger goes on the stack and re-checked at
// resolution; with no other permanent to receive them, the trigger
// is dropped (CR 603.3d).
//
// The activated ability's two target slots are two CLAUSES (#764):
// "target permanent you control" and "a SECOND target permanent you
// control", the second marked Distinct, so the engine refuses naming
// one permanent twice at announce (CR 601.2c) instead of resolving to
// a no-op, and re-checks each slot against its own clause at
// resolution (CR 608.2b).
//
// Sandbox simplifications, declared, both weaker than printed:
//
//   - The activated ability moves ALL the counters from the first
//     target to the second. "Any number" needs a count prompt for an
//     ability at resolution, which does not exist; moving everything
//     is a legal answer to the printed choice, and never more than
//     the player could have chosen.
//   - A token that leaves the battlefield ceases to exist, so the
//     trigger cannot read whose it was and does not fire for it.
//     A card that left is read where it sits now.
func init() {
	Register(Spec{
		OracleID:     "83e78565-e61f-4bbc-b834-f47941f7e3ec",
		Name:         "Resourceful Defense",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The activated ability moves every counter from the first target to the second — you can't choose how many.",
			"The trigger doesn't fire for a token that leaves the battlefield.",
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b14PermanentYouControlLeft(ev, source, g)
				return ok
			},
			Targets: TargetPermanent("target permanent you control", YouControl()),
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				counters := b14LastKnownCounterKinds(g, ev.CardID)
				return game.NewTriggeredItem(source, "Resourceful Defense — put the counters it had on target permanent you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if err := b14PutCounters(ctx, t.ID, counters); err != nil {
								return err
							}
						}
						return nil
					})
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{4}{W}: Move any number of counters from target permanent you control onto a second target permanent you control.",
			Cost:  ManaCost("{4}{W}"),
			Targets: Clauses(
				TargetPermanent("target permanent you control", YouControl()),
				Distinct(TargetPermanent("a second target permanent you control", YouControl())),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				from, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				to, ok := ctx.ClauseTarget(1)
				if !ok {
					return nil
				}
				return b14MoveAllCounters(ctx, from.ID, to.ID)
			},
		}},
	})
}
