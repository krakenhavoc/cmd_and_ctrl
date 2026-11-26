package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shatter Assumptions — Sorcery {1}{B}{B} (Modern Horizons):
//
//	"Choose one —
//	 • Target opponent reveals their hand and discards all colorless nonland
//	   cards.
//	 • Target opponent reveals their hand and discards all multicolored
//	   cards."
//
// Each bullet reveals the target opponent's hand (CR 701.20a) and that player
// discards every card of the stated kind: nothing is chosen, so no prompt
// goes to either player, and a hand with none is revealed and nothing is
// discarded (CR 609.3). The target is this mode's own and is re-checked at
// resolution (CR 608.2b). The cards are read as they are in the hand: a land
// is never "colorless nonland", and a multicolored card is one with two or
// more colors.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1372f18-9b33-40e8-bfe2-dc4ca3b3ef55",
		Name:         "Shatter Assumptions",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target opponent reveals their hand and discards all colorless nonland cards.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				shatterAssumptionsDiscardAll(effects.And(effects.Nonland(), effects.Colorless()))), game.Purpose{Answers: game.AnswerValue}),
			effects.ModeWithPurpose(effects.ModeDoing("Target opponent reveals their hand and discards all multicolored cards.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				shatterAssumptionsDiscardAll(effects.Multicolored())), game.Purpose{Answers: game.AnswerValue}),
		),
	})
}

// shatterAssumptionsDiscardAll is one bullet's body: reveal the mode's target
// player's hand, then that player discards every card of it the filter
// accepts.
func shatterAssumptionsDiscardAll(filter effects.CardPredicate) func(*game.StackItem, *effects.Context, int) error {
	return func(_ *game.StackItem, ctx *effects.Context, occ int) error {
		t, ok := effects.ModeTarget(ctx, occ)
		if !ok || t.Kind != game.TargetPlayer {
			return nil
		}
		p := ctx.Game.PlayerByIDForEffect(t.ID)
		if p == nil {
			return nil
		}
		ctx.Game.RevealHandForEffect(t.ID)
		var ids []uuid.UUID
		for _, c := range p.Hand.Cards {
			if filter(ctx.Game, ctx.Controller(), c) {
				ids = append(ids, c.InstanceID)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		return ctx.Game.DiscardCardsKeyedForEffect(t.ID, ids, ctx.Source(), game.KeyedTail{})
	}
}
