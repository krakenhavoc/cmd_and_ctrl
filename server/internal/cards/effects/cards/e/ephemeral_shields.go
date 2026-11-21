package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ephemeral Shields — Instant {1}{W}:
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Target creature gains indestructible until end of turn."
//
// Appeal to Eirdu's convoke (CR 702.51) over Divine Resilience's grant: the
// keyword lands on the target if it is still legal as the spell resolves
// (CR 608.2b). A creature tapped to pay for convoke can still be the target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1a89535-7126-48e0-b335-b732a677b157",
		Name:         "Ephemeral Shields",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"indestructible"},
					Label:    "Ephemeral Shields — indestructible until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
