package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Narstad Scrapper — {5} Artifact Creature — Construct 3/3: "{2}: This creature
// gets +1/+0 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16acd858-dba6-401e-bbfb-cabdd6d0120e",
		Name:         "Narstad Scrapper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Narstad Scrapper — +1/+0"),
		}},
	})
}
