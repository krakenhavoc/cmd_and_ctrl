package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Force of Rage — Instant {1}{R}{R} (Modern Horizons):
//
//	"If it's not your turn, you may exile a red card from your hand rather
//	 than pay this spell's mana cost.
//	 Create two 3/1 red Elemental creature tokens with trample and haste.
//	 Sacrifice those tokens at the beginning of your next upkeep."
//
// Force of Vigor's pitch (an alternative cost gated on "not your turn", the
// exiled red card named as the cost is paid). The two Elementals are created
// together, and the tail schedules one delayed trigger over the tokens that
// landed (CR 603.7): at the beginning of its controller's next upkeep, which
// is the next of their own turns even when the spell was cast on their own
// turn, it sacrifices them (a token that has already left is skipped).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c019ed4d-33a3-48be-a38e-93e5d6161b1f",
		Name:         "Force of Rage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		AlternativeCosts: []game.AlternativeCost{
			{
				Key:           "pitch",
				Label:         "If it's not your turn, exile a red card from your hand",
				ExileFromHand: effects.CardInYourHand("a red card from your hand", effects.OfColor("R")),
				PayLabel:      "a red card from your hand",
				Condition: func(g *game.Game, controller uuid.UUID) bool {
					return !effects.IsActivePlayer(g, controller)
				},
			},
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: ctx.Controller(),
				Source:     ctx.Source(),
				Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("3/1 red Elemental with trample and haste")), Count: 2}},
			}, ctx.ItemTail(forceOfRageSacrifice))
		},
	})
}

// forceOfRageSacrifice is "sacrifice those tokens at the beginning of your
// next upkeep", over the tokens that landed. An on-disk key.
var forceOfRageSacrifice = effects.CardTail("s82-old-b162/force-of-rage-sacrifice-the-elementals", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Controller:         ctx.Controller(),
		Label:              "Force of Rage — sacrifice the Elementals",
		Cards:              append([]uuid.UUID(nil), o.IDs...),
		Body:               effects.SacrificeListedCardsBody,
	}.Apply(ctx)
})
