package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marker Beetles — Creature — Insect {1}{G}{G}, 2/3 (Urza's Destiny,
// #5621):
//
//	"When this creature dies, target creature gets +1/+1 until end of
//	 turn.
//	 {2}, Sacrifice this creature: Draw a card."
//
// Disease Carriers' dies trigger the other way, with the shared
// boostTargetUntilEOT body: any player's creature, checked again as it
// resolves (CR 608.2b), +1/+1 until the cleanup step (CR 514.2). The
// target entry declares the benefit. The draw row is Brass Secretary's:
// the sacrifice is part of the cost, so sacrificing the Beetles for a
// card also triggers the +1/+1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b1614fd-2810-4d4d-a0d6-934f27adbe65",
		Name:         "Marker Beetles",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				WhenThisDies("Marker Beetles — target creature gets +1/+1 until end of turn",
					boostTargetUntilEOT("Marker Beetles — +1/+1 until end of turn", 1, 1)),
				TargetCreature("target creature"),
			), ForTargets(BenefitUntilEOT(0, 1, 1))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Effect:  ebDrawACard,
		}},
	})
}
