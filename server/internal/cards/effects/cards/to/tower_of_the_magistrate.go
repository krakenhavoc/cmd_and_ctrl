package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tower of the Magistrate — Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}: Target creature gains protection from artifacts until end of
//	 turn."
//
// The grant is a layer 6 keyword until end of turn on the target as it
// resolves, with the protection read as the engine reads Angelic Curator's
// printed one (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac08fae8-208c-4602-8d39-9bfd29b53a5e",
		Name:         "Tower of the Magistrate",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Target creature gains protection from artifacts until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"protection from artifacts"},
					Label: "Tower of the Magistrate — protection from artifacts"}.Apply(ctx)
			},
		}},
	})
}
