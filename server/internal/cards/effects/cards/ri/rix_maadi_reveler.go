package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rix Maadi Reveler — Creature — Human Shaman {1}{R}, 2/2 (Ravnica
// Allegiance):
//
//	"Spectacle {2}{B}{R} (You may cast this spell for its spectacle cost
//	 rather than its mana cost if an opponent lost life this turn.)
//	 When this creature enters, discard a card, then draw a card. If this
//	 creature's spectacle cost was paid, instead discard your hand, then draw
//	 three cards."
//
// One trigger that reads the cast provenance as it resolves (the entering
// object's record, so it is the same answer if the Reveler has left). The
// ordinary mode is The Ruinous Wrecking Crew's keyed "discard a card, then
// draw a card"; the spectacle mode discards the whole hand (a hand with
// nothing in it still draws) and draws three.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "bad5b8d4-9089-4ebd-8eaa-07552737c527",
		Name:             "Rix Maadi Reveler",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Spectacle("{2}{B}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rix Maadi Reveler — discard a card, then draw a card (spectacle: discard your hand, then draw three cards)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if obj := ctx.Trigger().Object; obj != nil && obj.AltCost == "spectacle" {
						if _, err := effects.DiscardWholeHand(g, ctx.Controller()); err != nil {
							return err
						}
						return effects.DrawCards{N: 3}.Apply(ctx)
					}
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player: item.Controller,
						Source: item.SourceCardID,
						N:      1,
					}, effects.TheRuinousWreckingCrewDraw, nil)
				}),
		},
	})
}
