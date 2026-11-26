package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dream Salvage — Instant {U/B} (Shadowmoor, #5591):
//
//	"Draw cards equal to the number of cards target opponent discarded
//	 this turn."
//
// The count is the target opponent's turn tally of cards discarded (one
// per card, CR 701.9a), read as the spell resolves, whatever the cause and
// whoever made them discard. An opponent who has become an illegal target
// makes the spell fizzle (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a05d478c-a703-4fc4-9496-2894a6acdaef",
		Name:         "Dream Salvage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				n := ctx.Game.TurnTallyFor(t.ID).CardsDiscarded
				return effects.DrawCards{Player: ctx.Controller(), N: n}.Apply(ctx)
			}
			return nil
		},
	})
}
