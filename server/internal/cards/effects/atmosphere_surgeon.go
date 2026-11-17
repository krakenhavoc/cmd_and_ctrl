package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Atmosphere Surgeon — Creature — Phyrexian Wizard {1}{U}, 2/1 (Phyrexia: All Will Be One):
//
//	"Whenever you cast a noncreature spell, put an oil counter on this creature.
//	 Remove an oil counter from this creature: Target creature gains flying until
//	 end of turn. Activate only as a sorcery."
//
// The counter is the whole cost, paid at announce; the target is checked again as
// the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "af22ed78-7054-4f5c-84ab-2a274c3d47fb",
		Name:         "Atmosphere Surgeon",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{s82OneB09OilOnNoncreatureCast("Atmosphere Surgeon")},
		Activated: []ActivatedAbility{{
			Label:        "Remove an oil counter from this creature: Target creature gains flying until end of turn. Activate only as a sorcery.",
			Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:         RemoveCountersFromThis(s82OneB09Oil, 1),
			Targets:      TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Atmosphere Surgeon — flying"}.Apply(ctx)
			},
		}},
	})
}
