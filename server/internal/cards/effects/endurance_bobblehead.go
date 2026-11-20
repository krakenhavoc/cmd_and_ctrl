package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6a69116b-3d2e-4eac-9d02-902373de3cb7",
		Name:         "Endurance Bobblehead",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{3}, {T}: Up to X target creatures you control get +1/+0 and gain indestructible until end of turn, where X is the number of Bobbleheads you control as you activate this ability. Activate only as a sorcery.",
			Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0, "indestructible")),
			Cost:         Plus(ManaCost("{3}"), TapCost()),
			Targets:      s82GapB60UpToXCreatures(),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (BoostUntilEOT{Target: t.ID, Power: 1, Label: "Endurance Bobblehead — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"}, Label: "Endurance Bobblehead — indestructible"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
