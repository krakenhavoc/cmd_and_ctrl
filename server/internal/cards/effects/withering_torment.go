package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Withering Torment — Instant for {2}{B}:
//
//	"Destroy target creature or enchantment. You lose 2 life."
//
// Infernal Grasp one mana later with a wider target clause. In
// Commander the enchantment half is the reason to run it: mono-black
// has essentially no other instant-speed answer to a Rhystic Study
// or a Smothering Tithe.
//
// "Target creature or enchantment" is TargetPermanent narrowed by
// Or(Creature(), Enchantment()) — the predicates read the
// post-layer type, so an enchantment creature is one legal target
// rather than two, and a creature that became an enchantment is
// still hit.
//
// Life LOSS, not damage, for the same reason as Infernal Grasp: the
// card says "you lose 2 life", so nothing that watches damage sees
// it.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "ffce81c5-1b58-4882-a4e7-6f8d7cb170de",
		Name:     "Withering Torment",
		Targets: TargetPermanent("target creature or enchantment",
			Or(Creature(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return GainLife{Player: item.Controller, Amount: -2}.Apply(ctx)
		},
	})
}
