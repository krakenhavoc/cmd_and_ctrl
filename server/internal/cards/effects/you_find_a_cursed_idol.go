package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// You Find a Cursed Idol — Sorcery {1}{G}:
//
//	"Choose one —
//	 • Smash It — Destroy target artifact.
//	 • Lift the Curse — Destroy target enchantment.
//	 • Steal Its Eyes — Create a Treasure token and venture into the
//	   dungeon."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa7b41ba-08eb-40b8-a8d5-e67dc09c88ba",
		Name:         "You Find a Cursed Idol",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Smash It — Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
			Mode("Lift the Curse — Destroy target enchantment.", TargetPermanent("target enchantment", Enchantment())),
			Mode("Steal Its Eyes — Create a Treasure token and venture into the dungeon."),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) || ctx.HasMode(1) {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
				}
			}
			if ctx.HasMode(2) {
				if err := (CreateToken{Template: TreasureToken(), N: 1}).Apply(ctx); err != nil {
					return err
				}
				return VentureIntoTheDungeon{}.Apply(ctx)
			}
			return nil
		},
	})
}
