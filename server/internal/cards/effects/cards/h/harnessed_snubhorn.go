package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harnessed Snubhorn — Creature — Dinosaur {3}{W}, 2/5 (March of the Machine:
// The Aftermath, #5476):
//
//	"Vigilance
//	 Whenever this creature deals combat damage to a player, return target
//	 artifact or enchantment card from your graveyard to the battlefield."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d) and
// re-checked as it resolves: a card that left the graveyard in response
// does nothing (CR 608.2b). The card enters under its controller's control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e1dd9dca-cb5f-46b4-b16d-a61ae651d653",
		Name:            "Harnessed Snubhorn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer("Harnessed Snubhorn — return target artifact or enchantment card from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					effects.MatReturnTargetToBattlefield(effects.NewContext(g, item))
					return nil
				}),
				effects.MatArtifactOrEnchantmentFromYourGraveyard("target artifact or enchantment card from your graveyard")),
		},
	})
}
