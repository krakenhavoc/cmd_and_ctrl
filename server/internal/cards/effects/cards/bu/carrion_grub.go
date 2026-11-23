package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carrion Grub — Creature — Insect {3}{B}, 0/5 (Duskmourn Commander):
//
//	"This creature gets +X/+0, where X is the greatest power among creature
//	 cards in your graveyard.
//	 When this creature enters, mill four cards."
//
// The bonus is a layer 7c static read off your graveyard as it is now (the
// engine recomputes whenever a card crosses a graveyard boundary), the
// greatest printed power among its creature cards, 0 with none. A card with
// a * power counts as 0 there (CR 208.3). The enters trigger is Aftermath
// Analyst's mill.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c54b375d-a456-4886-96f2-8dc7a0024a5b",
		Name:         "Carrion Grub",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			Label:     "This creature gets +X/+0, where X is the greatest power among creature cards in your graveyard.",
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				_, power := effects.Cmdr2S30CreatureCardsIn(g, source.Controller)
				c.Power += power
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Carrion Grub — mill four cards", effects.Do(effects.MillCards{N: 4})),
		},
	})
}
