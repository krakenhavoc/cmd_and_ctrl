package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "6f47a67e-1fcd-4f3f-81d0-1a6868606390",
		Name:         "Teardrop Kami",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: You may tap or untap target creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NemMayTapOrUntapTheTarget("Teardrop Kami"),
		}},
	})
}
