package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aquus Steed — Creature — Beast {3}{U}, 1/3 (Return to Ravnica, S82 OLD):
//
//	"{2}{U}, {T}: Target creature gets -2/-0 until end of turn."
//
// A shrink on any creature, paid with mana and the Steed's own tap (CR
// 602.5a: a summoning-sick Steed can't). A target that has gone is not
// shrunk (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3517abe3-eb18-49af-9f1c-b372d2039d99",
		Name:         "Aquus Steed",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}: Target creature gets -2/-0 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -2, 0)),
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				target := effects.FirstLegalBattlefieldTarget(ctx)
				if target == uuid.Nil {
					return nil
				}
				return effects.BoostUntilEOT{Target: target, Power: -2, Label: "Aquus Steed — -2/-0 until end of turn"}.Apply(ctx)
			},
		}},
	})
}
