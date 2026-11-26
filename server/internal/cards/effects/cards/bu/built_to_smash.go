package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Built to Smash — Instant {R} (Kaladesh, #5539):
//
//	"Target attacking creature gets +3/+3 until end of turn. If it's an
//	 artifact creature, it gains trample until end of turn."
//
// Run Amok's pump on an attacking creature, with the trample only for an
// artifact creature, judged as the spell resolves through the layered view
// (CR 608.2b: a target no longer attacking is illegal and the spell does
// nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c363f338-0777-45e6-b13c-f15ba026b63a",
		Name:         "Built to Smash",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Built to Smash — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			if c, ok := ctx.Game.LookupCardForEffect(id); !ok || !c.IsArtifact() || !c.IsCreature() {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Built to Smash — trample"}.Apply(ctx)
		},
	})
}
