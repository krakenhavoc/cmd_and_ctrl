package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cunning Breezedancer — Creature — Dragon {4}{W}{U}, 4/4 (Dragons of Tarkir,
// #5551):
//
//	"Flying
//	 Whenever you cast a noncreature spell, this creature gets +2/+2 until end of turn."
//
// Frenzied Devils' cast trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ba997e44-fc80-4463-ad3f-6497be26df0f",
		Name:            "Cunning Breezedancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Cunning Breezedancer — gets +2/+2 until end of turn",
				thisGetsUntilEndOfTurn(2, 2, "Cunning Breezedancer — +2/+2")),
		},
	})
}
