package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Insist — Sorcery {G}:
//
//	"The next creature spell you cast this turn can't be countered.
//	 Draw a card."
//
// A one-use promise (ADR 0106 §4 decision 3, #1806): the grant waits on
// the caster, and the first creature spell they cast this turn spends
// it as that spell becomes cast (CR 601.2i), whether or not anybody was
// going to counter it, and carries a "can't be countered" mark for as
// long as it is on the stack. A noncreature spell leaves the promise
// alone; a copy of a creature spell is not cast (CR 707.10) and spends
// nothing; an unspent promise ends at cleanup (CR 514.2). The table
// sees an unspent promise on the caster's panel.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "518684d9-467a-4936-a3d9-17d23db4622c",
		Name:         "Insist",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (GrantCounterShield{From: "Insist", Grant: NextSpellYouCastCantBeCountered(
				"The next creature spell you cast this turn can't be countered.",
				game.PermissionFilter{CreatureOnly: true})}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
