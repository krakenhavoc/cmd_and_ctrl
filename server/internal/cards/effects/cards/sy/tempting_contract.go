package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tempting Contract — Artifact {4} (Commander 2021, #5504):
//
//	"At the beginning of your upkeep, each opponent may create a Treasure
//	 token. For each opponent who does, you create a Treasure token."
//
// Each opponent is asked in turn order (the opponents still to ask ride the
// prompt as its carry); a yes gives that opponent a Treasure and the
// controller one more. Nothing is created unasked: the controller makes a
// Treasure only for each opponent who accepts.
//
// No simplification.
func init() {
	contract := newTemptingContract()
	trigger := effects.AtYourUpkeep("Tempting Contract — each opponent may create a Treasure; for each who does, you create a Treasure",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			return contract.offer(ctx, ctx.Opponents())
		})
	trigger.Purpose = game.Purpose{Tokens: 1}
	effects.Register(effects.Spec{
		OracleID:     "cc4f1023-09e4-4f27-8712-be72e9694a80",
		Name:         "Tempting Contract",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}

// temptingContract asks one opponent at a time. The keyed answer is built
// by the constructor, so it can name the method that asks the next one.
type temptingContract struct {
	answer game.ConfirmThen
}

func newTemptingContract() *temptingContract {
	t := &temptingContract{}
	t.answer = effects.MayAnswer("confirm/tempting-contract", func(ctx *effects.Context, r game.ConfirmAnswer) error {
		if r.Accepted {
			if err := (effects.CreateToken{Controller: r.Chooser, Template: effects.TreasureToken(), N: 1}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.CreateToken{Controller: ctx.Controller(), Template: effects.TreasureToken(), N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return t.offer(ctx, r.Carry)
	})
	return t
}

func (t *temptingContract) offer(ctx *effects.Context, opponents []uuid.UUID) error {
	if len(opponents) == 0 {
		return nil
	}
	return effects.MayChoice{
		Player:   opponents[0],
		Question: "Tempting Contract — create a Treasure token? (the Contract's controller creates one too)",
		Then:     t.answer,
		Carry:    opponents[1:],
	}.Apply(ctx)
}
