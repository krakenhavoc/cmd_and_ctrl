package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swat — Instant {1}{B}{B} (Urza's Legacy, #5623):
//
//	"Destroy target creature with power 2 or less.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Destroy a small creature, with cycling (CR 702.29a). The power is read
// through the layers at announce and again as it resolves, so a creature
// pumped above 2 in response is no longer a legal target and the spell does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7bb5c5af-6943-4958-a666-0c322d5b9f33",
		Name:         "Swat",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with power 2 or less", effects.PowerLE(2)),
		OnResolve:    effects.DestroyTheTargetPermanent,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
