package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flame Spill — Instant {2}{R} (Ikoria: Lair of Behemoths, #5512):
//
//	"Flame Spill deals 4 damage to target creature. Excess damage is dealt to
//	 that creature's controller instead."
//
// The split is made before the damage is dealt (CR 120.4a): what the creature
// still needs to die (toughness less damage already marked) goes to it and the
// rest to its controller, as one damage event (Gandalf's Sanction's shape). A
// creature whose toughness can't be read takes all of it. A creature that has
// gone leaves the spell without a legal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de6ebf55-f1df-4c3c-bba5-8655b7daf488",
		Name:         "Flame Spill",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB116DamageToTargetCreatureExcessToItsController(item, ctx, 4)
		},
	})
}
