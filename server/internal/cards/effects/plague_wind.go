package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plague Wind — Sorcery {7}{B}{B} (Prophecy, #5617):
//
//	"Destroy all creatures you don't control. They can't be
//	 regenerated."
//
// In Garruk's Wake's one-sided sweep without the planeswalkers, and
// Damnation's "can't be regenerated" (CR 701.19c). Nothing is
// targeted, and control is read as the spell resolves: a creature you
// gave away in response is destroyed, one you took is spared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "18ec721f-c1ac-4581-a61d-2f0b09d6bf92",
		Name:         "Plague Wind",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, OpponentsOnly: true}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{
				Match:             And(OpponentControls(), Creature()),
				CantBeRegenerated: true,
			}.Apply(ctx)
		},
	})
}
