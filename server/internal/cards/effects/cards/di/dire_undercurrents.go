package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dire Undercurrents — Enchantment {3}{U/B}{U/B} (Secret Lair Drop):
//
//	"Whenever a blue creature you control enters, you may have target player
//	 draw a card.
//	 Whenever a black creature you control enters, you may have target player
//	 discard a card."
//
// Two separate optional, targeted triggers. A creature that is both blue and
// black triggers both. The colour is read from the creature as it entered
// (its effective colours, CR 613), the target player is chosen as the trigger
// goes on the stack, and the "may" is asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7545adab-b6f8-496d-8c25-416c7ec85fa2",
		Name:         "Dire Undercurrents",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.On(game.EventETB, direUndercurrentsEntered("U"),
				"Dire Undercurrents — you may have target player draw a card",
				func(g *game.Game, item *game.StackItem) error {
					for _, t := range effects.NewContext(g, item).LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return g.DrawNForEffect(t.ID, 1)
						}
					}
					return nil
				}), effects.TargetPlayer("target player")),
				"Dire Undercurrents — have target player draw a card?"),
			effects.Optional(effects.Targeting(effects.On(game.EventETB, direUndercurrentsEntered("B"),
				"Dire Undercurrents — you may have target player discard a card",
				effects.TargetPlayerDiscards(1)), effects.TargetPlayer("target player")),
				"Dire Undercurrents — have target player discard a card?"),
		},
	})
}

// direUndercurrentsEntered is "a <colour> creature you control enters".
func direUndercurrentsEntered(color string) effects.When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
		return ok && c.IsCreature() && c.HasColor(color)
	}
}
