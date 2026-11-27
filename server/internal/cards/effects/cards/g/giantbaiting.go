package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Giantbaiting — Sorcery {2}{R/G} (Shadowmoor, S82 older sets):
//
//	"Create a 4/4 red and green Giant Warrior creature token with haste.
//	 Exile it at the beginning of the next end step.
//	 Conspire (As you cast this spell, you may tap two untapped creatures you
//	 control that share a color with it. When you do, copy it.)"
//
// Zektar Shrine Expedition's token and delayed exile: only the token this
// resolution made is exiled, at the next end step, wherever it then is. The
// conspire copy makes a second token and its own exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "894a220b-1372-4a76-b38f-bf0d7001af41",
		Name:          "Giantbaiting",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
		OptionalCosts: []game.AdditionalCost{effects.Conspire()},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			made, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(effects.TokenCard("4/4 red and green Giant Warrior with haste")), 1, game.TokenEntryOptions{})
			if err != nil || len(made) == 0 {
				return err
			}
			return effects.ScheduleDelayedTrigger{
				At:         game.StepEnd,
				Controller: ctx.Controller(),
				Label:      "Giantbaiting — exile the Giant Warrior token",
				Cards:      made,
				Body:       effects.ExileListedCardsBody,
			}.Apply(ctx)
		},
	})
}
