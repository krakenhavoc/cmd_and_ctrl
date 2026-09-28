package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightning Serpent — Creature — Elemental Serpent {X}{R}, 2/1 (#1664):
//
//	"Trample, haste
//	 This creature enters with X +1/+0 counters on it.
//	 At the beginning of the end step, sacrifice this creature."
//
// Waited on #1664: the X +1/+0 counters used to be stored and change
// nothing, so the Serpent swung as a 2/1 whatever X was. Every P/T
// counter kind now counts (game.PTCounterDelta, CR 122.1a). The
// counters are the CR 614.1c entry clause (XCounters, #1002), so they
// ride the entry replacement pipeline like Benevolent Hydra's.
//
// "At the beginning of the end step" is ANY end step, not only yours,
// so a Serpent flashed in or reanimated on another player's turn goes
// at that turn's end — AtEachStep, Underworld Breach's shape.
func init() {
	Register(Spec{
		OracleID:                   "d07b4bb6-0c8d-44ea-a5b3-eeb6e36f3631",
		Name:                       "Lightning Serpent",
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"trample", "haste"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters("+1/+0")},
		Triggered: []game.TriggeredAbility{
			AtEachStep(game.StepEnd, "Lightning Serpent — sacrifice it", func(g *game.Game, item *game.StackItem) error {
				return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
			}),
		},
	})
}
