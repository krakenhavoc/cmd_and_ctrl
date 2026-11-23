package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savior of the Small — Creature — Kor Survivor {3}{W}, 3/4 (Duskmourn:
// House of Horror, #4466):
//
//	"Survival — At the beginning of your second main phase, if this
//	 creature is tapped, return target creature card with mana value 3
//	 or less from your graveyard to your hand."
//
// Defiant Survivor's Survival trigger (survival.go), its "if" asked as
// the phase begins and again as it resolves (CR 603.4), with Wayspeaker
// Bodyguard's graveyard return (returnTargetGraveyardCardToHand) narrowed
// to a creature card with mana value 3 or less you own. The target is
// checked again as the trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15f2085f-d5e6-4bc6-870f-00fcaccd981b",
		Name:         "Savior of the Small",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Survival("Savior of the Small — return target creature card with mana value 3 or less from your graveyard to your hand",
				effects.ReturnTargetGraveyardCardToHand),
				effects.TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard",
					effects.Creature(), effects.ManaValueLE(3), effects.YouOwn())),
		},
	})
}
