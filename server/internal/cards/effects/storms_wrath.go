package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Storm's Wrath — Sorcery {2}{R}{R} (Tarkir: Dragonstorm Commander):
//
//	"Storm's Wrath deals 4 damage to each creature and each
//	 planeswalker."
//
// Brotherhood's End's damage mode as the whole card: one damage instance
// to every creature and planeswalker on the battlefield, the caster's
// included, so a permanent that is both is dealt 4 once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eeb6f5eb-0e16-468e-9dfe-6ec8d8172f86",
		Name:         "Storm's Wrath",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, Or(Creature(), Planeswalker()), 4)
		},
	})
}
