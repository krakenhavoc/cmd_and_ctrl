package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Imperious Inkmage — Creature — Orc Warlock {1}{W}{B}, 3/3 (Secrets of
// Strixhaven, #4456):
//
//	"Vigilance
//	 When this creature enters, surveil 2."
//
// Vigilance is a printed keyword; the enters trigger is Surveil (A.I.M.
// Synthoids' shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fcd31653-4964-46a6-820c-f17e8cbaa1fa",
		Name:            "Imperious Inkmage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Imperious Inkmage — surveil 2", effects.Do(effects.Surveil{N: 2})),
		},
	})
}
