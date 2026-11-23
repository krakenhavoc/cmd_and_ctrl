package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amphin Mutineer — Creature — Salamander Pirate {3}{U}, 3/3 (Tarkir:
// Dragonstorm Commander):
//
//	"When this creature enters, exile up to one target non-Salamander
//	 creature. That creature's controller creates a 4/3 blue Salamander
//	 Warrior creature token.
//	 Encore {4}{U}{U} (…)"
//
// Crib Swap's shape on an enters trigger: the token goes to the exiled
// creature's controller, read before the exile. With no target, or one
// that is illegal as the trigger resolves, nothing is exiled and no token
// is made (the printed ruling). Encore is the shared keyword
// (encore.go, CR 702.141a); each token copy has the enters trigger too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8910f79a-66c9-41b3-b55a-e3e651ae6f26",
		Name:         "Amphin Mutineer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Amphin Mutineer — exile up to one target non-Salamander creature; its controller creates a 4/3 blue Salamander Warrior creature token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						controller, ok := effects.ControllerOfTarget(ctx, t.ID)
						if !ok {
							return nil
						}
						if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
						return effects.CreateToken{Controller: controller, Template: effects.TokenCard("4/3 blue Salamander Warrior"), N: 1}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetCreature("up to one target non-Salamander creature", effects.Not(effects.HasSubtype("Salamander"))).WithCount(0, 1)),
		},
		Activated: []effects.ActivatedAbility{effects.Encore("{4}{U}{U}")},
	})
}
