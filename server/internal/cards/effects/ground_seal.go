package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ground Seal — Enchantment {1}{G}:
//
//	"When this enchantment enters, draw a card.
//	 Cards in graveyards can't be the targets of spells or abilities."
//
// The static is ADR 0109 §6's TargetingRestrictions (#1885), read at the
// engine's two targeting choke points: no spell or ability may target a
// card in any graveyard while the Seal is on the battlefield, its
// controller's included, and one already aimed at a graveyard card loses
// that target at resolution (CR 601.2c, 608.2b). A cost or a "choose"
// that is not a target (delve, "exile a card from a graveyard") is not
// stopped. The enters trigger is the ordinary cantrip.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13f6f960-ef79-4f2c-8874-90fe9e77099e",
		Name:         "Ground Seal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ground Seal — draw a card", Do(DrawCards{N: 1})),
		},
		TargetingRestrictions: []game.TargetingRestriction{
			CardsInGraveyardsCantBeTargeted("Cards in graveyards can't be the targets of spells or abilities."),
		},
	})
}
