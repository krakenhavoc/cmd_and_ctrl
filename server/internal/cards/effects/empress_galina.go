package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Empress Galina — Legendary Creature — Merfolk Noble {3}{U}{U}, 3/3:
//
//	"{U}{U}, {T}: Gain control of target legendary permanent. (This effect
//	 lasts indefinitely.)"
//
// Agent of Treachery's indefinite GainControl behind a tap ability, with
// Minamo School at Water's Edge's legendary-permanent target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ea2614bc-00ad-4859-98f5-689f33357fe6",
		Name:         "Empress Galina",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}{U}, {T}: Gain control of target legendary permanent.",
			Cost:    Plus(ManaCost("{U}{U}"), TapCost()),
			Targets: TargetPermanent("target legendary permanent", b05Legendary()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return GainControl{
					Target:     id,
					Controller: item.Controller,
					Duration:   game.IndefiniteDuration(),
					Label:      "Empress Galina — gain control (indefinitely)",
				}.Apply(ctx)
			},
		}},
	})
}
