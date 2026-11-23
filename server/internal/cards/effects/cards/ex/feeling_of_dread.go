package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feeling of Dread — Instant {1}{W} (Innistrad, #5572):
//
//	"Tap up to two target creatures.
//	 Flashback {1}{U} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Repel the Darkness's tap (tapEachLegalTarget over up to two creatures)
// without the draw, with the shared Flashback ({1}{U}).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "269208e1-353c-4334-941a-2177f6665fba",
		Name:             "Feeling of Dread",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{1}{U}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TapEachLegalTarget(ctx.Game, item)
		},
	})
}
