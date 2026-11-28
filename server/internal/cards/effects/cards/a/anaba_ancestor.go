package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Anaba Ancestor — Creature — Minotaur Spirit {1}{R}, 1/1 (Homelands, S82 OLD
// b823, #5635):
//
//	"{T}: Another target Minotaur creature gets +1/+1 until end of turn."
//
// A pump on another Minotaur (Another excludes the Ancestor itself, CR 115.7 via
// ExcludeSource), until the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d2e7422-397a-47d4-9a1e-1957c7db9a18",
		Name:         "Anaba Ancestor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Another target Minotaur creature gets +1/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.Another(effects.TargetCreature("another target Minotaur creature", effects.HasSubtype("Minotaur"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Anaba Ancestor — +1/+1"}.Apply(ctx)
			},
		}},
	})
}
