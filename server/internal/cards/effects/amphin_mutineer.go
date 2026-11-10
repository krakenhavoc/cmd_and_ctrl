package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Amphin Mutineer — Creature — Salamander Pirate {3}{U}, 3/3 (Tarkir:
// Dragonstorm Commander; first printed in Lost Caverns of Ixalan):
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
	Register(Spec{
		OracleID:     "8910f79a-66c9-41b3-b55a-e3e651ae6f26",
		Name:         "Amphin Mutineer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Amphin Mutineer — exile up to one target non-Salamander creature; its controller creates a 4/3 blue Salamander Warrior creature token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						controller, ok := controllerOfTarget(ctx, t.ID)
						if !ok {
							return nil
						}
						if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
						return CreateToken{Controller: controller, Template: TokenCard("4/3 blue Salamander Warrior"), N: 1}.Apply(ctx)
					}
					return nil
				}),
				TargetCreature("up to one target non-Salamander creature", Not(HasSubtype("Salamander"))).WithCount(0, 1)),
		},
		Activated: []ActivatedAbility{Encore("{4}{U}{U}")},
	})
}
