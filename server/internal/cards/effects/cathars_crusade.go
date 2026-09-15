package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cathars' Crusade — Enchantment {3}{W}{W} (EDHREC rank 553):
//
//	"Whenever a creature you control enters, put a +1/+1 counter on
//	 each creature you control."
//
// The token deck's exponential anthem: N creatures entering is N
// triggers, each growing the whole board. Every entering creature —
// token or card, cast or reanimated — fires it, and the set of
// creatures receiving counters is read when each trigger RESOLVES,
// so the creature that caused the trigger gets its own counter, as
// printed. Counters go through AddCounter, so Doubling Season and
// Hardened Scales apply to each one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cc65ac73-5bef-4ecb-ad8e-39199084c027",
		Name:         "Cathars' Crusade",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, CreatureEnteredUnderYourControl, "Cathars' Crusade — a +1/+1 counter on each creature you control", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, id := range b04CreatureIDsControlledBy(g, item.Controller) {
					if err := (AddCounter{Target: id, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}
