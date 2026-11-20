package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Volcanic Fallout — Instant {1}{R}{R} (Conflux, #5587):
//
//	"This spell can't be countered.
//	 Volcanic Fallout deals 2 damage to each creature and each player."
//
// Evincar's Justice's sweep through the shared
// b23DamageEachCreatureAndEachPlayer body: every creature and every
// player, the caster included, as one damage instance. "Can't be
// countered" is Spec.CantBeCountered, and a "can't" beats a "can" (CR
// 101.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bd6ce147-c991-40ac-b276-65b630fad8b6",
		Name:            "Volcanic Fallout",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}},
		CantBeCountered: true,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b23DamageEachCreatureAndEachPlayer(ctx, 2)
		},
	})
}
