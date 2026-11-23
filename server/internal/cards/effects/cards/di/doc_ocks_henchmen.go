package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "ae91c16c-5267-4957-b36c-5aa787fa6b5e",
		Name:            "Doc Ock's Henchmen",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ThisAttacked,
				"Doc Ock's Henchmen — it connives",
				effects.Do(effects.Instruction(effects.SourceConnives))),
		},
	})
}
