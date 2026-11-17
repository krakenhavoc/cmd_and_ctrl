package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tidal Bore — Instant {1}{U} (Mercadian Masques, #5619):
//
//	"You may return an Island you control to its owner's hand rather
//	 than pay this spell's mana cost.
//	 You may tap or untap target creature."
//
// Daze's alternative cost (ReturnInstead): the Island goes back to its
// owner's hand as a cost, with Tidal Bore already on the stack (CR
// 601.2h). The body is Rejoinder's "you may tap or untap target
// creature" (socB23MayTapOrUntap): the caster picks tap it, untap it or
// leave it as Tidal Bore resolves. A target that has become illegal
// means the spell doesn't resolve (CR 608.2b) and nothing is asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "513ad32a-2887-4825-adad-86a87cf1d2cb",
		Name:         "Tidal Bore",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		AlternativeCosts: []game.AlternativeCost{
			ReturnInstead(
				"Return an Island you control to its owner's hand",
				PermanentYouControl("an Island you control", HasSubtype("Island")),
				"an Island you control",
			),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return socB23MayTapOrUntap(ctx, "Tidal Bore", t.ID)
				}
			}
			return nil
		},
	})
}
