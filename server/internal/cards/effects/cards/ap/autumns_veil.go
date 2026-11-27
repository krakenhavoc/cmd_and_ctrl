package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Autumn's Veil — Instant {G} (Magic 2012, S82 OLD batch 452, #5573):
//
//	"Spells you control can't be countered by blue or black spells this
//	 turn, and creatures you control can't be the targets of blue or black
//	 spells this turn."
//
// Both halves are the #6263 seam already on develop: a counter-shield turn
// grant narrowed to counters made by blue or black spells
// (SpellsCantBeCounteredByColorsThisTurn) and a scoped "can't be the
// targets of <colour> spells" rule on creatures you control, until end of
// turn (CR 611.2c). The colour words are read through the spell's text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03f5946b-e000-4b39-a49f-1c31668b51bd",
		Name:         "Autumn's Veil",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			blue, black := ctx.TextColor("U"), ctx.TextColor("B")
			if err := (effects.GrantCounterShield{From: "Autumn's Veil", Grant: effects.SpellsCantBeCounteredByColorsThisTurn(
				"Spells you control can't be countered by blue or black spells this turn.",
				game.CounterShieldYouControl, game.PermissionFilter{}, blue+black)}).Apply(ctx); err != nil {
				return err
			}
			return effects.CantBeTargetedBySpellsOfColorsThisTurn{
				Colors: []string{blue, black},
				Label:  "Autumn's Veil — creatures you control can't be the targets of blue or black spells",
			}.Apply(ctx)
		},
	})
}
