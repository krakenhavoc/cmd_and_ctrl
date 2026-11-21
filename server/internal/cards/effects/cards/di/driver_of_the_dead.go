package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Driver of the Dead — Creature — Vampire {3}{B}, 3/2 (Foundations,
// #4465):
//
//	"When this creature dies, return target creature card with mana value
//	 2 or less from your graveyard to the battlefield."
//
// Devoted Crop-Mate's target on a dies trigger. The target is chosen as
// the trigger goes on the stack, from your graveyard, where the Driver
// now is; its own mana value is 4, so it can never choose itself. The
// card returns if it is still a legal target as the trigger resolves (CR
// 608.2b). With no such card, the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c179e54-3beb-4761-bddf-41fa98b082db",
		Name:         "Driver of the Dead",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Driver of the Dead — return target creature card with mana value 2 or less from your graveyard to the battlefield",
					effects.ReturnFirstLegalGraveyardTargetToBattlefield),
				effects.TargetCardInGraveyard("target creature card with mana value 2 or less from your graveyard",
					effects.Creature(), effects.YouOwn(), effects.ManaValueLE(2))),
		},
	})
}
