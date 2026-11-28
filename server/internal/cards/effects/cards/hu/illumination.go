package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Illumination — Instant {W}{W} (Mirage, S82 OLD batch 812, #5633):
//
//	"Counter target artifact or enchantment spell. Its controller gains life
//	 equal to its mana value."
//
// Psychic Barrier's counter-then-controller body with a gain: the spell's
// controller and mana value are read before it is countered (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd1a6508-aaba-4515-9496-057915d02040",
		Name:         "Illumination",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target artifact or enchantment spell", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve:    effects.S82OldB812CounterThenItsControllerGainsLifeEqualToItsManaValue,
	})
}
