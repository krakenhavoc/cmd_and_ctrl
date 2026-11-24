package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pulse of Murasa — Instant {2}{G} (Commander 2021, #5504):
//
//	"Return target creature or land card from a graveyard to its owner's
//	 hand. You gain 6 life."
//
// Any graveyard is a legal pick; the card goes to its OWNER's hand (the
// graveyard it sat in), and the life is the caster's. With the target gone
// by resolution the spell does nothing, so no life is gained (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5b446a4-a761-42f0-acf0-ea0a1a9538d6",
		Name:         "Pulse of Murasa",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("target creature or land card from a graveyard", effects.Or(effects.Creature(), effects.Land())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if _, ok := effects.B16FirstLegalTargetCard(ctx); !ok {
				return nil
			}
			if err := effects.ReturnLegalGraveyardTargetsToHand(ctx); err != nil {
				return err
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: 6}.Apply(ctx)
		},
	})
}
