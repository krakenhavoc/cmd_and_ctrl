package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wistful Thinking — Sorcery {2}{U} (Planar Chaos, #5595):
//
//	"Target player draws two cards, then discards four cards."
//
// Cephalid Broker's target-player loot as a spell: drawThenDiscard draws
// two, then asks that player for four discards from the post-draw hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a936b84c-c350-4a1f-b4a7-86316f11c67f",
		Name:         "Wistful Thinking",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, Discards: 4}),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					return effects.DrawThenDiscard(ctx.Game, t.ID, item.SourceCardID, 2, 4, "")
				}
			}
			return nil
		},
	})
}
