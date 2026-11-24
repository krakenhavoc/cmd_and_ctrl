package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Road // Ruin — split card with aftermath (Modern Horizons 2, #5501):
//
//	Road — Instant {2}{G}: "Search your library for a basic land card, put it
//	       onto the battlefield tapped, then shuffle."
//	Ruin — Sorcery {1}{R}{R}: "Aftermath (Cast this spell only from your
//	       graveyard. Then exile it.) Ruin deals damage to target creature
//	       equal to the number of lands you control."
//
// Each half is an ADR 0034 face (ADR 0103): Road under the bare oracle ID,
// Ruin under "#1". Road casts from the hand; the engine opens Ruin only from
// the graveyard and exiles it as it leaves the stack (CR 702.127a), as it
// does Indulge // Excess. Ruin reads the lands its caster controls as it
// resolves, and Road's search is Cultivate's (a basic land onto the
// battlefield tapped, then a shuffle).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f18d8939-7ee8-4a44-a1c8-6ae3a08f52fd",
		Name:         "Road",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		OnResolve:    effects.OldB032RoadSearch,
	})
	effects.Register(effects.Spec{
		OracleID:     "f18d8939-7ee8-4a44-a1c8-6ae3a08f52fd#1",
		Name:         "Ruin",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsLand() })
			return effects.DealDamageToTheTarget(n)(ctx.Game, item)
		},
	})
}
