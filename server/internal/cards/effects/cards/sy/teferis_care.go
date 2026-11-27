package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teferi's Care — Enchantment {2}{W} (Invasion, #5616):
//
//	"{W}, Sacrifice an enchantment: Destroy target enchantment.
//	 {3}{U}{U}: Counter target enchantment spell."
//
// Two printed abilities. The sacrifice is a cost paid at announce (Teferi's
// Care itself qualifies), so the destroy does not need what was sacrificed.
// The counter targets a spell on the stack, judged again at resolution
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c5c7b26-9cd6-4f29-bbb2-a0745153ee5d",
		Name:         "Teferi's Care",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}, Sacrifice an enchantment: Destroy target enchantment.",
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.SacrificeN(1, "an enchantment", effects.Enchantment())),
				Targets: effects.TargetPermanent("target enchantment", effects.Enchantment()),
				Effect:  effects.DestroyChosenPermanent,
			},
			{
				Label:   "{3}{U}{U}: Counter target enchantment spell.",
				Cost:    effects.ManaCost("{3}{U}{U}"),
				Targets: effects.TargetSpell("target enchantment spell", effects.Enchantment()),
				Effect:  teferisCareCounter,
			},
		},
	})
}

func teferisCareCounter(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return effects.CounterTarget{StackID: t.ID}.Apply(ctx)
		}
	}
	return nil
}
