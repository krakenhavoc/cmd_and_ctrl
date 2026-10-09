package effects

import (
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Graft Surgeon — Creature — Human Cleric {2}{W}, 2/2:
//
//	"This creature enters with a +1/+1 counter on it.
//	 When this creature dies, put its counters on up to one target
//	 creature you control."
//
// District Mascot's enters-with-counter replacement, and Arcbound
// Wanderer's look-back at the dying permanent's counters (CR 603.10a:
// the counters are read from last-known information, since the card is
// already in the graveyard). "Its counters" is every kind it had, not
// only +1/+1: each kind moves in full, in a fixed order so a log reads
// the same twice. The creature it could target can't be itself (it has
// died), and "up to one" lets the controller pick nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe7938ce-6289-4bf8-a8f5-85ecccbb7f86",
		Name:         "Graft Surgeon",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 1, "Graft Surgeon: enters with a +1/+1 counter"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Graft Surgeon — put its counters on up to one target creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok || len(info.Counters) == 0 {
						return nil
					}
					var target game.TargetRef
					found := false
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							target, found = t, true
							break
						}
					}
					if !found {
						return nil
					}
					kinds := make([]string, 0, len(info.Counters))
					for k, n := range info.Counters {
						if n > 0 {
							kinds = append(kinds, k)
						}
					}
					sort.Strings(kinds)
					for _, k := range kinds {
						if err := (AddCounter{Target: target.ID, Kind: k, N: info.Counters[k]}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
				TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1)),
		},
	})
}
