package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Facevaulter — Creature — Goblin Warrior {B}, 1/1 (Lorwyn, S82 OLD b545,
// #5593):
//
//	"{B}, Sacrifice a Goblin: This creature gets +2/+2 until end of turn."
//
// Weirding Shaman's cost: the Goblin is a cost, so the Facevaulter may
// sacrifice itself, and then the boost finds nothing to boost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "602b30a0-ecf3-4639-a512-ee020675e432",
		Name:         "Facevaulter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice a Goblin: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.B25SacrificeAGoblin()),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Facevaulter — +2/+2 until end of turn"),
		}},
	})
}
