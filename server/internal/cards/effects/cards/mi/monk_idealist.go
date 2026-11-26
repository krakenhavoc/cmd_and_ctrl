package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monk Idealist — {2}{W} Creature — Human Monk Cleric 2/2 (Eternal Masters,
// S82 OLD, #5542):
//
//	"When this creature enters, return target enchantment card from your
//	 graveyard to your hand."
//
// An enters trigger with a graveyard target chosen as it goes on the stack;
// the card is returned if it is still there as the ability resolves
// (CR 608.2b). With no enchantment card in the graveyard there is no target
// and the trigger is removed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f47ff5a-dd93-46de-a1ca-b412631218bf",
		Name:         "Monk Idealist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Monk Idealist — return target enchantment card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
				}),
				effects.TargetCardInGraveyard("target enchantment card from your graveyard", effects.YouOwn(), effects.Enchantment())),
		},
	})
}
