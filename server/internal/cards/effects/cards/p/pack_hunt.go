package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pack Hunt — Sorcery {3}{G} (Nemesis, S82 OLD, #5618):
//
//	"Search your library for up to three cards with the same name as target
//	 creature, reveal them, put them into your hand, then shuffle."
//
// The name is the target creature's as it resolves; a target that has become
// illegal makes the spell do nothing (CR 608.2b). A search naming a quality may
// find fewer than three or none (CR 701.23b); the library is shuffled either
// way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ee41993-9d14-43a8-bf92-433c4f3aab01",
		Name:         "Pack Hunt",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB677PackHunt(ctx.Game, item)
		},
	})
}
