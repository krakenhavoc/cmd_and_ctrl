package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1208cb6e-34b4-4a23-9a5b-a01229113e38",
		Name:         "Advocate of the Beast",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourEndStep("Advocate of the Beast — put a +1/+1 counter on target Beast creature you control",
				putACounterOnTheTarget(game.CounterPlusOne)),
				PermanentYouControl("target Beast creature you control", Creature(), HasSubtype("Beast"))),
		},
	})
}
