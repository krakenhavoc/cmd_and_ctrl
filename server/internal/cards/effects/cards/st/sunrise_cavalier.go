package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunrise Cavalier — {1}{R}{W} Creature — Human Knight 3/3 (#2561, ADR
// 0132):
//
//	"Trample, haste
//	 If it's neither day nor night, it becomes day as this creature enters.
//	 Whenever day becomes night or night becomes day, put a +1/+1 counter
//	 on target creature you control."
//
// The trigger is targeted, so it is chosen as the ability goes on the
// stack and re-checked as it resolves (CR 603.3d, 608.2b): with no
// creature to choose it is removed without a prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0718ea67-ced3-46bc-9164-8841e6bd73ae",
		Name:            "Sunrise Cavalier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		AsEnters:        effects.BecomesDayAsEnters(),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverDayBecomesNightOrNightBecomesDay("Sunrise Cavalier — +1/+1 counter on target creature you control",
					effects.PlusOneCounterOnChosenTargets),
				effects.TargetCreature("target creature you control", effects.YouControl()),
			),
		},
	})
}
