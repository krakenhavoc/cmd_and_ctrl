package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// S.H.I.E.L.D. Flying Car — Artifact — Vehicle {2}{U}, 3/3 (Marvel Super
// Heroes, #4455):
//
//	"Flash
//	 Flying
//	 When this Vehicle enters, exile up to one target creature you control.
//	 Return that card to the battlefield under its owner's control at the
//	 beginning of the next end step.
//	 Crew 1"
//
// Charming Prince's exile-and-return: the exile reports what actually left
// (a target that is gone is simply not scheduled), and the delayed trigger
// returns it under its owner's control at the next end step (CR 603.7).
// Crew is the shared Vehicle ability.
//
// No simplification.
func init() {
	const label = "S.H.I.E.L.D. Flying Car — exile up to one target creature you control, then return it at the beginning of the next end step"
	Register(Spec{
		OracleID:        "12d51534-a98d-487b-9ba9-c10861dff992",
		Name:            "S.H.I.E.L.D. Flying Car",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
					return exileTargetsThenScheduleReturn(NewContext(g, item), label)
				}),
				TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1)),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("S.H.I.E.L.D. Flying Car"),
		}},
	})
}
