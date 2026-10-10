package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knight Paladin — Artifact — Vehicle {5}, 6/6:
//
//	"Trample
//	 Rapid-fire Battle Cannon — When this Vehicle enters, it deals 4
//	 damage to each opponent.
//	 Crew 1 (Tap any number of creatures you control with total power 1
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// A Vehicle (CR 702.122a) with printed trample (CR 702.19a), which
// matters only while it is crewed, and an entry trigger. "Rapid-fire
// Battle Cannon" is an ability word with no rules meaning. The damage
// is dealt by the Vehicle, to each opponent at once as one damage
// event, and it needs no crew: the Vehicle deals it as an artifact.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e2ab0f84-8402-472d-be66-9661b85e08fc",
		Name:            "Knight Paladin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Knight Paladin"),
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Knight Paladin — Rapid-fire Battle Cannon: 4 damage to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					return damageToEachOpponent(g, item, 4)
				}),
		},
	})
}
