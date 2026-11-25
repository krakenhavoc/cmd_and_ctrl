package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Restoration Gearsmith — Creature — Human Artificer {2}{W}{B}, 3/3 (Kaladesh,
// #5539):
//
//	"When this creature enters, return target artifact or creature card from
//	 your graveyard to your hand."
//
// The target is chosen as the trigger goes on the stack and is judged again as
// it resolves (CR 608.2b): a card that left the graveyard in response is not
// returned. The trigger is mandatory, so with a card in the graveyard the
// controller must pick one. An artifact creature card is both, and counts once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "447c7fe0-6ec7-4163-9d9c-f08f0a4c7109",
		Name:         "Restoration Gearsmith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Restoration Gearsmith — return target artifact or creature card from your graveyard to your hand", effects.ReturnTargetedCardToHand),
				effects.TargetCardInGraveyard("target artifact or creature card from your graveyard",
					effects.Or(effects.Artifact(), effects.Creature()), effects.YouOwn())),
		},
	})
}
