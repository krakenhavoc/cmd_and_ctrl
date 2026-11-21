package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "aa7b41ba-08eb-40b8-a8d5-e67dc09c88ba",
		Name:         "You Find a Cursed Idol",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Smash It — Destroy target artifact.", effects.TargetPermanent("target artifact", effects.Artifact())),
			effects.Mode("Lift the Curse — Destroy target enchantment.", effects.TargetPermanent("target enchantment", effects.Enchantment())),
			effects.Mode("Steal Its Eyes — Create a Treasure token and venture into the dungeon."),
		),
		// The venture waits for a prompt the Treasure's creation asks (a CR 616
		// order; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(0) && !ctx.HasMode(1) {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(2) {
					return nil
				}
				return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(2) {
					return nil
				}
				return effects.VentureIntoTheDungeon{}.Apply(ctx)
			}),
		),
	})
}
