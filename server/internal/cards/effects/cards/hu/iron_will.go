package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iron Will — Instant {W} (Urza's Legacy, #5623):
//
//	"Target creature gets +0/+4 until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Lightshield Parry's shape at +0/+4: a toughness boost that ends in the
// cleanup step (CR 514.2), with cycling (CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc0decb9-32da-47aa-a51a-0d9c623c534a",
		Name:         "Iron Will",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 4)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(0, 4, "Iron Will — +0/+4"),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
