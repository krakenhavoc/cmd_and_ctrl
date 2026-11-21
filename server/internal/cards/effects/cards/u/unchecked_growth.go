package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unchecked Growth — Instant — Arcane {2}{G} (Betrayers of Kamigawa,
// #5602):
//
//	"Target creature gets +4/+4 until end of turn. If it's a Spirit, it
//	 gains trample until end of turn."
//
// Whether it is a Spirit is read as the spell resolves (CR 608.2c), by
// effective subtype, so a changeling gains trample and "Spirit" reads
// through a text change.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea4052bc-cd37-467c-9fba-3f1ed537b781",
		Name:         "Unchecked Growth",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 4, Toughness: 4, Label: "Unchecked Growth — +4/+4"}).Apply(ctx); err != nil {
				return err
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || !c.HasSubtypeWord(ctx.TextSubtype("Spirit")) {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Unchecked Growth — trample"}.Apply(ctx)
		},
	})
}
