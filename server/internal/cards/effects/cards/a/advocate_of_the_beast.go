package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Advocate of the Beast — Creature — Elf Shaman {2}{G}, 2/3 (Foundations
// Jumpstart):
//
//	"At the beginning of your end step, put a +1/+1 counter on target Beast
//	 creature you control."
//
// An end-step trigger that targets as it goes on the stack; a target gone by
// resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1208cb6e-34b4-4a23-9a5b-a01229113e38",
		Name:         "Advocate of the Beast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourEndStep("Advocate of the Beast — put a +1/+1 counter on target Beast creature you control",
				effects.PutACounterOnTheTarget(game.CounterPlusOne)),
				effects.PermanentYouControl("target Beast creature you control", effects.Creature(), effects.HasSubtype("Beast"))),
		},
	})
}
