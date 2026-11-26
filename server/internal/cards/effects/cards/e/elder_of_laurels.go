package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elder of Laurels — {2}{G} Creature — Human Advisor 2/3 (Innistrad, S82 OLD
// b446, #5572):
//
//	"{3}{G}: Target creature gets +X/+X until end of turn, where X is the
//	 number of creatures you control."
//
// X is counted as the ability resolves (CR 608.2h), by the ability's
// controller, through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0431b67e-ce09-4009-ba5b-9fd41e6f1df4",
		Name:         "Elder of Laurels",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}: Target creature gets +X/+X until end of turn, where X is the number of creatures you control.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{3}{G}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					x := effects.B04CreaturesControlled(g, item.Controller)
					return effects.BoostUntilEOT{Target: t.ID, Power: x, Toughness: x, Label: "Elder of Laurels — +X/+X"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
