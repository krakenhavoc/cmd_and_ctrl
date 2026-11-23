package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Starnheim Memento — Artifact {3} (Foundations Jumpstart):
//
//	"{T}: Add {W}.
//	 {1}{W}, {T}: Target creature gets +1/+1 and gains flying until end of
//	 turn. Activate only as a sorcery."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "887de1fc-8fbd-472a-bcf2-4f36ea9376d1",
		Name:         "Starnheim Memento",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{W}, {T}: Target creature gets +1/+1 and gains flying until end of turn. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "flying")),
			Cost:         effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Targets:      effects.TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := effects.BoostTargetUntilEOT("Starnheim Memento — +1/+1", 1, 1)(g, item); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}}.Apply(ctx)
			},
		}},
	})
}
