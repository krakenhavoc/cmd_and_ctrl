package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ac08fae8-208c-4602-8d39-9bfd29b53a5e",
		Name:         "Tower of the Magistrate",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Target creature gains protection from artifacts until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0)),
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"protection from artifacts"},
					Label: "Tower of the Magistrate — protection from artifacts"}.Apply(ctx)
			},
		}},
	})
}
