package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pillardrop Rescuer — Creature — Spirit Cleric {4}{W}, 2/2 (Strixhaven:
// School of Mages, #5502):
//
//	"Flying
//	 When this creature enters, return target creature card with mana value
//	 3 or less from your graveyard to your hand."
//
// Flying is the printed keyword; the enters trigger is Nimraiser Paladin's
// graveyard target, returned to your hand if it is still there as the
// trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5cf9002d-6c04-420b-8c3b-0a17f674bb09",
		Name:            "Pillardrop Rescuer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Pillardrop Rescuer — return target creature card with mana value 3 or less from your graveyard to your hand", effects.ReturnTargetedCardToHand),
				effects.TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard", effects.Creature(), effects.ManaValueLE(3), effects.YouOwn())),
		},
	})
}
