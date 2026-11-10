package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Pyramid of Mars — Plane — Mars (Doctor Who, #4870):
//
//	"When you planeswalk to The Pyramid of Mars and at the beginning of
//	 your upkeep, surveil 2.
//	 Whenever chaos ensues, return target creature card from your
//	 graveyard to the battlefield."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The chaos ability (CR 311.7) targets as it goes on the stack and returns
// the card under its owner's control (it is the planar controller's own).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5910ae7e-7a74-4296-9ae1-7fa093b4e640",
		Name:         "The Pyramid of Mars",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("The Pyramid of Mars — surveil 2", Do(Surveil{N: 2})),
			Targeting(WheneverChaosEnsues("The Pyramid of Mars — return target creature card from your graveyard to the battlefield",
				returnFirstLegalGraveyardTargetToBattlefield),
				TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn())),
		},
	})
}
