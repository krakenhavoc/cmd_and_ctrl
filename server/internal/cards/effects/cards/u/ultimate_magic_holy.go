package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ultimate Magic: Holy — Instant {2}{W}:
//
//	"Permanents you control gain indestructible until end of turn. If
//	 this spell was cast from exile, prevent all damage that would be
//	 dealt to you this turn.
//	 Foretell {2}{W}"
//
// ADR 0108 §7, Delivery PR 7 (#1904): Heroic Intervention's grant (the
// set fixed as the spell resolves, CR 611.2c), then, for a spell cast
// from exile (a foretold cast, or any other), the not-one-use shield
// protecting you for the rest of the turn. The zone is the one the
// spell was cast from, recorded on its stack item (CR 601.2a).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1eae48aa-f01d-4398-a68e-3b5af0542ab1",
		Name:         "Ultimate Magic: Holy",
		Completeness: effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{
			effects.Foretell("{2}{W}"),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GrantKeywordUntilEOT{
				Match:    effects.YouControl(),
				Keywords: []string{"indestructible"},
				Label:    "Ultimate Magic: Holy — indestructible",
			}).Apply(ctx); err != nil {
				return err
			}
			if item.CastFromZone != game.ZoneExile {
				return nil
			}
			return effects.PreventDamageFromSource{Protect: effects.ShieldYou}.Apply(ctx)
		},
	})
}
