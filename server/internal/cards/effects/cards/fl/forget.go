package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forget — Sorcery {U}{U} (Homelands, S82 OLD b826, #5635):
//
//	"Target player discards two cards, then draws as many cards as they
//	 discarded this way."
//
// The target chooses and discards the two cards (CR 701.9a: the player
// discarding picks them), a hand of one discards that one, and an empty hand
// discards none. Once the discard has settled the same player draws the number
// of cards that really left their hand, so the draw is sized off the keyed
// discard continuation rather than off the printed two.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "619ef7e1-33cd-4470-a1d4-83c5f1f5c31e",
		Name:         "Forget",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			who := effects.FirstLegalTargetPlayer(ctx)
			if who == uuid.Nil {
				return nil
			}
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player: who,
				Source: item.SourceCardID,
				N:      2,
			}, forgetDiscarded, nil)
		},
	})
}

// forgetDiscarded is "then draws as many cards as they discarded this way".
// An on-disk key: never renamed, never reused.
var forgetDiscarded = game.RegisterDiscardThen("choose-cards/s82-old-b826-forget-discarded", func(g *game.Game, r game.DiscardsDone) error {
	n := r.Discarded.Count()
	if n <= 0 {
		return nil
	}
	return g.DrawNForEffect(r.Player, n)
})
