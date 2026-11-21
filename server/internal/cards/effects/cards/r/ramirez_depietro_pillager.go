package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ramirez DePietro, Pillager — Legendary Creature — Human Pirate {2}{U}{B}, 4/3:
//
//	"When Ramirez DePietro enters, you lose 2 life and create two Treasure
//	 tokens.
//	 Whenever one or more Pirates you control deal combat damage to a
//	 player, exile the top card of that player's library. You may cast that
//	 card for as long as it remains exiled."
//
// The second ability is the batched Pirate damage trigger, once per player
// hit in a combat damage step (CR 603.2c); several Pirates connecting with
// one player are one trigger. The exiled card is cast for its mana cost
// (no "spend mana as though any type") and the window has no end but the
// card leaving exile; a land stays where it is.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3b8690e-ac47-444f-924e-ab2a12a85431",
		Name:         "Ramirez DePietro, Pillager",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ramirez DePietro — you lose 2 life and create two Treasure tokens",
				effects.Do(effects.GainLife{Amount: -2}, effects.CreateToken{Template: effects.TreasureToken(), N: 2})),
			effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(effects.OfCreatureType("Pirate"),
				"Ramirez DePietro — exile the top card of that player's library; you may cast it for as long as it remains exiled",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.ExileTopWithPermission{
						From:        item.Trigger.Event.Target,
						GrantTo:     item.Controller,
						N:           1,
						CastOnly:    true,
						WhileExiled: true,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
