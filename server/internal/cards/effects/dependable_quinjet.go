package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dependable Quinjet — {3} Artifact — Vehicle 3/3 (Marvel Super Heroes,
// #4455):
//
//	"Flying
//	 {T}: Add one mana of any color.
//	 Crew 4"
//
// The mana ability is a plain tap ability of the artifact, so it works
// while the Vehicle is not a creature (it has no summoning sickness to
// wait out then, CR 302.6 applies only to creatures). Crew is the
// vehicle family's CrewCost / CrewEffect (CR 702.122).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "60105eb8-bdca-44b9-a970-68a7615526e8",
		Name:            "Dependable Quinjet",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "Crew 4",
			Cost:    CrewCost(4),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Dependable Quinjet"),
		}},
	})
}
