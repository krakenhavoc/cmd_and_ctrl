package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smitten Swordmaster // Curry Favor — Adventure creature:
//
//	Smitten Swordmaster {1}{B} Creature — Human Knight, 2/1: "Lifelink"
//	Curry Favor {B} Sorcery — Adventure: "You gain X life and each
//	 opponent loses X life, where X is the number of Knights you control."
//
// The adventure half is registered under "<oracle_id>#1". X is counted
// as Curry Favor resolves (CR 608.2h), through the layered view so a
// changeling counts. The loss is life loss, to each opponent still in
// the game.
//
// No simplification.
func init() {
	const oracle = "11ad6e75-9bbd-40ea-b04a-e6a61db3bcda"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Smitten Swordmaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Curry Favor",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := effects.B43CreaturesOfSubtypeControlled(ctx.Game, ctx.Controller(), ctx.TextSubtype("Knight"))
			if x <= 0 {
				return nil
			}
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, ctx.Controller(), x); err != nil {
				return err
			}
			for _, opp := range ctx.Opponents() {
				if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, opp, -x); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
