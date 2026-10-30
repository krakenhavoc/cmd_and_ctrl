package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mjölnir's Might — Sorcery {3}{R} (Marvel Super Heroes Commander,
// #4455):
//
//	"Mjölnir's Might deals 4 damage to target player.
//	 Exile the top card of your library. Until the end of your next turn,
//	 you may play that card."
//
// Two instructions in printed order. The damage goes to the chosen
// player; the impulse exile is the caster's own top card with the
// shared window of "until the end of your next turn"
// (ExileTopNUntilYourNextTurn). If the target is illegal on resolution
// the whole spell does nothing (CR 608.2b), exile included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c944733-3fb5-4d1f-af1a-3fd31377eb9a",
		Name:         "Mjölnir's Might",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 4}.Apply(ctx)
					}
				}
				return nil
			}),
			Instruction(func(ctx *Context) error { return ExileTopNUntilYourNextTurn(ctx, 1) }),
		),
	})
}
