package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wyluli Wolf — Creature — Wolf {1}{G}, 1/1 (Arabian Nights, #5641):
//
//	"{T}: Target creature gets +1/+1 until end of turn."
//
// Any creature, the Wolf itself included. A {T} cost, so the Wolf can't
// use it the turn it arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00185f3b-2777-4417-a8e0-4691f41c0ec1",
		Name:         "Wyluli Wolf",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gets +1/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				target := effects.FirstLegalBattlefieldTarget(ctx)
				if target == uuid.Nil {
					return nil
				}
				return effects.BoostUntilEOT{Target: target, Power: 1, Toughness: 1, Label: "Wyluli Wolf — +1/+1 until end of turn"}.Apply(ctx)
			},
		}},
	})
}
