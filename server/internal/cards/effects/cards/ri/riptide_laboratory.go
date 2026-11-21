package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riptide Laboratory — Land (EDHREC rank 1771):
//
//	"{T}: Add {C}.
//	 {1}{U}, {T}: Return target Wizard you control to its owner's
//	 hand."
//
// The Wizard deck's ETB re-buy. A painless {C} and a mana-plus-tap
// activated ability with a "Wizard you control" target clause —
// effective subtypes, so a changeling counts — whose effect is the
// bounce primitive. A Wizard that left in response is skipped by the
// resolution re-check.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "444d50dd-a44a-42db-bbf6-d0978e3bd6a3",
		Name:          "Riptide Laboratory",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}: Return target Wizard you control to its owner's hand.",
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target Wizard you control", effects.YouControl(), effects.HasSubtype("Wizard")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BounceToHand{Target: id}.Apply(ctx)
			},
		}},
	})
}
