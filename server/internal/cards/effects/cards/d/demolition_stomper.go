package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demolition Stomper — Artifact — Vehicle {6}, 10/7 (Kaladesh, #5539):
//
//	"This Vehicle can't be blocked by creatures with power 2 or less.
//	 Crew 5 (Tap any number of creatures you control with total power 5
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// Air Response Unit's crew (vehicles.go, CR 702.122); the block restriction
// is Enraged Ceratok's, judged as blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "282e5e5f-365a-4e11-aeac-ead71d1cf68a",
		Name:         "Demolition Stomper",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 5",
			Cost:    effects.CrewCost(5),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Demolition Stomper"),
		}},
	})
}
