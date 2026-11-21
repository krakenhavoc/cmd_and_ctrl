package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Ghoul — Creature — Phyrexian Zombie {2}{B}, 2/2:
//
//	"Sacrifice a creature: This creature gets +2/+2 until end of turn."
//
// Fallen Ideal's sacrifice outlet and pump. The cost may sacrifice the
// Ghoul itself, in which case the pump has nothing to land on.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "748c2723-dc59-4aa7-9667-d4679448c8b6",
		Name:         "Phyrexian Ghoul",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    effects.SacrificeACreature(),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Phyrexian Ghoul — +2/+2 until end of turn"),
		}},
	})
}
