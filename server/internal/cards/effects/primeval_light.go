package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "b416d7ac-ec56-40e2-ba3d-bcecd1b73175",
		Name:         "Primeval Light",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Sweep:   game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy, OpponentsOnly: true},
			Answers: game.AnswerRemove,
		},
		Targets: TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := TargetedPlayer(ctx)
			if player == uuid.Nil {
				return nil
			}
			return DestroyAllMatching{Match: And(Enchantment(), ControlledBy(player))}.Apply(ctx)
		},
	})
}
