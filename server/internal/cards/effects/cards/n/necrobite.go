package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necrobite — Instant {2}{B} (Magic 2015, #5556):
//
//	"Target creature gains deathtouch until end of turn. Regenerate it."
//
// The keyword grant and the regeneration shield, both on the one target. The
// shield is the ordinary CR 701.19 one: the next time the creature would be
// destroyed this turn it taps, leaves combat and is healed instead.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2a3e219-08f3-4c05-b239-abf80a1c1c6b",
		Name:         "Necrobite",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, id := range effects.LegalTargetCardIDs(ctx) {
				if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: "Necrobite — deathtouch"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.Regenerate{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
