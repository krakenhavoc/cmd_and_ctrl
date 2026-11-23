package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Radiant's Judgment — Instant {2}{W} (Urza's Legacy, #5623):
//
//	"Destroy target creature with power 4 or greater.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Legion's Judgment at instant speed, with cycling (CR 702.29a). The power is
// read through the layers at announce and again as it resolves, so a
// creature shrunk below 4 in response is no longer a legal target and the
// spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85fd04e9-b39c-4ce0-97a1-58f9670a298f",
		Name:         "Radiant's Judgment",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
		OnResolve:    effects.DestroyTheTargetPermanent,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
