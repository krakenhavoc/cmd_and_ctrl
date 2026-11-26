package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pharika's Mender — Creature — Gorgon {3}{B}{G}, 4/3 (Conspiracy: Take the
// Crown, S82 OLD, #5540):
//
//	"When this creature enters, you may return target creature or
//	 enchantment card from your graveyard to your hand."
//
// Midnight Scavengers' optional enters-return with a creature-or-enchantment
// filter: the target is chosen as the trigger goes on the stack and the
// "you may" is asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00771a33-7443-43f9-a61e-1737b62b1f81",
		Name:         "Pharika's Mender",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Pharika's Mender — you may return target creature or enchantment card from your graveyard to your hand", effects.ReturnTargetedCardToHand),
				effects.TargetCardInGraveyard("target creature or enchantment card from your graveyard", effects.Or(effects.Creature(), effects.Enchantment()), effects.YouOwn())),
				"Pharika's Mender — return the card to your hand?"),
		},
	})
}
