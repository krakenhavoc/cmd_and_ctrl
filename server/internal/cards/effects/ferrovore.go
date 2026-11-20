package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ferrovore — Creature — Beast {2}{R}, 2/2 (Scars of Mirrodin, #5577):
//
//	"{R}, Sacrifice an artifact: This creature gets +3/+0 until end of turn."
//
// Atog's shape at {R} and +3/+0. The pump is pinned to this object and
// does nothing to one that left and came back (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f0f068a-fd4e-4a6f-9439-0e694e34a945",
		Name:         "Ferrovore",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, Sacrifice an artifact: This creature gets +3/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{R}"), SacrificeN(1, "an artifact", Artifact())),
			Effect:  thisGetsUntilEndOfTurn(3, 0, "Ferrovore — +3/+0"),
		}},
	})
}
