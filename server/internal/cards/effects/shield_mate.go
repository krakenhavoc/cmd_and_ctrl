package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shield Mate — Creature — Human Soldier {W}, 1/1 (Exodus, #5626):
//
//	"Sacrifice this creature: Target creature gets +0/+4 until end of
//	 turn."
//
// A sacrifice-self cost (CR 602.2b, so it can be activated the turn it
// arrives and while summoning sick) on the shared boostTargetUntilEOT
// body. The target is re-checked at resolution (CR 608.2b); one that
// left in response gets nothing. Shield Mate may target itself, but it
// is already gone by then, so that does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "98e7bfd4-fcb6-47b1-ad0b-c54b85dc2a79",
		Name:         "Shield Mate",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Target creature gets +0/+4 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(BenefitUntilEOT(0, 0, 4))},
			Cost:    SacrificeThis(),
			Targets: TargetCreature("target creature"),
			Effect:  boostTargetUntilEOT("Shield Mate — +0/+4", 0, 4),
		}},
	})
}
