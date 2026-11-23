package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spirit Water Revival — Sorcery {1}{U}{U} (Avatar: The Last Airbender,
// #4459):
//
//	"As an additional cost to cast this spell, you may waterbend {6}.
//	 (While paying a waterbend cost, you can tap your artifacts and
//	 creatures to help. Each one pays for {1}.)
//	 Draw two cards. If this spell's additional cost was paid, instead
//	 shuffle your graveyard into your library, draw seven cards, and you
//	 have no maximum hand size for the rest of the game.
//	 Exile Spirit Water Revival."
//
// The waterbend is an optional additional cost (OptionalWaterbend, CR
// 701.67a), the shape Secret of Bloodbending uses. The paid branch is
// Finale of Revelation's: the graveyard is shuffled in (the spell is on
// the stack, so it is not among it), seven cards are drawn, and the
// hand-size grant is the player-level one that outlives the spell. Either
// branch ends with the spell exiling itself (CR 608.2n).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "68979160-b5ce-4787-8a1e-1f40e614c3b0",
		Name:          "Spirit Water Revival",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Draws: 2},
		OptionalCosts: []game.AdditionalCost{effects.OptionalWaterbend("{6}")},
		OnResolve: effects.ResolveSteps(
			// The hand-size grant and the exile wait for a prompt the draw asks (CR 608.2c).
			effects.Instruction(func(ctx *effects.Context) error {
				if ctx.OptionalCostTimes(game.WaterbendKey) == 0 {
					return effects.DrawCards{N: 2}.Apply(ctx)
				}
				if err := effects.FinaleShuffleGraveyardIntoLibrary(ctx, ctx.Controller()); err != nil {
					return err
				}
				return effects.DrawCards{N: 7}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if ctx.OptionalCostTimes(game.WaterbendKey) > 0 {
					if err := ctx.Game.SetMaxHandSizeForEffect(ctx.Controller(), game.NoMaxHandSize); err != nil {
						return err
					}
				}
				return (effects.ExileTarget{Target: ctx.Source()}).Apply(ctx)
			}),
		),
	})
}
