package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruin Ghost — Creature — Spirit {1}{W}, 1/1 (Worldwake, S82 OLD b488):
//
//	"{W}, {T}: Exile target land you control, then return it to the battlefield
//	 under your control."
//
// A plain blink of a land you control (Meneldor's Flicker): it comes back as a new object
// under your control (CR 400.7), untapped, with its enters triggers firing again. A target
// that stopped being a land you control, or left, does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "169705c3-32c1-4628-b108-c37ca5f27e24",
		Name:         "Ruin Ghost",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Exile target land you control, then return it to the battlefield under your control.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				land := effects.FirstLegalBattlefieldTarget(ctx)
				if land == uuid.Nil {
					return nil
				}
				return effects.Flicker{Target: land, Controller: item.Controller}.Apply(ctx)
			},
		}},
	})
}
