package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stronghold Assassin — Creature — Phyrexian Zombie Assassin {1}{B}{B},
// 2/1 (Stronghold, #5627):
//
//	"{T}, Sacrifice a creature: Destroy target nonblack creature."
//
// Both costs are paid as it is activated (CR 601.2h, CR 602.2): the tap
// needs an Assassin that has been under its controller's control since
// the turn began (CR 302.6), and the sacrifice may be the Assassin
// itself. Doom Blade's "target nonblack creature" clause and the shared
// destroyFirstLegalCardTarget body, which skips a target gone in response
// (CR 608.2b); the sacrifice stays paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "35826db0-b6ff-47b5-be16-c5b878d9e842",
		Name:         "Stronghold Assassin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice a creature: Destroy target nonblack creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
			Cost:    Plus(TapCost(), SacrificeACreature()),
			Targets: TargetCreature("target nonblack creature", NonBlack()),
			Effect:  destroyFirstLegalCardTarget,
		}},
	})
}
