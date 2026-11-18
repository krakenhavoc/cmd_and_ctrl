package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Ghoul — Creature — Phyrexian Zombie {2}{B}, 2/2:
//
//	"Sacrifice a creature: This creature gets +2/+2 until end of turn."
//
// Fallen Ideal's sacrifice outlet and pump. The cost may sacrifice the
// Ghoul itself, in which case the pump has nothing to land on.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "748c2723-dc59-4aa7-9667-d4679448c8b6",
		Name:         "Phyrexian Ghoul",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a creature: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    SacrificeACreature(),
			Effect:  thisGetsUntilEndOfTurn(2, 2, "Phyrexian Ghoul — +2/+2 until end of turn"),
		}},
	})
}
