package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Endurance Bobblehead — Artifact — Bobblehead {3} (Fallout, #5948):
//
//	"{T}: Add one mana of any color.
//	 {3}, {T}: Up to X target creatures you control get +1/+0 and gain
//	 indestructible until end of turn, where X is the number of Bobbleheads
//	 you control as you activate this ability. Activate only as a sorcery."
//
// Agility Bobblehead's "up to X" (TargetSpec.MaxFrom, read once as the
// ability is activated) with a +1/+0 and indestructible grant for each
// target still legal at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a69116b-3d2e-4eac-9d02-902373de3cb7",
		Name:         "Endurance Bobblehead",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}, {T}: Up to X target creatures you control get +1/+0 and gain indestructible until end of turn, where X is the number of Bobbleheads you control as you activate this ability. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "indestructible")),
			Cost:         effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets:      effects.S82GapB60UpToXCreatures(),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Label: "Endurance Bobblehead — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"}, Label: "Endurance Bobblehead — indestructible"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
