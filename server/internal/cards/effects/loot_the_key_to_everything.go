package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "542be13f-b9c0-4b79-b486-be576c531a82",
		Name:            "Loot, the Key to Everything",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"ward"},
		Purpose:         game.Purpose{Impulse: 1},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{1}"), "Loot, the Key to Everything — ward {1}"),
			AtYourUpkeep("Loot, the Key to Everything — exile the top X cards of your library; you may play them this turn",
				func(g *game.Game, item *game.StackItem) error {
					x := b35CardTypesAmongOtherNonlandPermanents(g, item.Controller, item.SourceCardID)
					if x <= 0 {
						return nil
					}
					return ExileTopWithPermission{
						From:    item.Controller,
						GrantTo: item.Controller,
						N:       x,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
