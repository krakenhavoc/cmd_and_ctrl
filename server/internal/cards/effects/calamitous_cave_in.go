package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Calamitous Cave-In — Sorcery {3}{R} (The Lost Caverns of Ixalan,
// #4470):
//
//	"Calamitous Cave-In deals X damage to each creature and each
//	 planeswalker, where X is the number of Caves you control plus the
//	 number of Cave cards in your graveyard."
//
// X is counted as the spell resolves (CR 608.2h, caveCount, Sinuous
// Benthisaur's count), and the damage is Brotherhood's End's sweep: every
// creature and every planeswalker on the battlefield, the caster's too,
// dealt at once. With no Caves anywhere, X is 0 and nothing is dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "558135b3-babc-40b6-9d9d-57ab0965ee3d",
		Name:     "Calamitous Cave-In",
		// ADR 0126 §6: the damage is a count read as it resolves, not a fixed number or the spell's X, so it is declared as the destruction it can be, an upper bound (Chain Reaction's declaration).
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, Or(Creature(), Planeswalker()), caveCount(ctx.Game, ctx.Controller(), uuid.Nil))
		},
	})
}
