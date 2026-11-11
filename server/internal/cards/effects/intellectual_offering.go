package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Intellectual Offering — Instant {4}{U} (Bloomburrow Commander):
//
//	"Choose an opponent. You and that player each draw three cards.
//	 Choose an opponent. Untap all nonland permanents you control and all
//	 nonland permanents that player controls."
//
// Two separate choices, so the second opponent may differ from the first;
// each is a real prompt that stays a restore point (keyed). With no
// opponent left to choose, the player half of each clause still happens
// for you alone.
func init() {
	Register(Spec{
		OracleID:     "73ec7a0b-274a-4a10-a173-ff1625d62fce",
		Name:         "Intellectual Offering",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChoosePlayer{
				Among:    Opponents,
				Question: "Intellectual Offering — choose an opponent; you and that player each draw three cards",
				ThenKey:  intellectualOfferingDraw,
			}.Apply(ctx)
		},
	})
}

var intellectualOfferingDraw = ChoosePlayerThen("option-pick/intellectual-offering-draw", func(ctx *Context, chosen uuid.UUID, _ []uuid.UUID) error {
	if err := (DrawCards{N: 3}).Apply(ctx); err != nil {
		return err
	}
	if chosen != uuid.Nil {
		if err := (DrawCards{Player: chosen, N: 3}).Apply(ctx); err != nil {
			return err
		}
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Intellectual Offering — choose an opponent; untap all nonland permanents you and that player control",
		ThenKey:  intellectualOfferingUntap,
	}.Apply(ctx)
})

var intellectualOfferingUntap = ChoosePlayerThen("option-pick/intellectual-offering-untap", func(ctx *Context, chosen uuid.UUID, _ []uuid.UUID) error {
	nonland := func(c game.Card) bool { return !c.IsLand() }
	if err := b16UntapAllYouControlMatching(ctx, ctx.Controller(), nonland); err != nil {
		return err
	}
	if chosen == uuid.Nil {
		return nil
	}
	return b16UntapAllYouControlMatching(ctx, chosen, nonland)
})
