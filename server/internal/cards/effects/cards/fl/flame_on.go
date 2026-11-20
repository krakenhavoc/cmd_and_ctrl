package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "e715d33e-3d60-4623-b843-db0b906ae98b",
		Name:            "Flame On!",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         effects.TargetCreature("target creature"),
		// The flying waits for the counters' prompt (a CR 616 order; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				x := len(effects.GraveyardIDs(ctx.Game, ctx.Controller(), func(c game.Card) bool {
					return !c.IsCreature() && !c.IsLand()
				}))
				if x > 0 {
					return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"flying"},
					Label:    "Flame On! — flying",
				}.Apply(ctx)
			}),
		),
	})
}
