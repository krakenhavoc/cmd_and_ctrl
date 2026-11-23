package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "513ad32a-2887-4825-adad-86a87cf1d2cb",
		Name:         "Tidal Bore",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		AlternativeCosts: []game.AlternativeCost{
			effects.ReturnInstead(
				"Return an Island you control to its owner's hand",
				effects.PermanentYouControl("an Island you control", effects.HasSubtype("Island")),
				"an Island you control",
			),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return effects.SocB23MayTapOrUntap(ctx, "Tidal Bore", t.ID)
				}
			}
			return nil
		},
	})
}
