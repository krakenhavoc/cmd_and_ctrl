package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dispersal Shield — Instant {1}{U} (Scourge):
//
//	"Counter target spell if its mana value is less than or equal to the
//	 greatest mana value among permanents you control."
//
// Any spell is a legal target; the comparison is made as Dispersal Shield
// resolves, with the target's mana value as it sits on the stack ({X}
// counts as the value chosen, CR 202.3e) against scgB1GreatestManaValue.
// When the target's mana value is greater, the spell resolves and does
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aefb1463-76fb-419e-8cda-8c5bfb88a3d2",
		Name:         "Dispersal Shield",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 || spellManaValueForEffect(ctx.Game, targets[0].ID) > scgB1GreatestManaValue(ctx) {
				return nil
			}
			return CounterTarget{StackID: targets[0].ID}.Apply(ctx)
		},
	})
}
