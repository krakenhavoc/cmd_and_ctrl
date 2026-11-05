package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Planetary Bombardment — Sorcery {R} (Star Trek, #4973):
//
//	"Planetary Bombardment deals 2 damage to target creature. Planetary
//	 Bombardment deals 4 damage to that creature instead if a creature
//	 died this turn."
//
// The condition is read at resolution off the table-wide per-turn tally
// of creatures that died (b11CreaturesDiedThisTurn): any creature, any
// controller, tokens included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c2fdeff-a0c6-44cb-9603-4a9865f62985",
		Name:         "Planetary Bombardment",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			amount := 2
			if b11CreaturesDiedThisTurn(ctx.Game) > 0 {
				amount = 4
			}
			return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: amount}.Apply(ctx)
		},
	})
}
