package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Doom Blade — Instant for {1}{B}:
//
//	"Destroy target nonblack creature."
//
// S20 sub-PR 1: the exit-criteria card for structured targeting.
// Targets is TargetCreature(NonBlack()) — the picker offers only
// non-black creatures, the engine rejects a black one at announce
// (CR 601.2c), and a creature that turns black in response fizzles
// the spell at resolution (CR 608.2b). Colour comes from Scryfall's
// computed colours (falls back to mana-cost symbols for fixtures).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "59e7f2ae-4535-4191-98be-3e65b6b2befa",
		Name:         "Doom Blade",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:      effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
