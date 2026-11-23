package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shields Up! — Instant {1}{W} (Star Trek, #4973):
//
//	"Target artifact or creature you control gains hexproof and
//	 indestructible until end of turn. If it's a creature, put a +1/+1
//	 counter on it."
//
// "If it's a creature" is read as the spell resolves, after the layers,
// so an artifact animated by then gets the counter and a creature that
// stopped being one does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e14940b-87e1-4139-8f30-3629b3ad3a59",
		Name:         "Shields Up!",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target artifact or creature you control",
			effects.Or(effects.Artifact(), effects.Creature()), effects.YouControl()),
		Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "hexproof", "indestructible")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			id := ts[0].ID
			if err := (effects.GrantKeywordUntilEOT{
				Target:   id,
				Keywords: []string{"hexproof", "indestructible"},
				Label:    "Shields Up! — hexproof and indestructible",
			}).Apply(ctx); err != nil {
				return err
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || !c.IsCreature() {
				return nil
			}
			return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
		},
	})
}
