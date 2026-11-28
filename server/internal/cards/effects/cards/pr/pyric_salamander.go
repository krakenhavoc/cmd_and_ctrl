package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyric Salamander — Creature — Salamander {1}{R}, 1/1 (Mirage, #5633):
//
//	"{R}: This creature gets +1/+0 until end of turn. Sacrifice this creature at
//	 the beginning of the next end step."
//
// Each activation schedules its own delayed sacrifice (CR 603.7); the Salamander
// goes at the next end step whether it was pumped once or several times, and a
// second trigger finds it gone.
//
// No simplification.
func init() {
	const label = "Pyric Salamander — sacrifice it at the beginning of the next end step"
	pump := effects.ThisGetsUntilEndOfTurn(1, 0, "Pyric Salamander — +1/+0 until end of turn")
	sacrifice := effects.OldB507SacrificeThisAtTheNextEndStep(label)
	effects.Register(effects.Spec{
		OracleID:     "745a2815-8487-4b78-93b7-d1d4ef6adc8f",
		Name:         "Pyric Salamander",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn. Sacrifice this creature at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := pump(g, item); err != nil {
					return err
				}
				return sacrifice(g, item)
			},
		}},
	})
}
