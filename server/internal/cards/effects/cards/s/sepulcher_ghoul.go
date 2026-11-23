package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sepulcher Ghoul — Creature — Zombie {1}{B}, 2/1 (Adventures in the
// Forgotten Realms):
//
//	"Sacrifice another creature: This creature gets +2/+2 until end of
//	 turn. Activate only once each turn."
//
// "Another" is object identity (SacrificeAnotherN). The sacrifice is a
// cost, paid at the announce, so a creature that dies for it still pays
// even if the Ghoul is removed in response. "Once each turn" gates the
// offer (CR 602.5b).
//
// No simplification.
func init() {
	const label = "Sacrifice another creature: This creature gets +2/+2 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "04936cc9-c915-4182-b84a-3491d26be548",
		Name:         "Sepulcher Ghoul",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerPump},
			Cost:      effects.SacrificeAnotherN(1, "another creature", effects.Creature()),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 2, "Sepulcher Ghoul — +2/+2"),
		}},
	})
}
