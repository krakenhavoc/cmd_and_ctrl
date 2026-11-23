package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loot, the Key to Everything — Legendary Creature — Beast Noble {G}{U}{R},
// 3/3 (The Big Score, #4468):
//
//	"Ward {1}
//	 At the beginning of your upkeep, exile the top X cards of your library,
//	 where X is the number of card types among other nonland permanents you
//	 control. You may play those cards this turn."
//
// X is counted as the trigger resolves, through the layered view, over the
// card types of the controller's other nonland permanents (CR 205.2a's
// nonland types: artifact, battle, creature, enchantment, kindred,
// planeswalker): an Artifact Creature is two. The exiled cards may be played
// (lands included, spells cast normally) until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "542be13f-b9c0-4b79-b486-be576c531a82",
		Name:            "Loot, the Key to Everything",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"ward"},
		Purpose:         game.Purpose{Impulse: 1},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{1}"), "Loot, the Key to Everything — ward {1}"),
			effects.AtYourUpkeep("Loot, the Key to Everything — exile the top X cards of your library; you may play them this turn",
				func(g *game.Game, item *game.StackItem) error {
					x := effects.B35CardTypesAmongOtherNonlandPermanents(g, item.Controller, item.SourceCardID)
					if x <= 0 {
						return nil
					}
					return effects.ExileTopWithPermission{
						From:    item.Controller,
						GrantTo: item.Controller,
						N:       x,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
