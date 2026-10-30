package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doc Ock's Henchmen — Creature — Human Villain {2}{U}, 2/1 (Marvel Super
// Heroes Commander, #4455):
//
//	"Flash
//	 Whenever this creature attacks, it connives. (Draw a card, then
//	 discard a card. If you discarded a nonland card, put a +1/+1 counter
//	 on this creature.)"
//
// Flash is printed. The attack trigger is the ThisAttacked shape with the
// shared sourceConnives body (A.I.M. Scientists, Atlantean Skirmisher): the
// counter goes on the Henchmen only if they are still there (CR 701.50c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ae91c16c-5267-4957-b36c-5aa787fa6b5e",
		Name:            "Doc Ock's Henchmen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttacked,
				"Doc Ock's Henchmen — it connives",
				Do(Instruction(sourceConnives))),
		},
	})
}
