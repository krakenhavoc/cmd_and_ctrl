package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firebolt — Sorcery {R} (Odyssey, #5613):
//
//	"Firebolt deals 2 damage to any target.
//	 Flashback {4}{R} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Lava Dart's shape at sorcery speed with a mana flashback: the shared
// damageToFirstTarget body, and the Flashback constructor, which bundles
// the graveyard cast, the price and CR 702.34a's exile as the spell
// leaves the stack (resolved, countered, or with its target gone).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "7aa31280-12c3-479d-a6dd-f1c669043083",
		Name:             "Firebolt",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetAny(),
		Purpose:          effects.ForTargets(effects.DamageToTarget(0, 2)),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{4}{R}")},
		OnResolve:        effects.DamageToFirstTargetStep(2),
	})
}
