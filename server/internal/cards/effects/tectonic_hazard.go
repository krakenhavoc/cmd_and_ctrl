package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tectonic Hazard — Sorcery {R} (EDHREC rank 2637):
//
//	"Tectonic Hazard deals 1 damage to each opponent and each creature
//	 they control."
//
// A one-sided Electrickery for the table. One damage to every
// opponent, then one to every creature an opponent controls
// (damageEachMatching — the set is snapshotted before the first
// point lands, so nothing is hit twice and a creature that took
// lethal dies at the state-based sweep). The caster and their
// creatures are untouched, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9097243a-39fa-4e18-8316-c0e57699c783",
		Name:         "Tectonic Hazard",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, opp := range ctx.Opponents() {
				if err := (DealDamage{Source: ctx.Source(), Target: opp, Amount: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return damageEachMatching(ctx, And(Creature(), OpponentControls()), 1)
		},
	})
}
