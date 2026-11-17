package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hexgold Slash — Instant {R} (Phyrexia: All Will Be One, #5479):
//
//	"Hexgold Slash deals 2 damage to target creature. If that creature
//	 has toxic, Hexgold Slash deals 4 damage to that creature instead."
//
// "Has toxic" is read as the spell resolves, through the layers
// (game.ToxicTotal, CR 702.164b), so a toxic granted until end of turn
// counts. A target that is gone counters the spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "573a3afb-3a2a-41b7-ac1a-9685b90bffa3",
		Name:         "Hexgold Slash",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			amount := 2
			if s82OneB05HasToxic(ctx.Game, t.ID) {
				amount = 4
			}
			return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: amount}.Apply(ctx)
		},
	})
}
