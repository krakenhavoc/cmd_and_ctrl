package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gallant Strike — Instant {1}{W} (Aetherdrift, #4464):
//
//	"Destroy target creature with toughness 4 or greater.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Destroy Evil's first bullet as a whole spell, with cycling
// (CR 702.29a). The toughness is read through the layers at announce and
// again as it resolves, so a creature shrunk below 4 in response is no
// longer a legal target and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04dcda3b-3401-4523-8b97-b700a14feb36",
		Name:         "Gallant Strike",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with toughness 4 or greater", effects.ToughnessGE(4)),
		OnResolve:    effects.DestroyTheTargetPermanent,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
