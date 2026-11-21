package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Primeval Light — Sorcery {3}{G} (Guildpact, #5599):
//
//	"Destroy all enchantments target player controls."
//
// Rain of Daggers' shape with enchantments and any player: Cleanfall's
// DestroyAllMatching narrowed to the target's permanents (ControlledBy).
// A target who has left the game makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b416d7ac-ec56-40e2-ba3d-bcecd1b73175",
		Name:         "Primeval Light",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Sweep:   game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy, OpponentsOnly: true},
			Answers: game.AnswerRemove,
		},
		Targets: effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := effects.TargetedPlayer(ctx)
			if player == uuid.Nil {
				return nil
			}
			return effects.DestroyAllMatching{Match: effects.And(effects.Enchantment(), effects.ControlledBy(player))}.Apply(ctx)
		},
	})
}
