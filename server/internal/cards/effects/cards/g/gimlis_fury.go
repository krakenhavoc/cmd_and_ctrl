package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gimli's Fury — Instant {1}{R} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Target creature gets +3/+2 until end of turn. If it's legendary, it also
//	 gains trample until end of turn."
//
// "If it's legendary" is read as the spell resolves, through the layers: the
// trample is part of the same resolution (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b550c27-14b0-45fb-b17e-fcb061dc6ad0",
		Name:         "Gimli's Fury",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetCard {
				return nil
			}
			id := legal[0].ID
			if err := (effects.BoostUntilEOT{Target: id, Power: 3, Toughness: 2, Label: "Gimli's Fury — +3/+2"}).Apply(ctx); err != nil {
				return err
			}
			if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsLegendary() {
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Gimli's Fury — trample"}.Apply(ctx)
			}
			return nil
		},
	})
}
