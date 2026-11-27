package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Minotaur Illusionist — Creature — Minotaur Wizard {3}{U}{R}, 3/4
// (Apocalypse, #5614):
//
//	"{1}{U}: This creature gains shroud until end of turn.
//	 {R}, Sacrifice this creature: It deals damage equal to its power to
//	 target creature."
//
// Frost Raptor's shroud and Heartfire Immolator's sacrifice, which reads the
// power as the creature last existed (CR 113.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee5fc6e9-0c10-4015-8853-dd1f33dcc27e",
		Name:         "Minotaur Illusionist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{U}: This creature gains shroud until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    effects.ManaCost("{1}{U}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("shroud", "Minotaur Illusionist — shroud"),
			},
			{
				Label:   "{R}, Sacrifice this creature: It deals damage equal to its power to target creature.",
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.SacrificedSourceDealsItsPowerToTheTarget,
			},
		},
	})
}
