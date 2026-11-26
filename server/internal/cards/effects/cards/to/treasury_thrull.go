package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treasury Thrull — Creature — Thrull {4}{W}{B}, 4/4 (Commander 2015, S82 OLD
// #5546):
//
//	"Extort (Whenever you cast a spell, you may pay {W/B}. If you do, each
//	 opponent loses 1 life and you gain that much life.)
//	 Whenever this creature attacks, you may return target artifact,
//	 creature, or enchantment card from your graveyard to your hand."
//
// Extort is the shared trigger (Kingpin's Pet's). The attack trigger is
// Bishop of Rebirth's shape: the card is chosen as the trigger goes on the
// stack (CR 603.3d), the "may" is asked as it resolves, and the card must
// still be an artifact, creature or enchantment card in your graveyard
// (CR 608.2b).
//
// No simplification.
func init() {
	t := effects.Targeting(
		effects.WheneverThisAttacks("Treasury Thrull — return target artifact, creature, or enchantment card from your graveyard to your hand",
			effects.ReturnTargetCardToHand),
		effects.TargetCardInGraveyard("target artifact, creature, or enchantment card from your graveyard",
			effects.YouOwn(), effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment())),
	)
	t = effects.Optional(t, "Treasury Thrull — return a card from your graveyard to your hand?")
	effects.Register(effects.Spec{
		OracleID:     "34161ec4-5d99-45e6-b43c-c8d28958660a",
		Name:         "Treasury Thrull",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Extort("Treasury Thrull"), t},
	})
}
