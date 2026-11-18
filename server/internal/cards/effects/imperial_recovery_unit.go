package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imperial Recovery Unit — Artifact — Vehicle {2}{W}, 3/4 (Kamigawa: Neon
// Dynasty, S82):
//
//	"Whenever this Vehicle attacks, return target creature or Vehicle card
//	 with mana value 2 or less from your graveyard to your hand.
//	 Crew 2"
//
// Mirkwood Elk's graveyard-to-hand trigger with Back on Track's creature or
// Vehicle card filter and a mana value cap. A trigger whose one target has
// gone does nothing (CR 608.2b).
//
// No simplification.
func init() {
	trigger := WheneverThisAttacks("Imperial Recovery Unit — return target creature or Vehicle card with mana value 2 or less from your graveyard to your hand",
		returnTargetGraveyardCardToHand)
	trigger.Targets = TargetCardInGraveyard("target creature or Vehicle card with mana value 2 or less in your graveyard",
		YouOwn(), Or(Creature(), HasSubtype("Vehicle")), ManaValueLE(2))
	Register(Spec{
		OracleID:     "18ad0dd1-2d0e-46b4-8ab1-ece07180267f",
		Name:         "Imperial Recovery Unit",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Imperial Recovery Unit"),
		}},
	})
}
