package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flame On! — Sorcery {4}{R}:
//
//	"Put X +1/+1 counters on target creature, where X is the number of
//	 noncreature, nonland cards in your graveyard. It gains flying until
//	 end of turn.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// X is counted as the spell resolves, while Flame On! itself is still
// on the stack, so it never counts itself. With X of zero the creature
// still gains flying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e715d33e-3d60-4623-b843-db0b906ae98b",
		Name:            "Flame On!",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			x := len(graveyardIDs(ctx.Game, ctx.Controller(), func(c game.Card) bool {
				return !c.IsCreature() && !c.IsLand()
			}))
			if x > 0 {
				if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: x}).Apply(ctx); err != nil {
					return err
				}
			}
			return GrantKeywordUntilEOT{
				Target:   id,
				Keywords: []string{"flying"},
				Label:    "Flame On! — flying",
			}.Apply(ctx)
		},
	})
}
