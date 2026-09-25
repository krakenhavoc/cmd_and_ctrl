package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blowfly Infestation — Enchantment {2}{B} (EDHREC rank 3088):
//
//	"Whenever a creature dies, if it had a -1/-1 counter on it, put
//	 a -1/-1 counter on target creature."
//
// The -1/-1 counter chain. Any creature's death, anyone's; the
// intervening "if" reads the counters the creature had when it
// left, back off the event log (b29CreatureWithMinusCounterDied —
// MoveCard clears them on the way out), and holds at resolution
// for the same reason it held at trigger time. The target is the
// trigger's own clause: any creature, picked when the trigger goes
// on the stack, and a table with no creature left has no trigger at
// all (CR 603.3d). A creature that dies to the counter this places
// starts the chain again, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3859873f-2a0f-463b-8ed9-7f2ab5ed393a",
		Name:         "Blowfly Infestation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b29CreatureWithMinusCounterDied(ev, g)
			},
			Targets: TargetCreature("target creature"),
			Key:     "Blowfly Infestation — put a -1/-1 counter on target creature",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return AddCounter{Target: t.ID, Kind: "-1/-1", N: 1}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
