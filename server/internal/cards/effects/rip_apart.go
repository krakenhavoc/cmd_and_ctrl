package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rip Apart — Sorcery {R}{W} (EDHREC rank 2528):
//
//	"Choose one —
//	 • Rip Apart deals 3 damage to target creature or planeswalker.
//	 • Destroy target artifact or enchantment."
//
// Boros's two-mana answer to most things. Two targeted options on a
// "choose one", so the chosen option's clause is the whole cast's —
// Null Elemental Blast's shape. The damage is the spell's own, so a
// planeswalker loses loyalty and a creature carries the damage until
// the cleanup step; the destroy honours indestructible through the
// single-target verb.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cbdbf18f-0180-4ade-a79e-1e644dd42d6f",
		Name:         "Rip Apart",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Rip Apart deals 3 damage to target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker()))),
			Mode("Destroy target artifact or enchantment.",
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment()))),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				switch {
				case ctx.HasMode(0):
					if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
						return err
					}
				case ctx.HasMode(1):
					if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
