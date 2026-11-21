package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Without Weakness — Instant {1}{B} (Hour of Devastation, #5534):
//
//	"Target creature you control gains indestructible until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Divine Resilience's grant without its kicker: a target that left in
// response does nothing (CR 608.2b). The shared Cycling row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fffa9334-3576-4f70-9605-2ad062cdbc69",
		Name:         "Without Weakness",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.GrantKeywordUntilEOT{
				Target:   id,
				Keywords: []string{"indestructible"},
				Label:    "Without Weakness — indestructible until end of turn",
			}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
