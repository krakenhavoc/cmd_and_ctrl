package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghastlord of Fugue — Creature — Spirit Avatar {U/B}{U/B}{U/B}{U/B}{U/B}, 4/4
// (Shadowmoor, S82 OLD b530, #5591):
//
//	"This creature can't be blocked.
//	 Whenever this creature deals combat damage to a player, that player
//	 reveals their hand. You choose a card from it. That player exiles that
//	 card."
//
// Benthic Infiltrator's can't-be-blocked restriction, and Appetite for Brains'
// reveal-and-exile of a card the controller chooses from the damaged player's
// hand. The hand is revealed to every player (CR 701.20a); a hand with no cards
// is revealed and nothing is exiled. The player is read off the damage event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fce163aa-c5d0-4147-8566-512321c0be8e",
		Name:         "Ghastlord of Fugue",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayer,
				"Ghastlord of Fugue — that player reveals their hand; you choose a card from it and they exile it",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.ChooseFromRevealedHand{
						Player: item.Trigger.Event.Target,
						Label:  "card",
						Exile:  true,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
