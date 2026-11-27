package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flamekin Brawler — Creature — Elemental Warrior {R}, 0/2 (Lorwyn, S82 OLD
// b545, #5593):
//
//	"{R}: This creature gets +1/+0 until end of turn."
//
// Firebreathing through the shared ThisGetsUntilEndOfTurn body: no {T}, so
// it can be activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e3dd85c-ba19-4d22-9f90-252676f3cffa",
		Name:         "Flamekin Brawler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Flamekin Brawler — +1/+0 until end of turn"),
		}},
	})
}
