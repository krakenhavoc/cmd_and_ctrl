package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scythecat Cub — Creature — Cat {1}{G}, 2/2 (Foundations Jumpstart, S82
// BNS):
//
//	"Trample
//	 Landfall — Whenever a land you control enters, put a +1/+1 counter on
//	 target creature you control. If this is the second time this ability
//	 has resolved this turn, double the number of +1/+1 counters on that
//	 creature instead."
//
// The count is Belladonna Took's: the per-turn tally of this Cub's
// resolutions of this ability, which includes the one resolving
// (Game.ResolvedThisTurn), per object (CR 400.7). The second resolution
// doubles (puts as many more +1/+1 counters as the creature has); every
// other resolution puts one.
//
// No simplification.
func init() {
	landfall := Landfall(scythecatCubLabel, scythecatCubResolves)
	landfall.Targets = TargetCreature("target creature you control", YouControl())
	Register(Spec{
		OracleID:        "929117cd-caa6-402b-9d7b-1f684d75695c",
		Name:            "Scythecat Cub",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{landfall},
	})
}

const scythecatCubLabel = "Scythecat Cub — put a +1/+1 counter on target creature you control (the second time this turn, double its +1/+1 counters instead)"

func scythecatCubResolves(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	n := 1
	if g.ResolvedThisTurn(item.SourceCardID, scythecatCubLabel) == 2 {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return nil
		}
		n = c.Counters[game.CounterPlusOne]
		if n <= 0 {
			return nil
		}
	}
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
}
