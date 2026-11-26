package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirran Spy — Creature — Drone {2}{U}, 1/3 (Mirrodin Besieged, S82 OLD b468,
// #5576):
//
//	"Flying
//	 Whenever you cast an artifact spell, you may untap target creature."
//
// A cast trigger (artifact spells only, yours only) that targets a creature
// as it goes on the stack, anyone's, and asks "you may" as it resolves (CR
// 603.5). The target is read again then; one that has gone is not untapped
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8e5d72a1-6ae0-445d-bac7-62da3bd3a8d6",
		Name:            "Mirran Spy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Optional(effects.WheneverYouCast(effects.Artifact(), "Mirran Spy — you may untap target creature",
				effects.UntapFirstLegalTarget), "Mirran Spy — untap target creature?"),
				effects.TargetCreature("target creature")),
		},
	})
}
