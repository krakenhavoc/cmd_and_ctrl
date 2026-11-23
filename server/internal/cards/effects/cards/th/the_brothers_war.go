package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Brothers' War — Enchantment — Saga {3}{R} (The Brothers' War
// Commander, #5482):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Create two tapped Powerstone tokens.
//	 II — Choose two target players. Until your next turn, each creature
//	      they control attacks the other chosen player each combat if able.
//	 III — This Saga deals X damage to any target and X damage to any
//	       other target, where X is the number of artifacts you control."
//
// I makes the tokens yours and tapped. II is a requirement on the rules,
// so it reaches creatures that arrive later (CR 611.2c): one live record
// for each chosen player over the creatures they control, "attacks the
// other chosen player if able" (CR 508.1d, the same requirement as encore's),
// lasting until your next turn; the chosen players may include you. III is
// two target clauses, the second Distinct ("any other target"), one damage
// event whose X is counted as it resolves.
func init() {
	const name = "The Brothers' War"
	effects.Register(effects.Spec{
		OracleID:     "43040c34-acad-4d62-893d-fd606fa9d8fe",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create two tapped Powerstone tokens"),
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 2})),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "choose two target players; until your next turn, each creature they control attacks the other chosen player each combat if able"),
				effects.TargetPlayer("two target players").WithCount(2, 2), effects.S82BroB20BrothersWarClash),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "X damage to any target and X damage to any other target, where X is the number of artifacts you control"),
				effects.Clauses(effects.TargetAny(), effects.Distinct(effects.TargetAny())), effects.S82BroB20BrothersWarBurst),
		},
	})
}
