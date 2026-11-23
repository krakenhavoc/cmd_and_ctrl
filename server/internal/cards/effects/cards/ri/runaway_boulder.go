package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Runaway Boulder — Artifact {6} (The Lost Caverns of Ixalan, #4470):
//
//	"Flash
//	 When this artifact enters, it deals 6 damage to target creature an
//	 opponent controls.
//	 Cycling {2}"
//
// Flash is the printed keyword (CR 702.8a). The enters trigger targets as
// it goes on the stack (CR 603.3d), only a creature an opponent controls,
// and the damage is dealt by the Boulder, as it last existed if it has
// left (CR 608.2h). With no such creature the trigger is removed (CR
// 603.3d). Cycling is the shared CR 702.29a ability from the hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "57f15861-82cf-4faf-827c-9bc7f4844d2a",
		Name:            "Runaway Boulder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Runaway Boulder — 6 damage to target creature an opponent controls",
				effects.SourceDealsDamageToEachLegalTarget(6)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
