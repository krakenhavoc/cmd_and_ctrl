package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geode Golem — Artifact Creature — Golem {5}, 5/3:
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, you may cast
//	 your commander from the command zone without paying its mana cost.
//	 (You still pay any additional costs.)"
//
// The trigger offers the free cast through OfferFreeCommanderCast (#6242):
// the commander tax is still paid (CR 903.8) and nothing is asked when no
// commander of the controller's is in the command zone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aaed2f39-6d5d-4880-a7c4-5a045399dab3",
		Name:            "Geode Golem",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayer,
				"Geode Golem — you may cast your commander from the command zone without paying its mana cost",
				func(g *game.Game, item *game.StackItem) error {
					effects.OfferFreeCommanderCast(effects.NewContext(g, item), "Geode Golem")
					return nil
				}),
		},
	})
}
