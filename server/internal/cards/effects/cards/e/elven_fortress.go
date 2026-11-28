package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elven Fortress — Enchantment {G} (Fallen Empires, S82 OLD b854, #5637):
//
//	"{1}{G}: Target blocking creature gets +0/+1 until end of turn."
//
// The target is any blocking creature, whoever controls it, judged again as the
// ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29f34e18-2102-4ded-912f-4ac3e9e6744c",
		Name:         "Elven Fortress",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Target blocking creature gets +0/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 1)),
			Cost:    effects.ManaCost("{1}{G}"),
			Targets: effects.TargetCreature("target blocking creature", effects.BlockingCreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.BoostUntilEOT{Target: t.ID, Toughness: 1, Label: "Elven Fortress — +0/+1 until end of turn"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
