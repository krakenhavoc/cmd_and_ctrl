package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flare of Faith — Instant {1}{W} (Innistrad: Midnight Hunt, #5497):
//
//	"Target creature gets +2/+2 until end of turn. If it's a Human, instead
//	 it gets +3/+3 and gains indestructible until end of turn."
//
// "Instead" is one outcome or the other, decided as the spell resolves from
// the creature's current subtypes (a creature that is a Human through an
// effect counts, CR 205.3m).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13ebe447-1f8f-4b12-9e35-9d334a863c19",
		Name:         "Flare of Faith",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			c, ok := effects.BattlefieldCardForEffect(ctx.Game, t.ID)
			if !ok {
				return nil
			}
			if c.HasSubtype(ctx.TextSubtype("Human")) {
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 3, Label: "Flare of Faith — +3/+3"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"}, Label: "Flare of Faith — indestructible"}.Apply(ctx)
			}
			return effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Flare of Faith — +2/+2"}.Apply(ctx)
		},
	})
}
