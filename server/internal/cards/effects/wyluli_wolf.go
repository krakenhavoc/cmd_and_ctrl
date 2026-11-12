package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "00185f3b-2777-4417-a8e0-4691f41c0ec1",
		Name:         "Wyluli Wolf",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature gets +1/+1 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 1, 1)),
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target := FirstLegalBattlefieldTarget(ctx)
				if target == uuid.Nil {
					return nil
				}
				return BoostUntilEOT{Target: target, Power: 1, Toughness: 1, Label: "Wyluli Wolf — +1/+1 until end of turn"}.Apply(ctx)
			},
		}},
	})
}
