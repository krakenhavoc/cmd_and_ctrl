package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whiptongue Hydra — Creature — Lizard Hydra {5}{G}, 4/4:
//
//	"Reach
//	 When this creature enters, destroy all creatures with flying. Put a
//	 +1/+1 counter on this creature for each creature destroyed this way."
//
// Bane of Progress's shape: the sweep hands its tail the creatures it
// actually destroyed, so an indestructible creature, or one that
// regenerates or is saved by a replacement, is not counted (the ruling).
// The flying test reads the layered view, and the Hydra itself (reach, no
// flying) is not swept. If the Hydra has left, the counters have nowhere
// to go (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "53d5d961-d6a3-49b1-b335-2c2b49972008",
		Name:            "Whiptongue Hydra",
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Whiptongue Hydra — destroy all creatures with flying; put a +1/+1 counter on this creature for each creature destroyed this way",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return DestroyAllMatching{
						Match:    And(Creature(), HasKeyword("flying")),
						ThenTail: ctx.ItemTail(whiptongueHydraCounters),
					}.Apply(ctx)
				}),
		},
	})
}

var whiptongueHydraCounters = MassTail("mass/whiptongue-hydra-counters", func(ctx *Context, _ []game.Card, destroyed int) error {
	if destroyed <= 0 {
		return nil
	}
	return AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: destroyed}.Apply(ctx)
})
