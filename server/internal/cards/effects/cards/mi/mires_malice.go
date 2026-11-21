package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mire's Malice — Sorcery {3}{B}:
//
//	"Target opponent discards two cards.
//	 Awaken 3—{5}{B}"
//
// ADR 0135 §3 (#2411): the opponent's discard of two cards of their
// choice, then the awaken land (CR 702.113a). The discard is the
// opponent's prompt, queued as Mind Rot's is; nothing in the awaken reads
// which cards they pick.
//
// No simplifications.
func init() {
	t := effects.TargetPlayer("target opponent", effects.Opponent())
	effects.Register(effects.Spec{
		OracleID:     "1e7b18a3-43eb-4491-bb19-99f261a94719",
		Name:         "Mire's Malice",
		Completeness: effects.CompletenessFull,
		Targets:      t,
		AlternativeCosts: []game.AlternativeCost{
			effects.Awaken(3, "{5}{B}", t),
		},
		OnResolve: effects.AwakenAfter(3, func(item *game.StackItem, ctx *effects.Context) error {
			p, ok := ctx.ClauseTarget(0)
			if !ok || p.Kind != game.TargetPlayer {
				return nil
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player:   p.ID,
				Source:   item.SourceCardID,
				N:        2,
				Question: "Mire's Malice — discard two cards",
			})
			return nil
		}),
	})
}
