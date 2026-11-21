package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Might of the Masses — Instant {G}:
//
//	"Target creature gets +1/+1 until end of turn for each creature you
//	 control."
//
// The creatures are counted as the spell resolves, so the target itself
// counts when you control it and a creature that died in response does
// not. A target that is illegal means the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f73416a8-40e5-4791-948f-147ef2221ee7",
		Name:         "Might of the Masses",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetCard {
				return nil
			}
			n := len(effects.PermanentsControlledByMatching(ctx.Game, item.Controller, effects.Creature()))
			return effects.BoostUntilEOT{Target: legal[0].ID, Power: n, Toughness: n, Label: "Might of the Masses — +1/+1 per creature"}.Apply(ctx)
		},
	})
}
