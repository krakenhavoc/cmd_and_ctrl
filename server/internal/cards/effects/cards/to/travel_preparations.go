package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Travel Preparations — Sorcery {1}{G} (Innistrad, #5572):
//
//	"Put a +1/+1 counter on each of up to two target creatures.
//	 Flashback {1}{W} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Gird for Battle's body (a +1/+1 counter on each of up to two creatures)
// with the shared Flashback ({1}{W}).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "8432729e-5fd4-4661-8b79-e41b1d03eb7b",
		Name:             "Travel Preparations",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{1}{W}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PutPlusOneCounterOnEachLegalTarget(ctx.Game, item)
		},
	})
}
