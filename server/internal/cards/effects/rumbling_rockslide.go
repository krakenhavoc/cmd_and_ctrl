package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rumbling Rockslide — Sorcery {3}{R} (The Lost Caverns of Ixalan,
// #4470):
//
//	"Rumbling Rockslide deals damage to target creature equal to the
//	 number of lands you control."
//
// Bumi Bash's first mode on its own: the lands are counted as the spell
// resolves, so a land creature counts and an opponent's land does not.
// The damage is dealt by the spell. A target that is gone or no longer a
// creature makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0909cbce-b4a0-42bb-8a97-47eeb99606c0",
		Name:         "Rumbling Rockslide",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
			lands := countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsLand() })
			return DealDamage{Source: ctx.Source(), Target: id, Amount: lands}.Apply(ctx)
		})),
	})
}
