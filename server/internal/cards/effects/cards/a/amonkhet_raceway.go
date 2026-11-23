package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amonkhet Raceway — Land:
//
//	"Start your engines!
//	 {T}: Add {C}.
//	 Max speed — {T}: Target creature gains haste until end of turn."
//
// ADR 0138 (#2122). The haste ability is an ordinary activated ability
// that can be activated only while its controller has max speed
// (CR 702.178a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fe174586-d36b-40f6-babd-1f98e76eec22",
		Name:            "Amonkhet Raceway",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		ManaAbilities: []effects.ManaAbility{
			{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "{T}: Add {C}."},
		},
		Activated: []effects.ActivatedAbility{
			effects.MaxSpeedActivated(effects.ActivatedAbility{
				Label:   "Max speed — {T}: Target creature gains haste until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"}, Label: "Amonkhet Raceway — haste"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			}),
		},
	})
}
