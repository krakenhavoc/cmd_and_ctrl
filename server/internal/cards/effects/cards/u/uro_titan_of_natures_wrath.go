package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uro, Titan of Nature's Wrath — Legendary Creature — Elder Giant
// {1}{G}{U}, 6/6:
//
//	"When Uro enters, sacrifice it unless it escaped.
//	 Whenever Uro enters or attacks, you gain 3 life and draw a card,
//	 then you may put a land card from your hand onto the battlefield.
//	 Escape—{G}{G}{U}{U}, Exile five other cards from your graveyard."
//
// Phlage's sibling, and the one the seam list held back twice: it
// needed "you may put a land card from your hand onto the
// battlefield" (#654, shipped as PutFromHandOntoBattlefield) and then
// it needed the permanent to remember it escaped (#653, this issue).
// Both exist now, so it ships Full.
//
// The land goes in the CONTINUATION, not on the next line. The put is
// a PROMPT: Apply returns once the question is queued, so anything
// written after it would run before the answer. Uro has nothing after
// it, which is why the clause reads in printed order here — but the
// draw and the life gain go FIRST, which is the printed order too
// ("you gain 3 life and draw a card, THEN you may put a land").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "ee302659-59ed-4eef-babe-451b9ccf7f14",
		Name:             "Uro, Titan of Nature's Wrath",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Escape("{G}{G}{U}{U}", 5)},
		Triggered: []game.TriggeredAbility{
			effects.SacrificeThisUnlessItEscaped("Uro"),
			effects.WhenThisEntersOrAttacks("Uro — gain 3 life, draw a card, then you may put a land from your hand", effects.Do(
				// Each instruction waits for the prompts the one before asks (CR 608.2c, #4155).
				effects.GainLife{Amount: 3},
				effects.DrawCards{N: 1},
				effects.Instruction(func(ctx *effects.Context) error {
					put := effects.MayPutALandFromHand("Uro")
					put.Player = ctx.Controller()
					return put.Apply(ctx)
				}),
			)),
		},
	})
}
