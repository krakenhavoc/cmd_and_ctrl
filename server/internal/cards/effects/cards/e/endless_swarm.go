package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Endless Swarm — Sorcery {5}{G}{G}{G} (Saviors of Kamigawa, #5601):
//
//	"Create a 1/1 green Snake creature token for each card in your hand.
//	 Epic (For the rest of the game, you can't cast spells. At the
//	 beginning of each of your upkeeps, copy this spell except for its epic
//	 ability.)"
//
// The count is the caster's hand as the spell (or its copy) resolves
// (CR 608.2h), so the spell itself is not in it. Epic is effects.Epic
// (CR 702.50a). The copies are made at upkeep, when the hand holds what the
// draw step has not yet added.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c62a9e10-cfc6-48cc-a3f7-68d624182a5a",
		Name:         "Endless Swarm",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: effects.Epic(func(_ *game.StackItem, ctx *effects.Context) error {
			n := effects.SokB1CardsInYourHand(ctx)
			if n <= 0 {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 green Snake"), N: n}.Apply(ctx)
		}),
	})
}
