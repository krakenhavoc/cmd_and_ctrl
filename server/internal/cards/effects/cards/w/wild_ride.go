package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wild Ride — Sorcery {R} (#3931):
//
//	"Target creature gets +3/+0 and gains haste until end of turn.
//	 Harmonize {4}{R}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "61978202-538b-4f92-acfc-e95f4a6ce466",
		Name:             "Wild Ride",
		Purpose:          effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0, "haste")),
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Harmonize("{4}{R}")},
		Targets:          effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.FirstTargetGainsUntilEndOfTurn(3, "Wild Ride — +3/+0 and haste", "haste")(ctx.Game, item)
		},
	})
}
