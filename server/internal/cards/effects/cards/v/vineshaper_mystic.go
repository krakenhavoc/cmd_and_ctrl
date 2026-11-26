package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vineshaper Mystic — Creature — Merfolk Shaman {2}{G}, 1/3 (Ixalan,
// #5532):
//
//	"When this creature enters, put a +1/+1 counter on each of up to two
//	 target Merfolk you control."
//
// Up to two targets chosen as the trigger goes on the stack (CR 603.3d),
// the Mystic itself legal among them; each one still a legal Merfolk you
// control as the trigger resolves gets a counter (CR 608.2b), and none
// chosen is fine.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca514283-c58b-4108-82b2-3b6b15b467f4",
		Name:         "Vineshaper Mystic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Vineshaper Mystic — put a +1/+1 counter on each of up to two target Merfolk you control", effects.PlusOneCounterOnChosenTargets),
				effects.TargetCreature("up to two target Merfolk you control", effects.Subtype("Merfolk"), effects.YouControl()).WithCount(0, 2)),
		},
	})
}
