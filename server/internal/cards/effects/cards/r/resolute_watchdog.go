package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resolute Watchdog — Creature — Dog {W}, 1/3 (Ravnica Allegiance):
//
//	"Defender
//	 {1}, Sacrifice this creature: Target creature you control gains
//	 indestructible until end of turn."
//
// The target is chosen before the sacrifice is paid, so the Watchdog may name
// itself; it is gone by resolution and the ability then does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ca2aee09-398d-4eaa-acbb-98663216185b",
		Name:            "Resolute Watchdog",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Target creature you control gains indestructible until end of turn.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				targets := ctx.LegalTargets()
				if len(targets) == 0 || targets[0].Kind != game.TargetCard {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: targets[0].ID, Keywords: []string{"indestructible"},
					Label: "Resolute Watchdog — indestructible until end of turn"}.Apply(ctx)
			},
		}},
	})
}
