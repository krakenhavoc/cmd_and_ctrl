package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Detention Chariot — Artifact — Vehicle {4}{W}{W}, 6/6 (Aetherdrift,
// #4464):
//
//	"When this Vehicle enters, exile target artifact or creature an
//	 opponent controls until this Vehicle leaves the battlefield.
//	 Crew 3 (Tap any number of creatures you control with total power 3
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)
//	 Cycling {W} ({W}, Discard this card: Draw a card.)"
//
// Glass Casket's entry trigger (exileChosenTargetUntilThisLeaves) over an
// artifact or creature an opponent controls, re-checked as the trigger
// resolves (CR 608.2b). The card comes back under its owner's control
// when the Chariot leaves (CR 610.3c); if the Chariot has already left
// when the trigger resolves, nothing is exiled (CR 610.3b). The trigger
// watches the Vehicle entering, so it needs no crew. Crew is the
// engine's (CR 702.122a) and cycling is the hand ability (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ea2c5ad-2982-4041-8c9d-a1d31ced11aa",
		Name:         "Detention Chariot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Detention Chariot — exile target artifact or creature an opponent controls until this Vehicle leaves the battlefield",
				exileChosenTargetUntilThisLeaves("Detention Chariot — the exiled card returns when Detention Chariot leaves the battlefield")),
				TargetPermanent("target artifact or creature an opponent controls", Or(Artifact(), Creature()), OpponentControls())),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "Crew 3",
				Cost:    CrewCost(3),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Detention Chariot"),
			},
			Cycling("{W}"),
		},
	})
}
