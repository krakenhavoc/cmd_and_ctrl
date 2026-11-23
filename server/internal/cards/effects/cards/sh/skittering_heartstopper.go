package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skittering Heartstopper — Creature — Insect {B}, 1/2 (Ixalan, #5532):
//
//	"{B}: This creature gains deathtouch until end of turn."
//
// Twilight Panther's ability: the deathtouch lasts until the cleanup
// step (CR 514.2), and the ability can be activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d07f497e-4cf8-4e12-b17e-e3d58ff9df47",
		Name:         "Skittering Heartstopper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{B}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("deathtouch", "Skittering Heartstopper — deathtouch until end of turn"),
		}},
	})
}
