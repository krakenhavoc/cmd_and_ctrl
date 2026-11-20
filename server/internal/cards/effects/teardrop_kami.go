package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teardrop Kami — Creature — Spirit {U}, 1/1 (Betrayers of Kamigawa,
// #5602):
//
//	"Sacrifice this creature: You may tap or untap target creature."
//
// Trickster Mage's body (nemMayTapOrUntapTheTarget) behind a sacrifice
// cost: as the ability resolves its controller picks tap it, untap it or
// leave it. A target that has left makes the ability do nothing (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f47a67e-1fcd-4f3f-81d0-1a6868606390",
		Name:         "Teardrop Kami",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: You may tap or untap target creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    SacrificeThis(),
			Targets: TargetCreature("target creature"),
			Effect:  nemMayTapOrUntapTheTarget("Teardrop Kami"),
		}},
	})
}
