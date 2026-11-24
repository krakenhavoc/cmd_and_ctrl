package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Go Blank — Sorcery {2}{B} (Strixhaven: School of Mages, S82 batch 047,
// #5502):
//
//	"Target player discards two cards. Then exile that player's graveyard."
//
// Mind Rot's discard (the target picks, and a hand of one or none discards
// what it has), with the exile waiting on it as the registered continuation
// of the discard prompt (Mind Roots' shape): "then" means the cards just
// discarded are in the graveyard that is exiled. It runs at once when
// nothing was asked (an empty hand), so the graveyard is exiled either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60186579-285b-4034-a782-431f79df5b07",
		Name:         "Go Blank",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			victim := effects.TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player:   victim,
				Source:   item.SourceCardID,
				N:        2,
				Question: "Go Blank — discard two cards",
				Asked:    game.NewAskedItem(item),
			}, goBlankExile, nil)
		},
	})
}

// goBlankExile is the clause after the discard: exile the discarding
// player's graveyard. An on-disk key.
var goBlankExile = game.RegisterDiscardThen("choose-cards/go-blank-exile", func(g *game.Game, r game.DiscardsDone) error {
	if r.Item == nil {
		return nil
	}
	return effects.ExileGraveyardForEffect(g, r.Item, r.Player)
})
