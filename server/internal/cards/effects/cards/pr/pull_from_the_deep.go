package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pull from the Deep — Sorcery {2}{U}{U} (Journey into Nyx, S82 OLD, #5558):
//
//	"Return up to one target instant card and up to one target sorcery card from
//	 your graveyard to your hand. Exile Pull from the Deep."
//
// Two optional target clauses, each a card of its type in your graveyard; the
// ones still legal return (CR 608.2b), and the spell then exiles itself rather
// than going to the graveyard (Burning Wish's move). With neither target
// chosen the spell still resolves and exiles itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c0a72dc-4756-4d60-b6b4-4c2c69699efd",
		Name:         "Pull from the Deep",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("up to one target instant card from your graveyard", effects.YouOwn(), effects.Instant()).WithCount(0, 1),
			effects.TargetCardInGraveyard("up to one target sorcery card from your graveyard", effects.YouOwn(), effects.Sorcery()).WithCount(0, 1),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := effects.ReturnLegalGraveyardTargetsToHand(ctx); err != nil {
				return err
			}
			if effects.SourceIsNewObject(ctx.Game, ctx.Item) {
				return nil
			}
			return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
		},
	})
}
