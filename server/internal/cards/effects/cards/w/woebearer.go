package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woebearer — Creature — Zombie {4}{B}, 2/3 (Mirrodin, #5606):
//
//	"Fear (This creature can't be blocked except by artifact creatures and/or
//	 black creatures.)
//	 Whenever this creature deals combat damage to a player, you may return
//	 target creature card from your graveyard to your hand."
//
// Mnemonic Wall's optional return on a combat-damage trigger: the target is
// chosen as the trigger goes on the stack, and the "you may" is asked as it
// resolves (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5ed93d71-3f62-49d4-a9ca-5728c44b0633",
		Name:            "Woebearer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WheneverThisDealsCombatDamageToAPlayer("Woebearer — return target creature card from your graveyard to your hand", effects.ReturnTargetedCardToHand),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn())),
				"Woebearer — return the creature card to your hand?"),
		},
	})
}
