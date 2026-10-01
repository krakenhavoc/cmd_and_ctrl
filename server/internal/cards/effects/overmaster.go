package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overmaster — Sorcery {R}:
//
//	"The next instant or sorcery spell you cast this turn can't be
//	 countered.
//	 Draw a card."
//
// Insist's shape for instants and sorceries (ADR 0106 §4 decision 3,
// #1806): a one-use promise the caster's first instant or sorcery spell
// this turn spends as it becomes cast (CR 601.2i), becoming a "can't be
// countered" mark on that spell. A creature spell leaves it alone; a
// copy is not cast (CR 707.10); an unspent promise ends at cleanup
// (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "acc842a4-26e8-435a-b717-c1043b581ff8",
		Name:         "Overmaster",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (GrantCounterShield{From: "Overmaster", Grant: NextSpellYouCastCantBeCountered(
				"The next instant or sorcery spell you cast this turn can't be countered.",
				game.PermissionFilter{InstantOrSorceryOnly: true})}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
