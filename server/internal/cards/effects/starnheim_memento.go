package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starnheim Memento — Artifact {3} (Foundations Jumpstart):
//
//	"{T}: Add {W}.
//	 {1}{W}, {T}: Target creature gets +1/+1 and gains flying until end of
//	 turn. Activate only as a sorcery."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "887de1fc-8fbd-472a-bcf2-4f36ea9376d1",
		Name:         "Starnheim Memento",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{1}{W}, {T}: Target creature gets +1/+1 and gains flying until end of turn. Activate only as a sorcery.",
			Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1, "flying")),
			Cost:         Plus(ManaCost("{1}{W}"), TapCost()),
			Targets:      TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := boostTargetUntilEOT("Starnheim Memento — +1/+1", 1, 1)(g, item); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}}.Apply(ctx)
			},
		}},
	})
}
