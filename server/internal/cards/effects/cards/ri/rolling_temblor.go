package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rolling Temblor — Sorcery {2}{R} (Innistrad, #5572):
//
//	"Rolling Temblor deals 2 damage to each creature without flying.
//	 Flashback {4}{R}{R} (You may cast this card from your graveyard
//	 for its flashback cost. Then exile it.)"
//
// Seismic Rupture's body (2 damage to each creature without flying)
// with the shared Flashback ({4}{R}{R}).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e450014f-7152-476b-ae6e-0d49fcb91d18",
		Name:         "Rolling Temblor",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: a creature with flying is spared.
		Purpose:          game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{4}{R}{R}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.WithoutKeyword("flying")), 2)
		},
	})
}
