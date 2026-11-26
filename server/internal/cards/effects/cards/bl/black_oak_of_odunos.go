package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Black Oak of Odunos — Creature — Zombie Treefolk {2}{B}, 0/5 (Born of the
// Gods, S82 OLD b378, #5559):
//
//	"Defender
//	 {B}, Tap another untapped creature you control: This creature gets +1/+1
//	 until end of turn."
//
// Wanderbrine Trapper's cost shape: the second tap is not the {T} symbol
// (CR 302.6), so a creature that arrived this turn may pay it, and it does not
// target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5fc888d1-241f-45e1-9c20-3c5ae0aadc30",
		Name:            "Black Oak of Odunos",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label: "{B}, Tap another untapped creature you control: This creature gets +1/+1 until end of turn.",
			Cost: effects.Plus(effects.ManaCost("{B}"),
				effects.TapAnotherUntapped("another untapped creature you control", effects.Creature())),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Black Oak of Odunos — +1/+1"),
		}},
	})
}
